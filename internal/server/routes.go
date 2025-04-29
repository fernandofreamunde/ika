package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fernandofreamunde/ika/internal/auth"
	"github.com/fernandofreamunde/ika/internal/db"
	"github.com/fernandofreamunde/ika/internal/user"
	"github.com/google/uuid"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	//mux.HandleFunc("GET /", s.HelloWorldHandler)
	mux.HandleFunc("POST /api/users", s.RegisterUserHandler)
	mux.HandleFunc("PUT /api/users/{userID}", s.UpdateUserHandler)
	mux.HandleFunc("POST /api/login", s.LoginHandler)
	mux.HandleFunc("POST /api/refresh", s.RefreshLoginHandler)
	mux.HandleFunc("POST /api/revoke", s.RevokeLoginHandler)

	mux.HandleFunc("POST /api/new_message", s.NewMessageHandler)

	mux.HandleFunc("POST /api/chatrooms", s.CreateChatroomHandler)
	mux.HandleFunc("GET /api/chatrooms", s.GetChatroomsHandler)
	mux.HandleFunc("DELETE /api/chatrooms/{chatroomID}", s.LeaveChatroomHandler)

	mux.HandleFunc("GET /api/chatrooms/{chatroomID}/messages", s.ReadMessagesHandler)
	mux.HandleFunc("POST /api/chatrooms/{chatroomID}/messages", s.CreateMessageHandler)

	mux.HandleFunc("GET /api/health", s.healthHandler)

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

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {

	// TODO: refactor this... check order of events
	type Parameters struct {
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	userID, err := uuid.Parse(r.PathValue("userID"))
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	_ = decoder.Decode(&params)

	var hpw string
	if params.Password == "" {
		hpw = ""
	} else {
		hpw, err = auth.HashPassword(params.Password)
	}

	if err != nil {
		resp := map[string]string{"message": "Something whent wrong processing the request!"}
		respondWithJson(resp, 500, w)
		return
	}

	tokenString, _ := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, "IneedAnAppSecret")
	if err != nil {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	if userID != userId {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	u, _ := s.db.Queries().FindUserById(r.Context(), userId)
	resp, err := user.UpdateUser(u, params.Email, params.Nickname, hpw, r.Context(), s.db.Queries)
	if err != nil {
		resp := map[string]string{"message": err.Error()}
		log.Println(err)
		respondWithJson(resp, 422, w)
		return
	}

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
	hpw, err := auth.HashPassword(params.Password)
	if err != nil {
		resp := map[string]string{"message": "Something whent wrong processing the request!"}
		respondWithJson(resp, 500, w)
		return
	}

	resp, err := user.CreateUser(params.Email, params.Nickname, hpw, r.Context(), s.db.Queries)
	if err != nil {
		resp := map[string]string{"message": err.Error()}
		log.Println(err)
		respondWithJson(resp, 422, w)
		return
	}

	respondWithJson(resp, 201, w)
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	_ = decoder.Decode(&params)

	dbUser, err := s.db.Queries().FindUserByEmail(r.Context(), params.Email)

	if err != nil {
		respondSimpleMessage("Incorrect email of password", 401, w)
		return
	}

	e := auth.CheckPasswordHash(dbUser.HashedPassword, params.Password)
	if e != nil {
		respondSimpleMessage("Incorrect email of password", 401, w)
		return
	}

	resp, _ := auth.AuthenticateUser(dbUser, r.Context(), s.db.Queries)

	respondWithJson(resp, 200, w)
}

func (s *Server) RefreshLoginHandler(w http.ResponseWriter, r *http.Request) {

	jwt, err := auth.RefreshJWT(r.Header, r.Context(), s.db.Queries)
	if err != nil {
		log.Printf("Unauthorized with error: %v", err)
		respondSimpleMessage("Unauthorized.", 401, w)
		return
	}

	type Response struct {
		Token string `json:"token"`
	}

	respondWithJson(Response{
		Token: jwt,
	}, 200, w)
}

func (s *Server) RevokeLoginHandler(w http.ResponseWriter, r *http.Request) {
	auth.RevokeRefreshToken(r.Header, r.Context(), s.db.Queries)
	respondSimpleMessage("", 204, w)
}

func (s *Server) NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "New Message!"}
	respondWithJson(resp, 200, w)
}

