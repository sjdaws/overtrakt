package database

import (
	"time"
)

type TraktCredentials struct {
	ClientId     string
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
}

func (d *Database) GetTraktAuth(clientId string) (*TraktCredentials, error) {
	prepared, err := d.connection.Prepare("SELECT client_id, access_token, expires_at, refresh_token, token_type FROM trakt_credentials WHERE client_id = ?")
	if err != nil {
		return nil, err
	}

	stmt := &statement{
		prepared: prepared,
	}
	defer stmt.close()

	credentials := &TraktCredentials{}
	err = stmt.prepared.QueryRow(clientId).Scan(&credentials.ClientId, &credentials.AccessToken, &credentials.ExpiresAt, &credentials.RefreshToken, &credentials.TokenType)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

func (d *Database) SetTraktAuth(credentials *TraktCredentials) error {
	prepared, err := d.connection.Prepare("INSERT INTO trakt_credentials VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE access_token = ?, refresh_token = ?, token_type = ?, expires_at = ?")
	if err != nil {
		return err
	}

	stmt := &statement{
		prepared: prepared,
	}
	defer stmt.close()

	_, err = stmt.prepared.Exec(
		credentials.ClientId,
		credentials.AccessToken,
		credentials.RefreshToken,
		credentials.TokenType,
		credentials.ExpiresAt,
		credentials.AccessToken,
		credentials.RefreshToken,
		credentials.TokenType,
		credentials.ExpiresAt,
	)

	return err
}
