package main

import (
	"notification/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userController *controllers.UserController, notifierController *controllers.NotificationController, authMiddleware gin.HandlerFunc) {
	// Public routes
	router.POST("/signup", userController.Signup)
	router.POST("/login", userController.Login)
	router.GET("/notifications/channels/schemas", notifierController.GetChannelSchemas)

	// Protected routes
	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.POST("/notifications", notifierController.CreateNotification)
		protected.GET("/notifications", notifierController.ListNotifications)
		protected.GET("/notifications/:id", notifierController.GetNotification)
		protected.PATCH("/notifications/:id", notifierController.UpdateNotification)
		protected.DELETE("/notifications/:id", notifierController.DeleteNotification)
	}
}
