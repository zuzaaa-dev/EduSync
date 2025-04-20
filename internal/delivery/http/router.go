package http

import (
	chatHandler "EduSync/internal/delivery/http/chat"
	groupHandler "EduSync/internal/delivery/http/group"
	instituteHandler "EduSync/internal/delivery/http/institution"
	messageHandler "EduSync/internal/delivery/http/message"
	scheduleHandler "EduSync/internal/delivery/http/schedule"
	subjectHandler "EduSync/internal/delivery/http/subject"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/delivery/middleware"
	"EduSync/internal/repository"
	"EduSync/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func SetupRouter(
	tokenRepo repository.TokenRepository,
	chatRepo repository.ChatRepository,
	authHandler *user.AuthHandler,
	jwtManager *util.JWTManager,
	groupHandler *groupHandler.GroupHandler,
	instHandler *instituteHandler.InstitutionHandler,
	subjectHandler *subjectHandler.InstitutionHandler,
	scheduleHandler *scheduleHandler.ScheduleHandler,
	chatHandler *chatHandler.ChatHandler,
	messageHandler *messageHandler.MessageHandler,
	teacherInitHandler *scheduleHandler.TeacherInitialsHandler,
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
				schedule.GET("/initials", teacherInitHandler.ListHandler)
				schedule.GET("/teacher_initials/:initials_id", scheduleHandler.GetByTeacherInitialsHandler)
				schedule.PATCH("/:id", scheduleHandler.UpdateHandler)
				schedule.DELETE("/:id", scheduleHandler.DeleteHandler)
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
				group_id, _ := c.Get("group_id")
				institution_id, _ := c.Get("institution_id")
				c.JSON(http.StatusOK, gin.H{
					"user_id":        userID,
					"email":          email,
					"full_name":      fullName,
					"is_teacher":     isTeacher,
					"group_id":       group_id,
					"institution_id": institution_id,
				})
			})
			chatGroup := protected.Group("/chats")
			chatGroup.POST("/:id/join", chatHandler.JoinChatHandler)
			chatGroup.Use(middleware.ChatMembershipMiddleware(chatRepo))
			{
				chatGroup.POST("", chatHandler.CreateChatHandler)
				chatGroup.GET("/:id/participants", chatHandler.GetParticipantsHandler)
				chatGroup.PUT("/:id/invite", chatHandler.UpdateInviteHandler)
				chatGroup.DELETE("/:id", chatHandler.DeleteChatHandler)
				chatGroup.DELETE("/:id/participants/:userID", chatHandler.RemoveParticipantHandler)
				chatGroup.DELETE("/:id/leave", chatHandler.LeaveChatHandler)

				messages := chatGroup.Group("/:id/messages")
				{
					messages.GET("", messageHandler.GetMessagesHandler)
					messages.POST("", messageHandler.SendMessageHandler)
					messages.DELETE("/:messageID", messageHandler.DeleteMessageHandler)
					messages.POST("/:messageID/reply", messageHandler.ReplyMessageHandler)
					messages.GET("/search", messageHandler.SearchMessagesHandler)
				}
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
