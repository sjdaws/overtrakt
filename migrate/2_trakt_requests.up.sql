CREATE TABLE trakt_requests (
    imdb_id varchar(20) NOT NULL,
    request_type varchar(10) NOT NULL,
    tmdb_id varchar(20) NOT NULL,
    tvdb_id varchar(20) NOT NULL,
    added BOOL NOT NULL DEFAULT FALSE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (imdb_id, request_type, tmdb_id, tvdb_id)
);