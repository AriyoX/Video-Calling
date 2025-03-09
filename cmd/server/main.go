package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AriyoX/Video-Calling/internal/config"
	"github.com/AriyoX/Video-Calling/internal/controllers"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create a new router
	router := mux.NewRouter()

	// Create controllers
	meetingController := controllers.NewMeetingController()
	wsController := controllers.NewWebSocketController()

	// Register routes
	// Meeting routes
	router.HandleFunc("/", meetingController.Home).Methods("GET")
	router.HandleFunc("/meeting/create", meetingController.CreateMeeting).Methods("POST")
	router.HandleFunc("/meeting/{code}", meetingController.JoinMeeting).Methods("GET")
	router.HandleFunc("/meeting/{code}/admit/{participantId}", meetingController.AdmitParticipant).Methods("POST")
	router.HandleFunc("/meeting/{code}/reject/{participantId}", meetingController.RejectParticipant).Methods("POST")

	// WebSocket routes
	router.HandleFunc("/ws/{code}/{participantId}", wsController.HandleConnection)

	// Serve static files
	workDir, _ := os.Getwd()
	staticDir := filepath.Join(workDir, "internal/views/static")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Start the server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Starting server on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}
