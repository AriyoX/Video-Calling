
// DOM elements
let localVideo;
let remoteVideos = {};
let localStream;

// WebRTC connections
let peerConnections = {};

// WebSocket connection
let socket;

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
  // Get the meeting code and participant ID from the URL
  const urlParams = new URLSearchParams(window.location.search);
  const participantId = urlParams.get('participantId');
  const meetingCode = window.location.pathname.split('/').pop();
  
  // Set up WebSocket connection
  connectWebSocket(meetingCode, participantId);
  
  // Set up UI event listeners
  setupUI();
});

// Connect to the WebSocket server
function connectWebSocket(meetingCode, participantId) {
  const name = localStorage.getItem('participantName') || 'Guest';
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${wsProtocol}//${window.location.host}/ws/${meetingCode}/${participantId}?name=${encodeURIComponent(name)}`;
  
  socket = new WebSocket(wsUrl);
  
  socket.onopen = function() {
    console.log('WebSocket connection established');
  };
  
  socket.onmessage = function(event) {
    const message = JSON.parse(event.data);
    handleSocketMessage(message);
  };
  
  socket.onclose = function() {
    console.log('WebSocket connection closed');
    // Try to reconnect after a delay
    setTimeout(() => {
      connectWebSocket(meetingCode, participantId);
    }, 3000);
  };
  
  socket.onerror = function(error) {
    console.error('WebSocket error:', error);
  };
}

// Handle incoming WebSocket messages
function handleSocketMessage(message) {
  switch (message.type) {
    case 'init':
      handleInitMessage(message.content);
      break;
    case 'participant-joined':
      handleParticipantJoined(message.content);
      break;
    case 'participant-left':
      handleParticipantLeft(message.content);
      break;
    case 'signal':
      handleSignalingMessage(message.content);
      break;
    case 'chat':
      handleChatMessage(message.content);
      break;
    case 'waiting-room-update':
      updateWaitingRoom(message.content);
      break;
    case 'admitted':
      handleAdmitted();
      break;
  }
}