func (s *Server) CreateChatroomHandler(w http.ResponseWriter, r *http.Request) {

	tokenString, _ := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, "IneedAnAppSecret")
	if err != nil {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	type Parameters struct {
		FriendID string `json:"friend_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	_ = decoder.Decode(&params)

	user, _ := s.db.Queries().FindUserById(r.Context(), userId)

	frienID, err := uuid.Parse(params.FriendID)
	if err != nil {
		respondSimpleMessage("Invalid Friend ID.", 400, w)
		return
	}

	frien, err := s.db.Queries().FindUserById(r.Context(), frienID)
	if err != nil {
		log.Printf("Err Finding Fren: %v", err)
		respondSimpleMessage("Friend not found.", 404, w)
		return
	}

	room, err := s.db.Queries().CreateChatroom(r.Context(), db.CreateChatroomParams{
		ID:   uuid.New(),
		Name: sql.NullString{String: fmt.Sprintf("%s:%s", user.Nickname, frien.Nickname), Valid: true},
		Type: "direct",
	})

	if err != nil {
		log.Printf("Err Creating room: %v", err)
		respondSimpleMessage("Internal Server Error.", 500, w)
		return
	}

	err = s.db.Queries().ChatroomAddParticipant(r.Context(), db.ChatroomAddParticipantParams{
		ChatroomID:    uuid.NullUUID{UUID: room.ID, Valid: true},
		ParticipantID: uuid.NullUUID{UUID: user.ID, Valid: true},
	})
	if err != nil {
		log.Printf("Err Creating room: %v", err)
		respondSimpleMessage("Internal Server Error.", 500, w)
		return
	}

	err = s.db.Queries().ChatroomAddParticipant(r.Context(), db.ChatroomAddParticipantParams{
		ChatroomID:    uuid.NullUUID{UUID: room.ID, Valid: true},
		ParticipantID: uuid.NullUUID{UUID: frien.ID, Valid: true},
	})
	if err != nil {
		log.Printf("Err Creating room: %v", err)
		respondSimpleMessage("Internal Server Error.", 500, w)
		return
	}
	respondWithJson(room, 201, w)
}

func (s *Server) GetChatroomsHandler(w http.ResponseWriter, r *http.Request) {

	tokenString, _ := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, "IneedAnAppSecret")
	if err != nil {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	rooms, err := s.db.Queries().FindUsersChatrooms(r.Context(), uuid.NullUUID{UUID: userId, Valid: true})
	if err != nil {
		log.Printf("Err Geting users rooms: %v", err)
		respondSimpleMessage("Internal Server Error.", 500, w)
		return
	}

	respondWithJson(rooms, 200, w)
}

func (s *Server) LeaveChatroomHandler(w http.ResponseWriter, r *http.Request) {

	roomID, err := uuid.Parse(r.PathValue("chatroomID"))
	if err != nil {
		log.Printf("Room ID not set!")
		respondSimpleMessage("Bad Request", 400, w)
		return
	}

	tokenString, _ := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, "IneedAnAppSecret")
	if err != nil {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	err = s.db.Queries().ChatroomRemoveParticipant(r.Context(), db.ChatroomRemoveParticipantParams{
		ChatroomID:    uuid.NullUUID{UUID: roomID, Valid: true},
		ParticipantID: uuid.NullUUID{UUID: userId, Valid: true},
	})

	respondSimpleMessage("deleted", 204, w)
}

func (s *Server) CreateMessageHandler(w http.ResponseWriter, r *http.Request) {

	tokenString, _ := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, "IneedAnAppSecret")
	if err != nil {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	roomID, err := uuid.Parse(r.PathValue("chatroomID"))
	if err != nil {
		log.Printf("Room ID not set!")
		respondSimpleMessage("Bad Request", 400, w)
		return
	}

	type Parameters struct {
		Content string `json:"content"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	_ = decoder.Decode(&params)

	user, _ := s.db.Queries().FindUserById(r.Context(), userId)

	room, err := s.db.Queries().FindChatRoomById(r.Context(), roomID)
	if err != nil {
		log.Printf("Err Creating room: %v", err)
		respondSimpleMessage("Chatroom not found.", 404, w)
		return
	}

	msg, err := s.db.Queries().CreateMessage(r.Context(), db.CreateMessageParams{
		ID:         uuid.New(),
		Type:       "text",
		AuthorID:   uuid.NullUUID{UUID: user.ID, Valid: true},
		ChatroomID: uuid.NullUUID{UUID: room.ID, Valid: true},
		Content:    sql.NullString{String: params.Content, Valid: true},
	})

	respondWithJson(msg, 201, w)
}

func (s *Server) ReadMessagesHandler(w http.ResponseWriter, r *http.Request) {

	tokenString, _ := auth.GetBearerToken(r.Header)
	// userId
	_, err := auth.ValidateJWT(tokenString, "IneedAnAppSecret")
	if err != nil {
		log.Printf("JWT check Failed: %v", err)
		respondSimpleMessage("Unauthorized", 401, w)
		return
	}

	roomID, err := uuid.Parse(r.PathValue("chatroomID"))
	if err != nil {
		log.Printf("Room ID not set!")
		respondSimpleMessage("Bad Request", 400, w)
		return
	}

	type Parameters struct {
		Content string `json:"content"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	_ = decoder.Decode(&params)

	// user, _ := s.db.Queries().FindUserById(r.Context(), userId)

	room, err := s.db.Queries().FindChatRoomById(r.Context(), roomID)
	if err != nil {
		log.Printf("Err Creating room: %v", err)
		respondSimpleMessage("Chatroom not found.", 404, w)
		return
	}

	// TODO: check if user is participant of room
	msg, err := s.db.Queries().FindMessagesByRoomById(r.Context(), uuid.NullUUID{UUID: room.ID, Valid: true})

	respondWithJson(msg, 200, w)
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

func respondSimpleMessage(message string, statusCode int, w http.ResponseWriter) {
	resp := map[string]string{"message": message}
	respondWithJson(resp, statusCode, w)
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
