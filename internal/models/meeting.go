package models

import (
	"sync"
	"time"
)

// Meeting represents a video conference meeting
type Meeting struct {
	Code             string
	HostID           string
	HostName         string
	CreatedAt        time.Time
	AdmittedParticipants map[string]Participant
	WaitingParticipants  map[string]Participant
	mutex            sync.RWMutex
}

// NewMeeting creates a new meeting
func NewMeeting(code, hostID, hostName string) *Meeting {
	host := Participant{
		ID:   hostID,
		Name: hostName,
	}
	
	return &Meeting{
		Code:             code,
		HostID:           hostID,
		HostName:         hostName,
		CreatedAt:        time.Now(),
		AdmittedParticipants: map[string]Participant{
			hostID: host,
		},
		WaitingParticipants: make(map[string]Participant),
	}
}

// IsHost checks if a participant is the host
func (m *Meeting) IsHost(participantID string) bool {
	return participantID == m.HostID
}

// IsParticipantAdmitted checks if a participant is admitted to the meeting
func (m *Meeting) IsParticipantAdmitted(participantID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	_, exists := m.AdmittedParticipants[participantID]
	return exists
}

// IsParticipantWaiting checks if a participant is in the waiting room
func (m *Meeting) IsParticipantWaiting(participantID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	_, exists := m.WaitingParticipants[participantID]
	return exists
}

// AddToWaitingRoom adds a participant to the waiting room
func (m *Meeting) AddToWaitingRoom(participantID, name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.WaitingParticipants[participantID] = Participant{
		ID:   participantID,
		Name: name,
	}
}

// AdmitParticipant moves a participant from waiting room to the meeting
func (m *Meeting) AdmitParticipant(participantID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	participant, exists := m.WaitingParticipants[participantID]
	if !exists {
		return false
	}
	
	m.AdmittedParticipants[participantID] = participant
	delete(m.WaitingParticipants, participantID)
	return true
}

// RejectParticipant removes a participant from the waiting room
func (m *Meeting) RejectParticipant(participantID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	_, exists := m.WaitingParticipants[participantID]
	if !exists {
		return false
	}
	
	delete(m.WaitingParticipants, participantID)
	return true
}

// GetAdmittedParticipants returns all admitted participants
func (m *Meeting) GetAdmittedParticipants() []Participant {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	participants := make([]Participant, 0, len(m.AdmittedParticipants))
	for _, p := range m.AdmittedParticipants {
		participants = append(participants, p)
	}
	
	return participants
}

// GetWaitingParticipants returns all waiting participants
func (m *Meeting) GetWaitingParticipants() []Participant {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	participants := make([]Participant, 0, len(m.WaitingParticipants))
	for _, p := range m.WaitingParticipants {
		participants = append(participants, p)
	}
	
	return participants
}
