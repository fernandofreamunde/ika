package user

import (
	"context"
	"time"

	"github.com/fernandofreamunde/ika/internal/db"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Token        string    `json:"token"`
	// RefreshToken string    `json:"refresh_token"`
}

func CreateUser(email string, nick string, password string, ctx context.Context, createUser func(ctx context.Context, arg db.CreateUserParams) (db.User, error)) (User, error) {
	id, _ := uuid.NewUUID()

	data := db.CreateUserParams{
		ID:             id,
		Email:          email,
		Nickname:       nick,
		HashedPassword: password,
	}
	dbuser, err := createUser(ctx, data)

	if err != nil {
		return User{}, err
	}

	return User{
		ID:        dbuser.ID,
		Email:     dbuser.Email,
		Nickname:  dbuser.Nickname,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
	}, nil
}
