-- +goose Up
CREATE TABLE refresh_tokens (
    token          CHAR(64)    PRIMARY KEY,
    created_at  TIMESTAMP NOT NULL,
    updated_at  TIMESTAMP NOT NULL,
    user_id       UUID NOT NULL UNIQUE,
    expires_at      TIMESTAMP NOT NULL,
    revoked_at      TIMESTAMP DEFAULT NULL,
    CONSTRAINT users_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE refresh_tokens;