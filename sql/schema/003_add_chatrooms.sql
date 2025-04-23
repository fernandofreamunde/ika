-- +goose Up
CREATE TABLE chatrooms(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	type VARCHAR(256) NOT NULL,
	name VARCHAR(256) UNIQUE DEFAULT NULL
);

CREATE TABLE chatrooms_participants(
	chatroom_id UUID,
	participant_id UUID,
	CONSTRAINT fk_participant_id FOREIGN KEY (participant_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_chatroom_id FOREIGN KEY (chatroom_id) REFERENCES chatrooms(id) ON DELETE CASCADE,
	CONSTRAINT unique_participants UNIQUE (chatroom_id, participant_id)
);

-- +goose Down
DROP TABLE chatrooms_participants;
DROP TABLE chatrooms;
