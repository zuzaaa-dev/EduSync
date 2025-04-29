package http

import (
	_ "EduSync/docs/swagger"
	chatHandler "EduSync/internal/delivery/http/chat"
	"EduSync/internal/delivery/http/favorite"
	groupHandler "EduSync/internal/delivery/http/group"
	instituteHandler "EduSync/internal/delivery/http/institution"
	materialHandler "EduSync/internal/delivery/http/material"
	messageHandler "EduSync/internal/delivery/http/message"
	scheduleHandler "EduSync/internal/delivery/http/schedule"
	subjectHandler "EduSync/internal/delivery/http/subject"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/delivery/middleware"
	"EduSync/internal/repository"
	"EduSync/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	tokenRepo repository.TokenRepository,
	chatRepo repository.ChatRepository,
	authHandler *user.AuthHandler,
	jwtManager *util.JWTManager,
	groupHandler *groupHandler.GroupHandler,
	instHandler *instituteHandler.InstitutionHandler,
	subjectHandler *subjectHandler.SubjectHandler,
	scheduleHandler *scheduleHandler.ScheduleHandler,
	chatHandler *chatHandler.ChatHandler,
	messageHandler *messageHandler.MessageHandler,
	materialHandler *materialHandler.MaterialHandler,
	teacherInitHandler *scheduleHandler.TeacherInitialsHandler,
	fileFavHandler *favorite.FileFavoriteHandler,
	pollHandler *chatHandler.PollHandler,
	log *logrus.Logger,
) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/register", authHandler.RegisterHandler)
		api.POST("/login", authHandler.LoginHandler)

		api.POST("/refresh", authHandler.RefreshTokenHandler)

		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware(tokenRepo, jwtManager, log))
		{
			protected.PUT("/profile", authHandler.UpdateProfileHandler)
			protected.POST("/logout", authHandler.LogoutHandler)
			protected.GET("/profile", authHandler.ProfileHandler)
			schedule := protected.Group("/schedule")
			{
				schedule.GET("/", scheduleHandler.GetScheduleHandler)
				schedule.POST("/", scheduleHandler.CreateHandler)
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

			chatGroup := protected.Group("/chats")
			chatGroup.GET("", chatHandler.ListChatsHandler)
			chatGroup.POST("/:id/join", chatHandler.JoinChatHandler)
			chatGroup.POST("", chatHandler.CreateChatHandler)
			chatGroup.Use(middleware.ChatMembershipMiddleware(chatRepo))
			{
				chatGroup.GET("/:id/participants", chatHandler.GetParticipantsHandler)
				chatGroup.PUT("/:id/invite", chatHandler.UpdateInviteHandler)
				chatGroup.DELETE("/:id", chatHandler.DeleteChatHandler)
				chatGroup.DELETE("/:id/participants/:userID", chatHandler.RemoveParticipantHandler)
				chatGroup.DELETE("/:id/leave", chatHandler.LeaveChatHandler)
				//chatGroup.Static("/files", "./uploads")

				messages := chatGroup.Group("/:id/messages")
				{
					messages.GET("", messageHandler.GetMessagesHandler)
					messages.POST("", messageHandler.SendMessageHandler)
					messages.DELETE("/:messageID", messageHandler.DeleteMessageHandler)
					messages.POST("/:messageID/reply", messageHandler.ReplyMessageHandler)
					messages.GET("/search", messageHandler.SearchMessagesHandler)
				}
				protected.GET("/files/:id", materialHandler.GetFileHandler)
				protected.GET("/files/favorites", fileFavHandler.ListFavoriteFiles)
				protected.POST("/files/:id/favorite", fileFavHandler.AddFavoriteFile)
				protected.DELETE("/files/:id/favorite", fileFavHandler.RemoveFavoriteFile)

				polls := chatGroup.Group("/:id/polls")
				{
					polls.GET("", pollHandler.ListPollsHandler)
					polls.POST("", pollHandler.CreatePoll)
					polls.DELETE("/:poll_id", pollHandler.DeletePoll)
					polls.POST("/:poll_id/vote", pollHandler.Vote)
					polls.DELETE("/:poll_id/vote", pollHandler.UnvoteHandler)
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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
