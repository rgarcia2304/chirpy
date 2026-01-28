-- +goose Up
CREATE TABLE users(
	id UUID, 
	created_at TIMESTAMP NOT NULL, 
	updatad_at TIMESTAMP NOT NULL,
	email TEXT NOT NULL UNIQUE
); 

-- +goose Down
DROP TABLE users; 
