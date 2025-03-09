package controllers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/AriyoX/Video-Calling/internal/models"
	"github.com/AriyoX/Video-Calling/internal/services"
	"github.com/AriyoX/Video-Calling/pkg/utils"
)

// MeetingController handles HTTP requests related to meetings
type MeetingController struct {
	meetingService *services.MeetingService
	templates      map[string]*template.Template
}

// NewMeetingController creates a new meeting controller
func NewMeetingController() *MeetingController {
	// Load templates
	templatesDir := "internal/views/templates"
	templates := make(map[string]*template.Template)
	
	for _, page := range []string{"create", "join", "meeting"} {
		templates[page] = template.Must(template.ParseFiles(
			filepath.Join(templatesDir, page+".html"),
		))
	}

	return &MeetingController{
		meetingService: services.NewMeetingService(),
		templates:      templates,
	}
}

// Home renders the home page
func (c *MeetingController) Home(w http.ResponseWriter, r *http.Request) {
	c.templates["create"].Execute(w, nil)
}

// CreateMeeting creates a new meeting and redirects to it
func (c *MeetingController) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	// Generate a random meeting code
	meetingCode := utils.GenerateRandomCode(8)
	
	// Create the meeting with the host as creator
	hostName := r.FormValue("name")
	if hostName == "" {
		hostName = "Host"
	}
	
	hostID := utils.GenerateRandomID()
	meeting := models.NewMeeting(meetingCode, hostID, hostName)
	
	// Save the meeting
	c.meetingService.SaveMeeting(meeting)
	
	// Redirect to the meeting page
	http.Redirect(w, r, "/meeting/"+meetingCode+"?participantId="+hostID, http.StatusSeeOther)
}

// JoinMeeting handles a request to join a meeting
func (c *MeetingController) JoinMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	meetingCode := vars["code"]
	
	// Check if the meeting exists
	meeting, exists := c.meetingService.GetMeeting(meetingCode)
	if !exists {
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}
	
	// Check if the participant is already in the meeting
	participantID := r.URL.Query().Get("participantId")
	if participantID != "" {
		// If this is the host or an admitted participant, show the meeting
		if meeting.IsHost(participantID) || meeting.IsParticipantAdmitted(participantID) {
			data := map[string]interface{}{
				"MeetingCode":   meetingCode,
				"ParticipantID": participantID,
				"IsHost":        meeting.IsHost(participantID),
				"Participants":  meeting.GetAdmittedParticipants(),
				"WaitingRoom":   meeting.GetWaitingParticipants(),
			}
			c.templates["meeting"].Execute(w, data)
			return
		}
	}
	
	// If not in the meeting, show the join page
	c.templates["join"].Execute(w, map[string]string{
		"MeetingCode": meetingCode,
	})
}

// AdmitParticipant admits a participant from the waiting room
func (c *MeetingController) AdmitParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	meetingCode := vars["code"]
	participantID := vars["participantId"]
	
	// Get the host ID from request
	hostID := r.FormValue("hostId")
	
	// Check if the meeting exists
	meeting, exists := c.meetingService.GetMeeting(meetingCode)
	if !exists {
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}
	
	// Check if the requester is the host
	if !meeting.IsHost(hostID) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Admit the participant
	success := meeting.AdmitParticipant(participantID)
	if !success {
		http.Error(w, "Participant not found in waiting room", http.StatusBadRequest)
		return
	}
	
	// Save the updated meeting
	c.meetingService.SaveMeeting(meeting)
	
	// Return success
	w.WriteHeader(http.StatusOK)
}

// RejectParticipant rejects a participant from the waiting room
func (c *MeetingController) RejectParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	meetingCode := vars["code"]
	participantID := vars["participantId"]
	
	// Get the host ID from request
	hostID := r.FormValue("hostId")
	
	// Check if the meeting exists
	meeting, exists := c.meetingService.GetMeeting(meetingCode)
	if !exists {
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}
	
	// Check if the requester is the host
	if !meeting.IsHost(hostID) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Reject the participant
	success := meeting.RejectParticipant(participantID)
	if !success {
		http.Error(w, "Participant not found in waiting room", http.StatusBadRequest)
		return
	}
	
	// Save the updated meeting
	c.meetingService.SaveMeeting(meeting)
	
	// Return success
	w.WriteHeader(http.StatusOK)
}
