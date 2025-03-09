package controllers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/AriyoX/Video-Calling/internal/services"
	"github.com/AriyoX/Video-Calling/pkg/utils"
)

// WebSocketController handles WebSocket connections
type WebSocketController struct {
	meetingService *services.MeetingService
	upgrader       websocket.Upgrader
}

// NewWebSocketController creates a new WebSocket controller
func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		meetingService: services.NewMeetingService(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections for now
			},
		},
	}
}

// HandleConnection handles WebSocket connections
func (c *WebSocketController) HandleConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	meetingCode := vars["code"]
	participantID := vars["participantId"]
	
	// Check if the meeting exists
	meeting, exists := c.meetingService.GetMeeting(meetingCode)
	if !exists {
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}
	
	// Check if the participant is the host or admitted
	if !meeting.IsHost(participantID) && !meeting.IsParticipantAdmitted(participantID) {
		// If not waiting, add to waiting room
		if !meeting.IsParticipantWaiting(participantID) {
			name := r.URL.Query().Get("name")
			if name == "" {
				name = "Guest"
			}
			meeting.AddToWaitingRoom(participantID, name)
			c.meetingService.SaveMeeting(meeting)
		}
		
		// Upgrade to WebSocket for waiting participants too so they can be notified
	}
	
	// Upgrade HTTP connection to WebSocket
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	
	// Register connection
	utils.HandleWebSocket(conn, meeting, participantID, c.meetingService)
}
