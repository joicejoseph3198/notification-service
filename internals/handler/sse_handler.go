package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joicejoseph3198/notification-service/internals/service"
)

type SSEHandler struct {
    SSEService *service.SSEService
}

func NewSSEHandler(sseService *service.SSEService) *SSEHandler {
    return &SSEHandler{SSEService: sseService}
}

func (h *SSEHandler) HandleLiveNotifications(c *gin.Context) {
    auctionId := c.Query("auctionId")
    email := c.Query("subscriberEmail")

    if auctionId == "" || email == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing auction/subscriber details"})
        return
    }

    log.Printf("New SSE connection: auctionId=%s, subscriberEmail=%s", auctionId, email)

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("Access-Control-Allow-Origin", "*")

    clientChan := make(chan service.SSEMessage, 10)
    h.SSEService.RegisterClient(auctionId, email, clientChan)
    defer h.SSEService.RemoveClient(auctionId, email)

    c.Stream(func(w io.Writer) bool {
        if msg, ok := <-clientChan; ok {
            data, _ := json.Marshal(msg)
            c.SSEvent("data", string(data))
            return true
        }
        return false
    })
}

func (h *SSEHandler) HandleMessageBroadcast(c *gin.Context) {
    auctionId := c.Query("auctionId")

    var message service.SSEMessage
    if err := c.BindJSON(&message); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    if message.EventType == "" || message.Message == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    h.SSEService.BroadcastMessage(auctionId, message)
    c.JSON(http.StatusOK, gin.H{"message": "Message broadcasted successfully"})
}