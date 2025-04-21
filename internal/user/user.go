package user

import (
	"context"
	"fmt"
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

func UpdateUser(dbUser db.User, email string, nick string, password string, ctx context.Context, dbq func() *db.Queries) (User, error) {

	if email == "" {
		email = dbUser.Email
	}

	if nick == "" {
		nick = dbUser.Nickname
	}

	if password == "" {
		password = dbUser.HashedPassword
	}

	data := db.UpdateUserParams{
		Email:          email,
		Nickname:       nick,
		HashedPassword: password,
		ID:             dbUser.ID,
	}

	updatedUser, err := dbq().UpdateUser(ctx, data)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		Nickname:  updatedUser.Nickname,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func CreateUser(email string, nick string, password string, ctx context.Context, dbq func() *db.Queries) (User, error) {
	id, _ := uuid.NewUUID()

	if email == "" || password == "" || nick == "" {
		return User{}, fmt.Errorf("email, password and nickname are mandatory fields!")
	}

	data := db.CreateUserParams{
		ID:             id,
		Email:          email,
		Nickname:       nick,
		HashedPassword: password,
	}

	_, err := dbq().FindUserByEmail(ctx, email)
	if err == nil {
		return User{}, fmt.Errorf("User with this email already exists!")
	}
	dbuser, err := dbq().CreateUser(ctx, data)

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
