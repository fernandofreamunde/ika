package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fernandofreamunde/ika/internal/user"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", s.HelloWorldHandler)
	mux.HandleFunc("POST /register", s.RegisterUserHandler)
	mux.HandleFunc("/login", s.LoginHandler)
	mux.HandleFunc("/new_message", s.NewMessageHandler)
	mux.HandleFunc("/chatrooms", s.NewChatRoomHandler)

	mux.HandleFunc("/health", s.healthHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	respondWithJson(resp, 200, w)
}

func (s *Server) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {

	type Parameters struct {
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	_ = decoder.Decode(&params)

	resp, _ := user.CreateUser(params.Email, params.Password, params.Nickname, r.Context(), s.db.Queries().CreateUser)
	respondWithJson(resp, 201, w)
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Login User!"}
	respondWithJson(resp, 200, w)
}

func (s *Server) NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "New Message!"}
	respondWithJson(resp, 200, w)
}

func (s *Server) NewChatRoomHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Register Chat Rooms!"}
	respondWithJson(resp, 200, w)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func respondWithJson(payload interface{}, statusCode int, w http.ResponseWriter) {

	jsonResp, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
