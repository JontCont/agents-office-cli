package workforce

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for local execution
		return true
	},
}

type WSServer struct {
	coordinator *Coordinator
	clients     map[*websocket.Conn]bool
	clientsMu   sync.Mutex
	broadcast   chan Event
}

func NewWSServer(coord *Coordinator) *WSServer {
	return &WSServer{
		coordinator: coord,
		clients:     make(map[*websocket.Conn]bool),
		broadcast:   make(chan Event, 100),
	}
}

func (s *WSServer) Start(addr string) error {
	// Run the broadcast loop
	go s.runBroadcastLoop()

	http.HandleFunc("/ws", s.handleConnection)
	return http.ListenAndServe(addr, nil)
}

// ServeOnMux registers the /ws route on the provided mux and starts the
// broadcast loop. The caller is responsible for calling http.Serve.
func (s *WSServer) ServeOnMux(mux *http.ServeMux) {
	go s.runBroadcastLoop()
	mux.HandleFunc("/ws", s.handleConnection)
}

func (s *WSServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	log.Printf("Client connected to WebSocket server")

	// Emit initial state when a client connects
	initialStateEvent := Event{
		ID:        "init-state",
		Type:      "state.change",
		Timestamp: 0,
		Sender:    "system",
		Content:   string(s.coordinator.GetState()),
	}
	_ = conn.WriteJSON(initialStateEvent)

	// Read loop for upstream commands
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
		log.Printf("Client disconnected from WebSocket server")
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var cmd Command
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Error unmarshaling command: %v", err)
			continue
		}

		log.Printf("Received upstream command: %+v", cmd)
		switch cmd.Type {
		case "run.start":
			_ = s.coordinator.StartRun(cmd.Message)
		case "run.interrupt":
			_ = s.coordinator.Interrupt()
		case "run.resume":
			_ = s.coordinator.Resume(cmd.Message)
		case "run.abort":
			_ = s.coordinator.Abort()
		}
	}
}

func (s *WSServer) BroadcastEvent(evt Event) {
	s.broadcast <- evt
}

func (s *WSServer) runBroadcastLoop() {
	for evt := range s.broadcast {
		s.clientsMu.Lock()
		for conn := range s.clients {
			err := conn.WriteJSON(evt)
			if err != nil {
				conn.Close()
				delete(s.clients, conn)
			}
		}
		s.clientsMu.Unlock()
	}
}
