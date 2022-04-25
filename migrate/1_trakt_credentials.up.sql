CREATE TABLE trakt_credentials (
    client_id varchar(64) NOT NULL,
    access_token varchar(64) NOT NULL,
    refresh_token varchar(64) NOT NULL,
    token_type varchar(10) NOT NULL,
    expires_at datetime NOT NULL,
    PRIMARY KEY (client_id)
);