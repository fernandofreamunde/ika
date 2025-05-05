-- +goose Up
CREATE TABLE messages(
	id UUID PRIMARY KEY,
	sent_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	author_id UUID,
	chatroom_id UUID,
	type VARCHAR(256) NOT NULL,
	content TEXT DEFAULT NULL,
	CONSTRAINT fk_author_id FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_chatroom_id FOREIGN KEY (chatroom_id) REFERENCES chatrooms(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE messages;
