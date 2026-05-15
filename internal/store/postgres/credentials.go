package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
)

func (r *Repo) ListCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := r.pool.Query(ctx, `SELECT data FROM passkey_credentials WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("ListCredentials: %w", err)
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("ListCredentials scan: %w", err)
		}
		var c webauthn.Credential
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			return nil, fmt.Errorf("ListCredentials unmarshal: %w", err)
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

func (r *Repo) SaveCredential(ctx context.Context, userID string, cred *webauthn.Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("SaveCredential marshal: %w", err)
	}
	id := base64.URLEncoding.EncodeToString(cred.ID)
	_, err = r.pool.Exec(ctx,
		`INSERT INTO passkey_credentials (id, data, user_id) VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET data = $2, user_id = $3`,
		id, string(data), userID,
	)
	return err
}

func (r *Repo) DeleteAllCredentials(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM passkey_credentials`)
	if err != nil {
		return 0, fmt.Errorf("DeleteAllCredentials: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) UpdateCredential(ctx context.Context, cred *webauthn.Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("UpdateCredential marshal: %w", err)
	}
	id := base64.URLEncoding.EncodeToString(cred.ID)
	_, err = r.pool.Exec(ctx,
		`UPDATE passkey_credentials SET data = $2 WHERE id = $1`,
		id, string(data),
	)
	return err
}
