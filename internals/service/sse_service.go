package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"github.com/joicejoseph3198/notification-service/internals/config"
)

// Defines structure of the message being sent across the channels
type SSEMessage struct {
	EventType string `json:"eventType"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type RedisMessage struct {
	AuctionId        string `json:"auctionId"`
	SubscriberEmail  string `json:"subscriberEmail,omitempty"`
	EventType        string `json:"eventType"`
	Message          string `json:"message"`
	Timestamp        int64  `json:"timestamp"`
}


// keeps track of all SSE connections established with the server.
// Map<AuctionId, Map<Subscriber, Channel>> 
type SSEService struct{
	mu sync.Mutex
	clients map[string]map[string]chan SSEMessage
}

func NewSSEService() *SSEService {
	return &SSEService{
		clients: make(map[string]map[string]chan SSEMessage),
	}
}

func (s *SSEService) RegisterClient(auctionId string, subscriberEmail string, channel chan SSEMessage){
	s.mu.Lock()
	defer s.mu.Unlock()

	// create a map for the auction if it doesnt exists,
	// otherwise simply add to the map
	if _,exists := s.clients[auctionId]; !exists{
		s.clients[auctionId] = make(map[string]chan SSEMessage)
	}
	s.clients[auctionId][subscriberEmail] = channel
}

func (s *SSEService) RemoveClient(auctionId string, subscriberEmail string){
	s.mu.Lock()
	defer s.mu.Unlock()
	if clients,exists := s.clients[auctionId]; exists{
		if channel, present := clients[subscriberEmail]; present{
			close(channel)
			delete(clients,subscriberEmail)
		}
		if len(clients) == 0{
			delete(s.clients,auctionId)
		}
	}
}

func (s *SSEService) BroadcastMessage(auctionId string, message SSEMessage) error{
	s.mu.Lock()
	clients, exists := s.clients[auctionId]
	
	// "get what you need under lock, release quickly, then work with the copy" is a common Go concurrency pattern,
	// especially when dealing with maps and slices that need to be iterated over.
	
	s.mu.Unlock()
	if !exists{
		log.Printf("Auction not found: auctionId=%s", auctionId)
		return fmt.Errorf("auction %s not found", auctionId)
	}

	for email, ch := range clients{
		// Start a goroutine for each client
		// Launch anonymous function/closure as goroutine
		go func(email string, ch chan SSEMessage){
			select{
			case ch <- message:
			default:
				log.Printf("Removing client due to failed broadcast: auctionId=%s, subscriberEmail=%s", auctionId, email)
				s.RemoveClient(auctionId,email)
			}
		}(email, ch)  //Pass current values of email and ch
	}
	log.Printf("broadcasting message: auctionId=%s", auctionId)

	return nil;
}

func (s *SSEService) BroadcastMessageToClient(auctionId string, subscriberEmail string, message SSEMessage) error{
	s.mu.Lock()

	clients, exists := s.clients[auctionId]
	if !exists{
		s.mu.Unlock()
		return fmt.Errorf("auction %s not found", auctionId)
	}

	ch, exists := clients[subscriberEmail]
	s.mu.Unlock()
	if !exists {
		return fmt.Errorf("subscriber %s not found for auction %s", subscriberEmail, auctionId)
	}

	go func(email string, ch chan SSEMessage){
		select{
		case ch <- message:
		default:
			log.Printf("Removing client due to failed broadcast: auctionId=%s, subscriberEmail=%s", auctionId, email)
			s.RemoveClient(auctionId,email)
		}
	}(subscriberEmail, ch) 
	log.Printf("broadcasting message: auctionId=%s , subscriberId=%s", auctionId, subscriberEmail)
	return nil;
}


func (s *SSEService) TriggerHeartBeatEvents(interval time.Duration){
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C{
		s.mu.Lock()
		auctionIds := make([]string, 0, len(s.clients))
		for id := range s.clients {
			auctionIds = append(auctionIds, id)
		}
		s.mu.Unlock()
		for _,auctionId := range auctionIds{
			s.BroadcastMessage(auctionId, SSEMessage{"HEART BEAT","Ping", time.Now().Unix()})
		}
	}
}

func (s *SSEService) SubscribeToChannel(channel string) {
	pubsub := config.RedisClient.Subscribe(context.Background(), channel)
	defer pubsub.Close()

	fmt.Printf("Subscribed to Redis channel: %s\n", channel)

	for msg := range pubsub.Channel() {
		var message RedisMessage

		err := json.Unmarshal([]byte(msg.Payload), &message)
		if err != nil {
			log.Println("Error processing message:", err)
			continue
		}
		// Broadcast the received message using the provided function
		broadcastUpdate := SSEMessage{message.EventType, message.Message, message.Timestamp}
		s.BroadcastMessage(message.AuctionId, broadcastUpdate)
		if message.SubscriberEmail != "" {
			userSpecificUpdate := SSEMessage{message.EventType, message.Message, message.Timestamp}
			s.BroadcastMessageToClient(message.AuctionId, message.SubscriberEmail,userSpecificUpdate)
		}
	}
}