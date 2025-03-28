package http

import (
	groupHandler "EduSync/internal/delivery/http/group"
	instituteHandler "EduSync/internal/delivery/http/institution"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/delivery/middleware"
	userRepository "EduSync/internal/repository"
	"EduSync/internal/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupRouter(tokenRepo userRepository.TokenRepository,
	authHandler *user.AuthHandler,
	jwtManager *util.JWTManager,
	groupHandler *groupHandler.GroupHandler,
	instHandler *instituteHandler.InstitutionHandler) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/register", authHandler.RegisterHandler)
		api.POST("/login", authHandler.LoginHandler)
		api.POST("/logout", authHandler.LogoutHandler)
		api.POST("/refresh", authHandler.RefreshTokenHandler)

		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware(tokenRepo, jwtManager))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				email, _ := c.Get("email")
				fullName, _ := c.Get("full_name")
				isTeacher, _ := c.Get("is_teacher")
				c.JSON(http.StatusOK, gin.H{
					"user_id":    userID,
					"email":      email,
					"full_name":  fullName,
					"is_teacher": isTeacher,
				})
			})
		}
		group := api.Group("/group")
		{
			group.GET("/institution/:institution_id", groupHandler.GetGroupsByInstitutionID)
			group.GET("/:id", groupHandler.GetGroupByID)
		}

		api.GET("/institutions", instHandler.GetAllInstitutions)
		api.GET("/institutions/:id", instHandler.GetInstitutionByID)

	}

	return router
}
