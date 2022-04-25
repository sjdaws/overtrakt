package trakt

import (
	"encoding/json"
	"fmt"
	db "github.com/sjdaws/overtrakt/database"
	"github.com/sjdaws/overtrakt/notify"
	"log"
	"strings"
)

type addUserListRequest struct {
	Movies []movieIds `json:"movies"`
	Shows  []showIds  `json:"shows"`
}

type addUserListResponse struct {
	Added    addResult          `json:"added"`
	Existing addResult          `json:"existing"`
	NotFound addUserListRequest `json:"not_found"`
}

type addResult struct {
	Movies int `json:"movies"`
	Shows  int `json:"shows"`
}

type movieId struct {
	ImdbId string `json:"imdb"`
	TmdbId string `json:"tmdb"`
}

type movieIds struct {
	Ids movieId `json:"ids"`
}

type showId struct {
	ImdbId string `json:"imdb"`
	TvdbId string `json:"tvdb"`
}

type showIds struct {
	Ids showId `json:"ids"`
}

func (c *Client) AddMovieToUserList(imdbId string, tmdbId string, userId string, userListId string) error {
	if imdbId == "" && tmdbId == "" {
		return fmt.Errorf("user_list: unable to add movie to trakt, no ids are supplied")
	}

	request := &db.TraktRequest{
		ImdbId:      imdbId,
		RequestType: "movie",
		TmdbId:      tmdbId,
		TvdbId:      "",
	}

	// Don't die on db error, we can continue anyway
	err := c.database.AddTraktRequest(request)
	if err != nil {
		log.Printf("user_list: error adding movie request to database: %v", err)
	}

	var ids movieId
	if tmdbId != "" {
		ids = movieId{
			TmdbId: tmdbId,
		}
	} else {
		ids = movieId{
			ImdbId: imdbId,
		}
	}

	httpResponse, err := c.queryApi(requestParameters{
		body: addUserListRequest{
			Movies: []movieIds{
				{
					Ids: ids,
				},
			},
		},
		path: fmt.Sprintf("/users/%s/lists/%s/items", userId, userListId),
	})
	if err != nil {
		return fmt.Errorf("user_list: %v", err)
	}

	defer c.close(httpResponse.Body)

	var response addUserListResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("user_list: %v", err)
	}

	errors := make([]string, 0)
	if len(response.NotFound.Movies) > 0 {
		for _, id := range response.NotFound.Movies {
			if id.Ids.TmdbId != "" {
				errors = append(errors, fmt.Sprintf("tmdb: %s", id.Ids.TmdbId))
			}
			if id.Ids.ImdbId != "" {
				errors = append(errors, fmt.Sprintf("imdb: %s", id.Ids.ImdbId))
			}
		}
	}

	success := response.Added.Movies + response.Existing.Movies
	total := success + len(errors)

	message := ""

	if len(errors) > 0 {
		message = fmt.Sprintf("Error adding %d/%d movie(s) to trakt: %s\n", len(errors), total, strings.Join(errors, ","))
	}
	if success > 0 {
		message = fmt.Sprintf("Successfully added %d/%d movie(s) to trakt\n", success, total)
	}

	log.Printf("user_list: %s", message)
	notify.Message(message)

	if success > 0 {
		request.Added = true
		err = c.database.UpdateTraktRequest(request)
		if err != nil {
			log.Printf("user_list: error updating movie request in database: %v", err)
		}
	}

	return nil
}

func (c *Client) AddShowToUserList(imdbId string, tvdbId string, userId string, userListId string) error {
	if imdbId == "" && tvdbId == "" {
		return fmt.Errorf("user_list: unable to add tv show to trakt, no ids are supplied")
	}

	request := &db.TraktRequest{
		ImdbId:      imdbId,
		RequestType: "show",
		TmdbId:      "",
		TvdbId:      tvdbId,
	}

	// Don't die on db error, we can continue anyway
	err := c.database.AddTraktRequest(request)
	if err != nil {
		log.Printf("user_list: error adding tv show request to database: %v", err)
	}

	var ids showId
	if tvdbId != "" {
		ids = showId{
			TvdbId: tvdbId,
		}
	} else {
		ids = showId{
			ImdbId: imdbId,
		}
	}

	httpResponse, err := c.queryApi(requestParameters{
		body: addUserListRequest{
			Shows: []showIds{
				{
					Ids: ids,
				},
			},
		},
		path: fmt.Sprintf("/users/%s/lists/%s/items", userId, userListId),
	})
	if err != nil {
		return fmt.Errorf("user_list: %v", err)
	}

	defer c.close(httpResponse.Body)

	var response addUserListResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("user_list: %v", err)
	}

	errors := make([]string, 0)
	if len(response.NotFound.Shows) > 0 {
		for _, id := range response.NotFound.Shows {
			if id.Ids.TvdbId != "" {
				errors = append(errors, fmt.Sprintf("tvdb: %s", id.Ids.TvdbId))
			}
			if id.Ids.ImdbId != "" {
				errors = append(errors, fmt.Sprintf("imdb: %s", id.Ids.ImdbId))
			}
		}
	}

	success := response.Added.Shows + response.Existing.Shows
	total := success + len(errors)

	message := ""

	if len(errors) > 0 {
		message = fmt.Sprintf("Error adding %d/%d tv show(s) to trakt: %s\n", len(errors), total, strings.Join(errors, ","))
	}
	if success > 0 {
		message = fmt.Sprintf("Successfully added %d/%d tv show(s) to trakt\n", success, total)
	}

	log.Printf("user_list: %s", message)
	notify.Message(message)

	if success > 0 {
		request.Added = true
		err = c.database.UpdateTraktRequest(request)
		if err != nil {
			log.Printf("user_list: error updating tv show request in database: %v", err)
		}
	}

	return nil
}
