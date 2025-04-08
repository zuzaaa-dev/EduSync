package http

import (
	chatHandler "EduSync/internal/delivery/http/chat"
	groupHandler "EduSync/internal/delivery/http/group"
	instituteHandler "EduSync/internal/delivery/http/institution"
	scheduleHandler "EduSync/internal/delivery/http/schedule"
	subjectHandler "EduSync/internal/delivery/http/subject"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/delivery/middleware"
	userRepository "EduSync/internal/repository"
	"EduSync/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func SetupRouter(
	tokenRepo userRepository.TokenRepository,
	authHandler *user.AuthHandler,
	jwtManager *util.JWTManager,
	groupHandler *groupHandler.GroupHandler,
	instHandler *instituteHandler.InstitutionHandler,
	subjectHandler *subjectHandler.InstitutionHandler,
	scheduleHandler *scheduleHandler.ScheduleHandler,
	chatHandler *chatHandler.ChatHandler,
	log *logrus.Logger,
) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/register", authHandler.RegisterHandler)
		api.POST("/login", authHandler.LoginHandler)
		api.POST("/logout", authHandler.LogoutHandler)
		api.POST("/refresh", authHandler.RefreshTokenHandler)

		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware(tokenRepo, jwtManager, log))
		{
			schedule := protected.Group("/schedule")
			{
				schedule.GET("/", scheduleHandler.GetScheduleHandler)
				schedule.POST("/update", scheduleHandler.UpdateScheduleHandler)
			}
			subject := protected.Group("/subject")
			{
				subject.GET("/institution/:institution_id", subjectHandler.GetSubjectsByInstitution)
				subject.GET("/group/:group_id", subjectHandler.GetSubjectsByGroup)
			}
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

			chatGroup := protected.Group("/chats")
			{
				chatGroup.POST("", chatHandler.CreateChatHandler)                                       // Создание чата
				chatGroup.POST("/:id/join", chatHandler.JoinChatHandler)                                // Присоединиться к чату (для участников)
				chatGroup.GET("/:id/participants", chatHandler.GetParticipantsHandler)                  // Список участников
				chatGroup.DELETE("/:chatID", chatHandler.DeleteChatHandler)                             // Удаление чата
				chatGroup.PUT("/:id/invite", chatHandler.UpdateInviteHandler)                           // Пересоздание приглашения
				chatGroup.DELETE("/:chatID/participants/:userID", chatHandler.RemoveParticipantHandler) // Удаление участника
				chatGroup.DELETE("/:chatID/leave", chatHandler.LeaveChatHandler)                        // Покинуть чат
			}

		}
		group := api.Group("/group")
		{
			group.GET("/institution/:institution_id", groupHandler.GetGroupsByInstitutionID)
			group.GET("/:id", groupHandler.GetGroupByID)
		}
		institutions := api.Group("/institutions")
		{
			institutions.GET("/", instHandler.GetAllInstitutions)
			institutions.GET("/:id", instHandler.GetInstitutionByID)
			institutions.GET("/mask", instHandler.GetAllMasks)
		}

	}

	return router
}
