package routes

import (
	"gin-quickstart/config"
	"gin-quickstart/internal/handler"
	"gin-quickstart/internal/middleware"
	"gin-quickstart/internal/repository"
	"gin-quickstart/internal/service"
	"net/http"

	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	wp := *workerpool.New(20)

	r.Use(middleware.AttachmentMiddleware(&wp))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	{
		v1 := r.Group("/v1")
		db, err := config.InitDB()
		// Initialize your GORM DB connection here

		if err != nil {
			panic("failed to connect database: " + err.Error())
		}

		{
			userRepo := repository.NewUserRepository(db)
			userService := service.NewUserService(userRepo)
			userHandler := handler.NewUserHandler(userService)

			user := v1.Group("/users")

			user.GET("/", middleware.JWTMiddleware(), middleware.IsAdminLogged(), userHandler.GetAllUsers)
			user.POST("/", userHandler.CreateUser)
			user.POST("/login", userHandler.Login)
			user.GET("/:id", middleware.JWTMiddleware(), middleware.IsAdminLogged(), userHandler.GetUserByID)
			user.GET("/username/:username", userHandler.GetUserByUsername)
			user.PATCH("/:id", middleware.JWTMiddleware(), middleware.IsAdminLogged(), userHandler.UpdateUser)

			userUtility := user.Group("/utility")
			userUtility.GET("/me", middleware.JWTMiddleware(), userHandler.GetUserByID)
			userUtility.PATCH("/me", middleware.JWTMiddleware(), userHandler.UpdateUser)

			user.DELETE("/:id", middleware.JWTMiddleware(), middleware.IsAdminLogged(), userHandler.DeleteUser)
		}

		{
			categoryRepo := repository.NewCategoryRepository(db)
			categoryService := service.NewCategoryService(categoryRepo)
			categoryHandler := handler.NewCategoryHandler(categoryService)

			category := v1.Group("/categories")

			category.GET("/", categoryHandler.GetAllCategories)
			category.POST("/", middleware.JWTMiddleware(), middleware.IsAdminLogged(), categoryHandler.Create)
			category.GET("/:id", categoryHandler.GetCategoryByID)
			category.GET("/slug/:slug", categoryHandler.GetCategoryBySlug)
			category.PATCH("/:id", middleware.JWTMiddleware(), middleware.IsAdminLogged(), categoryHandler.Update)
			category.DELETE("/:id", middleware.JWTMiddleware(), middleware.IsAdminLogged(), categoryHandler.Delete)

		}

		{
			threadRepo := repository.NewThreadRepository(db)
			threadService := service.NewThreadService(threadRepo)
			threadHandler := handler.NewThreadHandler(threadService)

			thread := v1.Group("/threads")

			thread.GET("/", threadHandler.GetAllThreads)
			thread.POST("/", middleware.JWTMiddleware(), middleware.IsUserBanned(db), middleware.S3Middleware(), threadHandler.Create)
			thread.GET("/:id", threadHandler.GetThreadByID)
			thread.GET("/slug/:slug", threadHandler.GetThreadBySlug)
			thread.GET("/category/:category_id", threadHandler.GetThreadsByCategoryID)
			thread.GET("/author/:author_id", threadHandler.GetThreadsByAuthorID)
			thread.GET("/tag/:tag_id", threadHandler.GetThreadsByTagID)
			thread.PATCH("/:id", middleware.JWTMiddleware(), middleware.IsCanUpdateThread(db, threadService), threadHandler.Update)
			thread.DELETE("/:id", middleware.JWTMiddleware(), middleware.IsCanUpdateThread(db, threadService), threadHandler.Delete)
		}

		{
			postRepo := repository.NewPostRepository(db)
			postService := service.NewPostService(postRepo)
			postHandler := handler.NewPostHandler(postService)

			post := v1.Group("/posts")
			post.GET("/", postHandler.GetAllPosts)
			post.GET("/:id", postHandler.GetPostByID)
			post.GET("/thread/:thread_id", postHandler.GetPostsByThreadID)
			post.GET("/author/:author_id", postHandler.GetPostsByAuthorID)
			post.POST("/", middleware.JWTMiddleware(), middleware.IsUserBanned(db), middleware.S3Middleware(), postHandler.Create)
			post.POST("/:id/votes", middleware.JWTMiddleware(), middleware.IsUserBanned(db), postHandler.VotePost)
			post.GET("/:id/votes", postHandler.GetPostVotes)
			post.POST("/:id/reactions", middleware.JWTMiddleware(), middleware.IsUserBanned(db), postHandler.ReactPost)
			post.PATCH("/:id", postHandler.Update)
			post.DELETE("/:id", postHandler.Delete)
		}

		{
			tagRepo := repository.NewTagRepository(db)
			tagService := service.NewTagService(tagRepo)
			tagHandler := handler.NewTagHandler(tagService)

			tag := v1.Group("/tags")

			tag.Use(middleware.JWTMiddleware())

			tag.GET("/", tagHandler.GetAllTags)
			tag.POST("/", middleware.IsAdminLogged(), tagHandler.CreateTag)
			tag.GET("/:id", tagHandler.GetTagByID)
			tag.GET("/slug/:slug", tagHandler.GetTagBySlug)
			tag.PATCH("/:id", middleware.IsAdminLogged(), tagHandler.UpdateTag)
			tag.DELETE("/:id", middleware.IsAdminLogged(), tagHandler.DeleteTag)
		}

	}

	return r
}
