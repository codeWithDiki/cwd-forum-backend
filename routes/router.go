package routes

import (
	"gin-quickstart/config"
	"gin-quickstart/internal/handler"
	"gin-quickstart/internal/middleware"
	"gin-quickstart/internal/repository"
	"gin-quickstart/internal/service"
	"gin-quickstart/pkg/logger"

	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	wp := *workerpool.New(20)

	redis := config.RedisClient

	log, err := logger.New(logger.Config{
		Production: true,
	})

	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	r.Use(middleware.LoggerMiddleware(log))
	r.Use(middleware.FileUploadMiddleware(&wp))

	{
		v1 := r.Group("/v1")
		db, err := config.InitDB()
		// Initialize your GORM DB connection here

		if err != nil {
			panic("failed to connect database: " + err.Error())
		}

		{
			userRepo := repository.NewUserRepository(db, redis)
			userService := service.NewUserService(userRepo)
			userHandler := handler.NewUserHandler(userService)

			user := v1.Group("/users")

			user.GET("/", middleware.JWTMiddleware(redis), userHandler.GetAllUsers)
			user.POST("/", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), userHandler.CreateUser)
			user.GET("/:id", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), userHandler.GetUserByID)
			user.GET("/username/:username", middleware.JWTMiddleware(redis), userHandler.GetUserByUsername)
			user.PATCH("/:id", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), userHandler.UpdateUser)

			userUtility := user.Group("/utility")
			userUtility.GET("/me", middleware.JWTMiddleware(redis), userHandler.GetUserByID)
			userUtility.PATCH("/me", middleware.JWTMiddleware(redis), userHandler.UpdateUser)
			userUtility.POST("/follow/:id", middleware.JWTMiddleware(redis), userHandler.Follow)
			userUtility.POST("/unfollow/:id", middleware.JWTMiddleware(redis), userHandler.Unfollow)

			user.DELETE("/:id", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), userHandler.DeleteUser)
		}

		{
			categoryRepo := repository.NewCategoryRepository(db, redis)
			categoryService := service.NewCategoryService(categoryRepo)
			categoryHandler := handler.NewCategoryHandler(categoryService)

			category := v1.Group("/categories")

			category.GET("/", categoryHandler.GetAllCategories)
			category.POST("/", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), categoryHandler.Create)
			category.GET("/:id", categoryHandler.GetCategoryByID)
			category.GET("/slug/:slug", categoryHandler.GetCategoryBySlug)
			category.PATCH("/:id", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), categoryHandler.Update)
			category.DELETE("/:id", middleware.JWTMiddleware(redis), middleware.IsAdminLogged(redis), categoryHandler.Delete)

		}

		{
			threadRepo := repository.NewThreadRepository(db, redis)
			threadService := service.NewThreadService(threadRepo)
			threadHandler := handler.NewThreadHandler(threadService)

			thread := v1.Group("/threads")

			thread.GET("/", threadHandler.GetAllThreads)
			thread.POST("/", middleware.JWTMiddleware(redis), middleware.IsUserBanned(db), middleware.S3Middleware(), threadHandler.Create)
			thread.GET("/:id", threadHandler.GetThreadByID)
			thread.GET("/slug/:slug", threadHandler.GetThreadBySlug)
			thread.GET("/category/:category_id", threadHandler.GetThreadsByCategoryID)
			thread.GET("/author/:author_id", threadHandler.GetThreadsByAuthorID)
			thread.GET("/tag/:tag_id", threadHandler.GetThreadsByTagID)
			thread.PATCH("/:id", middleware.JWTMiddleware(redis), middleware.IsCanUpdateThread(db, threadService), threadHandler.Update)
			thread.DELETE("/:id", middleware.JWTMiddleware(redis), middleware.IsCanUpdateThread(db, threadService), threadHandler.Delete)
		}

		{
			postRepo := repository.NewPostRepository(db, redis)
			postService := service.NewPostService(postRepo)
			postHandler := handler.NewPostHandler(postService)

			post := v1.Group("/posts")
			post.GET("/", postHandler.GetAllPosts)
			post.GET("/:id", postHandler.GetPostByID)
			post.GET("/thread/:thread_id", postHandler.GetPostsByThreadID)
			post.GET("/author/:author_id", postHandler.GetPostsByAuthorID)
			post.POST("/", middleware.JWTMiddleware(redis), middleware.IsUserBanned(db), middleware.S3Middleware(), postHandler.Create)
			post.POST("/:id/votes", middleware.JWTMiddleware(redis), middleware.IsUserBanned(db), postHandler.VotePost)
			post.GET("/:id/votes", postHandler.GetPostVotes)
			post.POST("/:id/reactions", middleware.JWTMiddleware(redis), middleware.IsUserBanned(db), postHandler.ReactPost)
			post.PATCH("/:id", postHandler.Update)
			post.DELETE("/:id", postHandler.Delete)
			post.POST("/:id/mark-as-solution", middleware.JWTMiddleware(redis), postHandler.MarkAsSolution)
		}

		{
			tagRepo := repository.NewTagRepository(db, redis)
			tagService := service.NewTagService(tagRepo)
			tagHandler := handler.NewTagHandler(tagService)

			tag := v1.Group("/tags")

			tag.Use(middleware.JWTMiddleware(redis))

			tag.GET("/", tagHandler.GetAllTags)
			tag.POST("/", middleware.IsAdminLogged(redis), tagHandler.CreateTag)
			tag.GET("/:id", tagHandler.GetTagByID)
			tag.GET("/slug/:slug", tagHandler.GetTagBySlug)
			tag.PATCH("/:id", middleware.IsAdminLogged(redis), tagHandler.UpdateTag)
			tag.DELETE("/:id", middleware.IsAdminLogged(redis), tagHandler.DeleteTag)
		}

		{
			attachmentRepo := repository.NewAttachmentRepository(db, redis)
			attachmentService := service.NewAttachmentService(attachmentRepo)
			attachmentHandler := handler.NewAttachmentHandler(attachmentService)

			attachment := v1.Group("/attachments")

			attachment.Use(middleware.JWTMiddleware(redis))
			attachment.Use(middleware.IsAdminLogged(redis))

			attachment.GET("/", attachmentHandler.GetAllAttachments)
			attachment.GET("/:id", attachmentHandler.GetAttachmentByID)
			attachment.DELETE("/:id", attachmentHandler.DeleteAttachment)
			attachment.GET("/post/:post_id", attachmentHandler.GetAttachmentsByPostID)
		}

		{
			badgeRepo := repository.NewBadgeRepository(db, redis)
			badgeService := service.NewBadgeService(badgeRepo)
			badgeHandler := handler.NewBadgeHandler(badgeService)

			badge := v1.Group("/badges")

			badge.Use(middleware.JWTMiddleware(redis))

			badge.GET("/", badgeHandler.GetAllBadges)
			badge.POST("/", middleware.IsAdminLogged(redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(&wp), badgeHandler.Create)
			badge.GET("/:id", badgeHandler.GetBadgeByID)
			badge.PATCH("/:id", middleware.IsAdminLogged(redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(&wp), badgeHandler.Update)
			badge.DELETE("/:id", middleware.IsAdminLogged(redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(&wp), badgeHandler.Delete)
		}

		{
			authRepo := repository.NewAuthRepository(db, redis)
			authService := service.NewAuthService(authRepo)
			authHandler := handler.NewAuthHandler(authService)

			auth := v1.Group("/auth")

			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.JWTMiddleware(redis), authHandler.Logout)
			auth.PATCH("/profile", middleware.JWTMiddleware(redis), middleware.S3Middleware(), middleware.FileUploadMiddleware(&wp), authHandler.UpdateProfile)
		}

	}

	return r
}
