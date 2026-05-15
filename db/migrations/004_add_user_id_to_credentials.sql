-- +goose Up
ALTER TABLE passkey_credentials ADD COLUMN user_id TEXT NOT NULL DEFAULT 'owner';
CREATE INDEX idx_passkey_credentials_user_id ON passkey_credentials (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_passkey_credentials_user_id;
ALTER TABLE passkey_credentials DROP COLUMN user_id;
