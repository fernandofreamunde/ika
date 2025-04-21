-- +goose Up
CREATE TABLE users (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	hashed_password TEXT NOT NULL,
	nickname VARCHAR(256) NOT NULL,
	email VARCHAR(256) UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;
