package utils

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/AriyoX/Video-Calling/internal/models"
	"github.com/AriyoX/Video-Calling/internal/services"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// HandleWebSocket manages a WebSocket connection
func HandleWebSocket(conn *websocket.Conn, meeting *models.Meeting, participantID string, meetingService *services.MeetingService) {
	// Set connection parameters
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Send initial state
	initialMsg := Message{
		Type: "init",
		Content: map[string]interface{}{
			"meetingCode":   meeting.Code,
			"participantId": participantID,
			"isHost":        meeting.IsHost(participantID),
			"isAdmitted":    meeting.IsParticipantAdmitted(participantID),
			"isWaiting":     meeting.IsParticipantWaiting(participantID),
		},
	}
	if err := conn.WriteJSON(initialMsg); err != nil {
		log.Println("Error sending initial message:", err)
		return
	}

	// Handle messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process message based on type
		switch msg.Type {
		case "signal":
			// Forward WebRTC signaling messages to other participants
			forwardSignalingMessage(conn, msg, meeting, participantID)
		case "chat":
			// Forward chat messages to all participants
			forwardChatMessage(conn, msg, meeting, participantID)
		}
	}
}

// forwardSignalingMessage forwards WebRTC signaling messages
func forwardSignalingMessage(conn *websocket.Conn, msg Message, meeting *models.Meeting, senderID string) {
	// Extract message content
	content, ok := msg.Content.(map[string]interface{})
	if !ok {
		log.Println("Invalid signal message format")
		return
	}

	// Add sender information
	content["senderId"] = senderID

	// Forward message to the target participant
	// This would be handled by a more sophisticated message broker in a production app
	// Here we're simplified for the example
	log.Printf("Signal message from %s: %v", senderID, content)
}

// forwardChatMessage forwards chat messages
func forwardChatMessage(conn *websocket.Conn, msg Message, meeting *models.Meeting, senderID string) {
	// Add sender information to the message
	chatMsg, ok := msg.Content.(map[string]interface{})
	if !ok {
		log.Println("Invalid chat message format")
		return
	}

	// Get sender name
	var senderName string
	if meeting.IsHost(senderID) {
		senderName = meeting.HostName
	} else if participant, isAdmitted := meeting.AdmittedParticipants[senderID]; isAdmitted {
		senderName = participant.Name
	} else {
		log.Println("Message from non-participant")
		return
	}

	chatMsg["sender"] = senderName
	chatMsg["senderId"] = senderID
	chatMsg["timestamp"] = time.Now().Unix()

	// In a real implementation, you would broadcast this to all participants
	log.Printf("Chat message from %s: %v", senderName, chatMsg["text"])
}
