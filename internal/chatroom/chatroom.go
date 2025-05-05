package chatroom

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fernandofreamunde/ika/internal/db"
	"github.com/fernandofreamunde/ika/internal/user"
	"github.com/google/uuid"
)

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
