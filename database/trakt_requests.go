package database

import (
	"time"
)

type TraktRequest struct {
	ImdbId      string
	RequestType string
	TmdbId      string
	TvdbId      string
	Added       bool
	CreatedAt   time.Time
}

func (d *Database) AddTraktRequest(request *TraktRequest) error {
	prepared, err := d.connection.Prepare("INSERT INTO trakt_requests VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE imdb_id = ?, request_type = ?, tmdb_id = ?, tvdb_id = ?")
	if err != nil {
		return err
	}

	stmt := &statement{
		prepared: prepared,
	}
	defer stmt.close()

	_, err = stmt.prepared.Exec(
		request.ImdbId,
		request.RequestType,
		request.TmdbId,
		request.TvdbId,
		false,
		nil,
		request.ImdbId,
		request.RequestType,
		request.TmdbId,
		request.TvdbId,
	)

	return err
}

func (d *Database) UpdateTraktRequest(request *TraktRequest) error {
	prepared, err := d.connection.Prepare("INSERT INTO trakt_requests VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE imdb_id = ?, request_type = ?, tmdb_id = ?, tvdb_id = ?, added = ?")
	if err != nil {
		return err
	}

	stmt := &statement{
		prepared: prepared,
	}
	defer stmt.close()

	_, err = stmt.prepared.Exec(
		request.ImdbId,
		request.RequestType,
		request.TmdbId,
		request.TvdbId,
		request.Added,
		nil,
		request.ImdbId,
		request.RequestType,
		request.TmdbId,
		request.TvdbId,
		request.Added,
	)

	return err
}
