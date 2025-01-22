package main

import (
    "fmt"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/joicejoseph3198/notification-service/internals/config"
    "github.com/joicejoseph3198/notification-service/internals/handler"
    "github.com/joicejoseph3198/notification-service/internals/service"
)

func main() {
    config.InitRedis()
    
    // Initialize services and handlers
    sseService := service.NewSSEService()
    handler := handlers.NewSSEHandler(sseService)
    
    // Setup Gin router
    router := gin.Default()
    
    // Routes
    router.GET("/live-notification", handler.HandleLiveNotifications)
    router.POST("/push-message", handler.HandleMessageBroadcast)
    
    // Background tasks
    go sseService.SubscribeToChannel("auction-updates")
    go sseService.TriggerHeartBeatEvents(30 * time.Second)
    
    // Start server
    port := ":8080"
    fmt.Printf("Server started on %s\n", port)
    router.Run(port)
}