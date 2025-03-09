package services

import (
	"sync"

	"github.com/AriyoX/Video-Calling/internal/models"
)

// MeetingService manages meetings
type MeetingService struct {
	meetings map[string]*models.Meeting
	mutex    sync.RWMutex
}

// NewMeetingService creates a new meeting service
func NewMeetingService() *MeetingService {
	return &MeetingService{
		meetings: make(map[string]*models.Meeting),
	}
}

// SaveMeeting saves or updates a meeting
func (s *MeetingService) SaveMeeting(meeting *models.Meeting) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.meetings[meeting.Code] = meeting
}

// GetMeeting retrieves a meeting by its code
func (s *MeetingService) GetMeeting(code string) (*models.Meeting, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	meeting, exists := s.meetings[code]
	return meeting, exists
}

// DeleteMeeting removes a meeting
func (s *MeetingService) DeleteMeeting(code string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.meetings, code)
}
