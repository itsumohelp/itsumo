-- +goose Up
CREATE TABLE passkey_credentials (
    id         TEXT        NOT NULL PRIMARY KEY,  -- base64url encoded credential ID
    data       TEXT        NOT NULL,              -- JSON serialized webauthn.Credential
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS passkey_credentials;
