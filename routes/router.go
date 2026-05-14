package routes

import (
	"gin-quickstart/internal/app"
	"gin-quickstart/internal/handler"
	"gin-quickstart/internal/middleware"
	"gin-quickstart/internal/repository"
	"gin-quickstart/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(deps app.Dependencies) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.LoggerMiddleware(deps.Logger))
	r.Use(middleware.FileUploadMiddleware(deps.Worker.Worker))

	{
		v1 := r.Group("/v1")

		//repostiory
		userRepo := repository.NewUserRepository(deps.Logger, deps.DB, deps.Redis)
		userService := service.NewUserService(deps.Logger, userRepo)
		userHandler := handler.NewUserHandler(deps.Logger, userService)

		user := v1.Group("/users")

		user.GET("/", middleware.JWTMiddleware(deps.Redis), userHandler.GetAllUsers)
		user.POST("/", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), userHandler.CreateUser)
		user.GET("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), userHandler.GetUserByID)
		user.GET("/username/:username", middleware.JWTMiddleware(deps.Redis), userHandler.GetUserByUsername)
		user.PATCH("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), userHandler.UpdateUser)

		userUtility := user.Group("/utility")
		userUtility.GET("/me", middleware.JWTMiddleware(deps.Redis), userHandler.GetUserByID)
		userUtility.PATCH("/me", middleware.JWTMiddleware(deps.Redis), userHandler.UpdateUser)
		userUtility.POST("/follow/:id", middleware.JWTMiddleware(deps.Redis), userHandler.Follow)
		userUtility.POST("/unfollow/:id", middleware.JWTMiddleware(deps.Redis), userHandler.Unfollow)

		user.DELETE("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), userHandler.DeleteUser)

		categoryRepo := repository.NewCategoryRepository(deps.DB, deps.Redis)
		categoryService := service.NewCategoryService(categoryRepo)
		categoryHandler := handler.NewCategoryHandler(categoryService)

		category := v1.Group("/categories")

		category.GET("/", categoryHandler.GetAllCategories)
		category.POST("/", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), categoryHandler.Create)
		category.GET("/:id", categoryHandler.GetCategoryByID)
		category.GET("/slug/:slug", categoryHandler.GetCategoryBySlug)
		category.PATCH("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), categoryHandler.Update)
		category.DELETE("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsAdminLogged(*userRepo, deps.Redis), categoryHandler.Delete)

		threadRepo := repository.NewThreadRepository(deps.DB, deps.Redis)
		threadService := service.NewThreadService(threadRepo)
		threadHandler := handler.NewThreadHandler(threadService)

		thread := v1.Group("/threads")

		thread.GET("/", threadHandler.GetAllThreads)
		thread.POST("/", middleware.JWTMiddleware(deps.Redis), middleware.IsUserBanned(deps.DB), middleware.S3Middleware(), threadHandler.Create)
		thread.GET("/:id", threadHandler.GetThreadByID)
		thread.GET("/slug/:slug", threadHandler.GetThreadBySlug)
		thread.GET("/category/:category_id", threadHandler.GetThreadsByCategoryID)
		thread.GET("/author/:author_id", threadHandler.GetThreadsByAuthorID)
		thread.GET("/tag/:tag_id", threadHandler.GetThreadsByTagID)
		thread.PATCH("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsCanUpdateThread(deps.DB, threadService), threadHandler.Update)
		thread.DELETE("/:id", middleware.JWTMiddleware(deps.Redis), middleware.IsCanUpdateThread(deps.DB, threadService), threadHandler.Delete)

		postRepo := repository.NewPostRepository(deps.DB, deps.Redis)
		postService := service.NewPostService(postRepo)
		postHandler := handler.NewPostHandler(postService)

		post := v1.Group("/posts")
		post.GET("/", postHandler.GetAllPosts)
		post.GET("/:id", postHandler.GetPostByID)
		post.GET("/thread/:thread_id", postHandler.GetPostsByThreadID)
		post.GET("/author/:author_id", postHandler.GetPostsByAuthorID)
		post.POST("/", middleware.JWTMiddleware(deps.Redis), middleware.IsUserBanned(deps.DB), middleware.S3Middleware(), postHandler.Create)
		post.POST("/:id/votes", middleware.JWTMiddleware(deps.Redis), middleware.IsUserBanned(deps.DB), postHandler.VotePost)
		post.GET("/:id/votes", postHandler.GetPostVotes)
		post.POST("/:id/reactions", middleware.JWTMiddleware(deps.Redis), middleware.IsUserBanned(deps.DB), postHandler.ReactPost)
		post.PATCH("/:id", postHandler.Update)
		post.DELETE("/:id", postHandler.Delete)
		post.POST("/:id/mark-as-solution", middleware.JWTMiddleware(deps.Redis), postHandler.MarkAsSolution)

		tagRepo := repository.NewTagRepository(deps.DB, deps.Redis)
		tagService := service.NewTagService(tagRepo)
		tagHandler := handler.NewTagHandler(tagService)

		tag := v1.Group("/tags")

		tag.Use(middleware.JWTMiddleware(deps.Redis))

		tag.GET("/", tagHandler.GetAllTags)
		tag.POST("/", middleware.IsAdminLogged(*userRepo, deps.Redis), tagHandler.CreateTag)
		tag.GET("/:id", tagHandler.GetTagByID)
		tag.GET("/slug/:slug", tagHandler.GetTagBySlug)
		tag.PATCH("/:id", middleware.IsAdminLogged(*userRepo, deps.Redis), tagHandler.UpdateTag)
		tag.DELETE("/:id", middleware.IsAdminLogged(*userRepo, deps.Redis), tagHandler.DeleteTag)

		attachmentRepo := repository.NewAttachmentRepository(deps.DB, deps.Redis)
		attachmentService := service.NewAttachmentService(attachmentRepo)
		attachmentHandler := handler.NewAttachmentHandler(attachmentService)

		attachment := v1.Group("/attachments")

		attachment.Use(middleware.JWTMiddleware(deps.Redis))
		attachment.Use(middleware.IsAdminLogged(*userRepo, deps.Redis))

		attachment.GET("/", attachmentHandler.GetAllAttachments)
		attachment.GET("/:id", attachmentHandler.GetAttachmentByID)
		attachment.DELETE("/:id", attachmentHandler.DeleteAttachment)
		attachment.GET("/post/:post_id", attachmentHandler.GetAttachmentsByPostID)

		badgeRepo := repository.NewBadgeRepository(deps.DB, deps.Redis)
		badgeService := service.NewBadgeService(badgeRepo)
		badgeHandler := handler.NewBadgeHandler(badgeService)

		badge := v1.Group("/badges")

		badge.Use(middleware.JWTMiddleware(deps.Redis))

		badge.GET("/", badgeHandler.GetAllBadges)
		badge.POST("/", middleware.IsAdminLogged(*userRepo, deps.Redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(deps.Worker.Worker), badgeHandler.Create)
		badge.GET("/:id", badgeHandler.GetBadgeByID)
		badge.PATCH("/:id", middleware.IsAdminLogged(*userRepo, deps.Redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(deps.Worker.Worker), badgeHandler.Update)
		badge.DELETE("/:id", middleware.IsAdminLogged(*userRepo, deps.Redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(deps.Worker.Worker), badgeHandler.Delete)

		notificationRepo := repository.NewNotificationRepository(deps.DB, deps.Redis)
		notificationService := service.NewNotificationService(notificationRepo)
		notificationHandler := handler.NewNotificationHandler(notificationService)

		notification := v1.Group("/notifications")
		notification.Use(middleware.JWTMiddleware(deps.Redis))

		notification.GET("/", notificationHandler.GetNotifications)
		notification.GET("/:id", notificationHandler.GetNotificationByID)
		notification.POST("/", notificationHandler.CreateNotification)
		notification.PATCH("/:id/read", notificationHandler.MarkAsRead)
		notification.PATCH("/:id", notificationHandler.UpdateNotification)
		notification.DELETE("/:id", notificationHandler.DeleteNotification)

		authRepo := repository.NewAuthRepository(deps.DB, deps.Redis)
		authService := service.NewAuthService(authRepo)
		authHandler := handler.NewAuthHandler(authService)

		auth := v1.Group("/auth")

		auth.GET("/profile", middleware.JWTMiddleware(deps.Redis), authHandler.GetProfile)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", middleware.JWTMiddleware(deps.Redis), authHandler.Logout)
		auth.PATCH("/profile", middleware.JWTMiddleware(deps.Redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(deps.Worker.Worker), authHandler.UpdateProfile)

	}

	return r
}
