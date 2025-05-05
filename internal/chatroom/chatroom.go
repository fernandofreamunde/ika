package chatroom

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fernandofreamunde/ika/internal/db"
	"github.com/fernandofreamunde/ika/internal/user"
	"github.com/google/uuid"
)

type SendMessageParams struct {
	AuthorID uuid.UUID
	ChatroomID uuid.UUID
	Content string
}

func CreateChatRoomWithParticipants(p1, p2 user.User, ctx context.Context, dbq func() *db.Queries) (db.Chatroom, error) {
	
	room, err := dbq().CreateChatroom(ctx, db.CreateChatroomParams{
		ID:   uuid.New(),
		Name: sql.NullString{String: fmt.Sprintf("%s:%s", p1.Nickname, p2.Nickname), Valid: true},
		Type: "direct",
	})

	if err != nil {
		return db.Chatroom{}, fmt.Errorf("Err Creating room: %v", err)
	}

	err = dbq().ChatroomAddParticipant(ctx, db.ChatroomAddParticipantParams{
		ChatroomID:    uuid.NullUUID{UUID: room.ID, Valid: true},
		ParticipantID: uuid.NullUUID{UUID: p1.ID, Valid: true},
	})
	if err != nil {
		return db.Chatroom{}, fmt.Errorf("Err Adding participant to room: %v", err)
	}

	err = dbq().ChatroomAddParticipant(ctx, db.ChatroomAddParticipantParams{
		ChatroomID:    uuid.NullUUID{UUID: room.ID, Valid: true},
		ParticipantID: uuid.NullUUID{UUID: p2.ID, Valid: true},
	})

	if err != nil {
		return db.Chatroom{}, fmt.Errorf("Err Adding participant to room: %v", err)
	}

	return room, nil
}

func IsUserParticipantInChatroom(userId uuid.UUID, chatroomId uuid.UUID, ctx context.Context, dbq func() *db.Queries) (bool, error) {
	
	room, err := dbq().FindChatRoomById(ctx, chatroomId)
	if err != nil {
		return false, err
	}
	participants, _ := dbq().FindParticipantIdsByChatRoomId(ctx, uuid.NullUUID{UUID: room.ID, Valid: true})

	in := false
	for _, p := range participants {
		if p.ParticipantID.UUID.String() == userId.String() {
			in = true
		}
	}
	return in, nil
}

func SendMessageInChatroom(params SendMessageParams, ctx context.Context, dbq func() *db.Queries) (db.Message, error) {

	msg, err := dbq().CreateMessage(ctx, db.CreateMessageParams{
		ID:         uuid.New(),
		Type:       "text",
		AuthorID:   uuid.NullUUID{UUID: params.AuthorID, Valid: true},
		ChatroomID: uuid.NullUUID{UUID: params.ChatroomID, Valid: true},
		Content:    sql.NullString{String: params.Content, Valid: true},
	})

	if err != nil {
		return db.Message{}, fmt.Errorf("Err creating message: %v", err)
	}

	return msg, nil
}

