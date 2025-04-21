package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/fernandofreamunde/ika/internal/auth"
	"github.com/fernandofreamunde/ika/internal/user"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	//mux.HandleFunc("GET /", s.HelloWorldHandler)
	mux.HandleFunc("POST /api/users", s.RegisterUserHandler)
	mux.HandleFunc("POST /api/login", s.LoginHandler)
	mux.HandleFunc("POST /api/new_message", s.NewMessageHandler)
	mux.HandleFunc("POST /api/chatrooms", s.NewChatRoomHandler)

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

	expiresIn := 60 * 60
	jwt, err := auth.MakeJWT(dbUser.ID, "IneedAnAppSecret", time.Duration(expiresIn)*time.Second)
	//refreshToken, _ := auth.MakeRefreshToken()

	//_, err = s..queries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
	//	Token:     refreshToken,
	//	UpdatedAt: time.Now(),
	//	CreatedAt: time.Now(),
	//	ExpiresAt: time.Now().Add(time.Duration(60 * 24 * time.Hour)),
	//	UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	//})

	type loginResponse struct {
		User         user.User `json:"user"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	resp := loginResponse{
		User: user.User{
			ID:        dbUser.ID,
			Email:     dbUser.Email,
			Nickname:  dbUser.Nickname,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		},
		Token:        jwt,
		RefreshToken: "asdfasdf",
	}
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