// Handle initial connection message
function handleInitMessage(content) {
  const { isHost, isAdmitted, isWaiting } = content;
  
  if (isWaiting) {
    // Show waiting room UI
    document.getElementById('waiting-screen').style.display = 'block';
    document.getElementById('meeting-screen').style.display = 'none';
  } else if (isAdmitted || isHost) {
    // Start the meeting
    document.getElementById('waiting-screen').style.display = 'none';
    document.getElementById('meeting-screen').style.display = 'block';
    
    // If host, show the waiting room controls
    if (isHost) {
      document.getElementById('waiting-room-container').style.display = 'block'; } else {
        document.getElementById('waiting-room-container').style.display = 'none';
      }
      
      // Initialize WebRTC
      initializeWebRTC();
    }
  }
  
  // Initialize WebRTC
  function initializeWebRTC() {
    // Request user media (camera and microphone)
    navigator.mediaDevices.getUserMedia({
      video: true,
      audio: true
    })
    .then(stream => {
      localStream = stream;
      
      // Display local video
      localVideo = document.getElementById('local-video');
      localVideo.srcObject = stream;
      
      // Send offer to other participants
      sendOfferToExistingParticipants();
    })
    .catch(error => {
      console.error('Error accessing media devices:', error);
      alert('Could not access camera or microphone. Please check permissions.');
    });
  }
  
  // Handle a participant joining the meeting
  function handleParticipantJoined(content) {
    const { participantId, name } = content;
    
    // Add to participants list
    addParticipantToList(participantId, name);
    
    // Create a new peer connection for this participant
    createPeerConnection(participantId, true);
  }
  
  // Handle a participant leaving the meeting
  function handleParticipantLeft(content) {
    const { participantId } = content;
    
    // Remove from participants list
    removeParticipantFromList(participantId);
    
    // Close peer connection
    if (peerConnections[participantId]) {
      peerConnections[participantId].close();
      delete peerConnections[participantId];
    }
    
    // Remove video element
    if (remoteVideos[participantId]) {
      const videoContainer = remoteVideos[participantId].parentElement;
      videoContainer.remove();
      delete remoteVideos[participantId];
    }
  }
  
  // Handle signaling messages for WebRTC
  function handleSignalingMessage(content) {
    const { senderId, type, sdp, candidate } = content;
    
    // Create peer connection if it doesn't exist
    if (!peerConnections[senderId]) {
      createPeerConnection(senderId, false);
    }
    
    const peerConnection = peerConnections[senderId];
    
    // Handle different types of signaling messages
    if (type === 'offer') {
      peerConnection.setRemoteDescription(new RTCSessionDescription({
        type: 'offer',
        sdp: sdp
      }))
      .then(() => peerConnection.createAnswer())
      .then(answer => peerConnection.setLocalDescription(answer))
      .then(() => {
        // Send answer to the sender
        sendSignalingMessage(senderId, {
          type: 'answer',
          sdp: peerConnection.localDescription.sdp
        });
      })
      .catch(error => console.error('Error handling offer:', error));
    } else if (type === 'answer') {
      peerConnection.setRemoteDescription(new RTCSessionDescription({
        type: 'answer',
        sdp: sdp
      }))
      .catch(error => console.error('Error handling answer:', error));
    } else if (type === 'candidate') {
      peerConnection.addIceCandidate(new RTCIceCandidate(candidate))
      .catch(error => console.error('Error adding ICE candidate:', error));
    }
  }
  
  // Create a peer connection to a participant
  function createPeerConnection(participantId, isInitiator) {
    // STUN servers to help with NAT traversal
    const configuration = {
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'stun:stun1.l.google.com:19302' }
      ]
    };
    
    // Create the peer connection
    const peerConnection = new RTCPeerConnection(configuration);
    peerConnections[participantId] = peerConnection;
    
    // Add local stream to the connection
    localStream.getTracks().forEach(track => {
      peerConnection.addTrack(track, localStream);
    });
    
    // Handle ICE candidates
    peerConnection.onicecandidate = event => {
      if (event.candidate) {
        sendSignalingMessage(participantId, {
          type: 'candidate',
          candidate: event.candidate
        });
      }
    };
    
    // Handle incoming tracks (remote streams)
    peerConnection.ontrack = event => {
      // Create or get remote video element
      if (!remoteVideos[participantId]) {
        createRemoteVideoElement(participantId);
      }
      
      // Set the stream
      remoteVideos[participantId].srcObject = event.streams[0];
    };
    
    // If we're the initiator, create and send an offer
    if (isInitiator) {
      peerConnection.createOffer()
        .then(offer => peerConnection.setLocalDescription(offer))
        .then(() => {
          sendSignalingMessage(participantId, {
            type: 'offer',
            sdp: peerConnection.localDescription.sdp
          });
        })
        .catch(error => console.error('Error creating offer:', error));
    }
    
    return peerConnection;
  }
  
  // Send a signaling message through the WebSocket
  function sendSignalingMessage(targetId, data) {
    if (socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: 'signal',
        content: {
          targetId: targetId,
          ...data
        }
      }));
    }
  }
  
  // Create a video element for a remote participant
  function createRemoteVideoElement(participantId) {
    const videoGrid = document.getElementById('video-grid');
    
    // Create container
    const videoContainer = document.createElement('div');
    videoContainer.className = 'video-container';
    videoContainer.id = `video-container-${participantId}`;
    
    // Create video element
    const videoElement = document.createElement('video');
    videoElement.autoplay = true;
    videoElement.playsInline = true;
    
    // Add name label
    const nameLabel = document.createElement('div');
    nameLabel.className = 'video-name';
    nameLabel.textContent = `Participant ${participantId}`;
    
    // Add to DOM
    videoContainer.appendChild(videoElement);
    videoContainer.appendChild(nameLabel);
    videoGrid.appendChild(videoContainer);
    
    // Store reference
    remoteVideos[participantId] = videoElement;
  }
  
  // Handle chat messages
  function handleChatMessage(content) {
    const { sender, text, timestamp } = content;
    
    // Add message to chat
    const chatMessages = document.getElementById('chat-messages');
    const messageElement = document.createElement('div');
    messageElement.className = 'message';
    
    const time = new Date(timestamp * 1000).toLocaleTimeString();
    
    messageElement.innerHTML = `
      <span class="message-sender">${sender}</span>
      <span class="message-time">${time}</span>
      <div class="message-text">${text}</div>
    `;
    
    chatMessages.appendChild(messageElement);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }
  
  // Handle being admitted to the meeting
  function handleAdmitted() {
    // Hide waiting screen and show meeting screen
    document.getElementById('waiting-screen').style.display = 'none';
    document.getElementById('meeting-screen').style.display = 'block';
    
    // Initialize WebRTC
    initializeWebRTC();
  }
  
  // Update the waiting room UI
  function updateWaitingRoom(content) {
    const { waitingParticipants } = content;
    const waitingList = document.getElementById('waiting-list');
    
    // Clear current list
    waitingList.innerHTML = '';
    
    // Add waiting participants
    waitingParticipants.forEach(participant => {
      const listItem = document.createElement('li');
      listItem.className = 'participant-item';
      listItem.innerHTML = `
        <span>${participant.name}</span>
        <div class="action-buttons">
          <button class="admit-button" data-id="${participant.id}">Admit</button>
          <button class="reject-button" data-id="${participant.id}">Reject</button>
        </div>
      `;
      waitingList.appendChild(listItem);
    });
    
    // Update waiting room count
    const waitingCount = document.getElementById('waiting-count');
    waitingCount.textContent = waitingParticipants.length;
    
    // Show/hide waiting room section
    const waitingRoom = document.getElementById('waiting-room');
    if (waitingParticipants.length > 0) {
      waitingRoom.style.display = 'block';
    } else {
      waitingRoom.style.display = 'none';
    }
  }
  
  // Add participant to the participants list
  function addParticipantToList(participantId, name) {
    const participantsList = document.getElementById('participants-list');
    
    // Check if already in list
    if (document.getElementById(`participant-${participantId}`)) {
      return;
    }
    
    // Create list item
    const listItem = document.createElement('li');
    listItem.id = `participant-${participantId}`;
    listItem.textContent = name || `Participant ${participantId.substring(0, 5)}`;
    
    participantsList.appendChild(listItem);
  }
  
  // Remove participant from the participants list
  function removeParticipantFromList(participantId) {
    const listItem = document.getElementById(`participant-${participantId}`);
    if (listItem) {
      listItem.remove();
    }
  }
  
  // Set up UI event listeners
  function setupUI() {
    // Chat input
    const chatInput = document.getElementById('chat-input');
    const sendButton = document.getElementById('send-button');
    
    function sendChatMessage() {
      const text = chatInput.value.trim();
      if (text) {
        socket.send(JSON.stringify({
          type: 'chat',
          content: {
            text: text
          }
        }));
        chatInput.value = '';
      }
    }
    
    sendButton.addEventListener('click', sendChatMessage);
    chatInput.addEventListener('keypress', event => {
      if (event.key === 'Enter') {
        sendChatMessage();
      }
    });
    
    // Media controls
    const muteButton = document.getElementById('mute-button');
    const videoButton = document.getElementById('video-button');
    const screenButton = document.getElementById('screen-button');
    const hangupButton = document.getElementById('hangup-button');
    
    let isAudioMuted = false;
    let isVideoOff = false;
    let isScreenSharing = false;
    
    muteButton.addEventListener('click', () => {
      isAudioMuted = !isAudioMuted;
      
      // Toggle microphone
      localStream.getAudioTracks().forEach(track => {
        track.enabled = !isAudioMuted;
      });
      
      // Update button appearance
      muteButton.classList.toggle('disabled', isAudioMuted);
      muteButton.innerHTML = isAudioMuted ? '<i class="fas fa-microphone-slash"></i>' : '<i class="fas fa-microphone"></i>';
    });
    
    videoButton.addEventListener('click', () => {
      isVideoOff = !isVideoOff;
      
      // Toggle camera
      localStream.getVideoTracks().forEach(track => {
        track.enabled = !isVideoOff;
      });
      
      // Update button appearance
      videoButton.classList.toggle('disabled', isVideoOff);
      videoButton.innerHTML = isVideoOff ? '<i class="fas fa-video-slash"></i>' : '<i class="fas fa-video"></i>';
    });
    
    screenButton.addEventListener('click', () => {
      if (!isScreenSharing) {
        // Start screen sharing
        navigator.mediaDevices.getDisplayMedia({ video: true })
          .then(screenStream => {
            // Replace video track with screen track
            const videoTrack = screenStream.getVideoTracks()[0];
            
            // Replace track in all peer connections
            for (const participantId in peerConnections) {
              const sender = peerConnections[participantId]
                .getSenders()
                .find(s => s.track.kind === 'video');
                
              if (sender) {
                sender.replaceTrack(videoTrack);
              }
            }
            
            // Replace local video display
            const oldVideoTrack = localStream.getVideoTracks()[0];
            localStream.removeTrack(oldVideoTrack);
            localStream.addTrack(videoTrack);
            
            // Update local video
            localVideo.srcObject = localStream;
            
            // Handle when screen sharing stops
            videoTrack.onended = () => {
              screenButton.click();
            };
            
            isScreenSharing = true;
            screenButton.classList.add('disabled');
            screenButton.innerHTML = '<i class="fas fa-desktop"></i>';
          })
          .catch(error => {
            console.error('Error sharing screen:', error);
          });
      } else {
        // Stop screen sharing and switch back to camera
        navigator.mediaDevices.getUserMedia({ video: true })
          .then(cameraStream => {
            const videoTrack = cameraStream.getVideoTracks()[0];
            
            // Replace track in all peer connections
            for (const participantId in peerConnections) {
              const sender = peerConnections[participantId]
                .getSenders()
                .find(s => s.track.kind === 'video');
                
              if (sender) {
                sender.replaceTrack(videoTrack);
              }
            }
            
            // Replace in local stream
            const oldVideoTrack = localStream.getVideoTracks()[0];
            localStream.removeTrack(oldVideoTrack);
            localStream.addTrack(videoTrack);
            
            // Update local video
            localVideo.srcObject = localStream;
            
            isScreenSharing = false;
            screenButton.classList.remove('disabled');
            screenButton.innerHTML = '<i class="fas fa-desktop"></i>';
          })
          .catch(error => {
            console.error('Error switching to camera:', error);
          });
      }
    });
    
    hangupButton.addEventListener('click', () => {
      // Close all peer connections
      for (const participantId in peerConnections) {
        peerConnections[participantId].close();
      }
      
      // Stop all tracks
      if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
      }
      
      // Close WebSocket
      if (socket) {
        socket.close();
      }
      
      // Redirect to home
      window.location.href = '/';
    });
    
    // Waiting room controls (for host)
    document.addEventListener('click', event => {
      if (event.target.classList.contains('admit-button')) {
        const participantId = event.target.getAttribute('data-id');
        const meetingCode = window.location.pathname.split('/').pop();
        const hostId = new URLSearchParams(window.location.search).get('participantId');
        
        // Send admit request
        fetch(`/meeting/${meetingCode}/admit/${participantId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: `hostId=${hostId}`
        });
      } else if (event.target.classList.contains('reject-button')) {
        const participantId = event.target.getAttribute('data-id');
        const meetingCode = window.location.pathname.split('/').pop();
        const hostId = new URLSearchParams(window.location.search).get('participantId');
        
        // Send reject request
        fetch(`/meeting/${meetingCode}/reject/${participantId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: `hostId=${hostId}`
        });
      }
    });
    
    // Copy meeting link button
    const copyLinkButton = document.getElementById('copy-link-button');
    if (copyLinkButton) {
      copyLinkButton.addEventListener('click', () => {
        const meetingLink = window.location.origin + window.location.pathname;
        
        // Copy to clipboard
        navigator.clipboard.writeText(meetingLink)
          .then(() => {
            // Show success message
            copyLinkButton.textContent = 'Copied!';
            setTimeout(() => {
              copyLinkButton.textContent = 'Copy Link';
            }, 2000);
          })
          .catch(error => {
            console.error('Error copying link:', error);
          });
      });
    }
  }
  
  // Send offer to existing participants
  function sendOfferToExistingParticipants() {
    // In a real implementation, you'd get the list of existing participants
    // and create peer connections with each one
    console.log('Ready to connect with existing participants');
  }
  
 