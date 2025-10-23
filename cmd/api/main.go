// @title Notification API
// @version 1.0
// @description API for managing notifications.
// @description
// @description **Channel Meta Requirements:**
// @description - Email: channels.ValidEmailMeta
// @description - SMS: channels.ValidSMSMeta
// @description - Push: channels.ValidPushMeta
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification/channels"
	"notification/cmd/api/middleware"
	"notification/controllers"
	_ "notification/docs"
	"notification/models"
	"notification/models/channel"
	"notification/services/notifier"
	usersvc "notification/services/user"
	"notification/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	_ = godotenv.Load(".env", "../../.env", "../.env")

	db, err := storage.NewConnection(storage.Config{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		DBName:   os.Getenv("MYSQL_DB"),
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	db.Debug()
	db.AutoMigrate(&models.User{}, &models.Notification{}, &models.Outbox{})

	// Initialize notifier service
	channelList := map[string]channel.Channel{
		"email": &channels.EmailChannel{},
		"sms":   &channels.SMSChannel{},
		"push":  &channels.PushChannel{},
	}

	notifierService := notifier.NewNotifierService(db, channelList)
	notifierController := controllers.NewNotificationController(notifierService)

	// Initialize worker
	worker := notifier.NewWorker(db, notifierService, 30*time.Second, 1)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go worker.Start(ctx)

	router := gin.Default()
	// Users routes
	userService := usersvc.New(db)
	userController := controllers.NewUserController(userService)

	// Setup routes and middleware
	SetupRoutes(router, userController, notifierController, middleware.AuthMiddleware(userService))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{Addr: ":8080", Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()
	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
}
