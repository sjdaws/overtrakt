package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"

	db "github.com/sjdaws/overtrakt/database"
	"github.com/sjdaws/overtrakt/notify"
	"github.com/sjdaws/overtrakt/trakt"
)

var (
	client   *trakt.Client
	database *db.Database

	databaseDbName    = os.Getenv("DATABASE_DBNAME")
	databaseHost      = os.Getenv("DATABASE_HOST")
	databasePassword  = os.Getenv("DATABASE_PASSWORD")
	databaseUsername  = os.Getenv("DATABASE_USERNAME")
	httpPort          = os.Getenv("HTTP_PORT")
	traktClientId     = os.Getenv("TRAKT_CLIENT_ID")
	traktClientSecret = os.Getenv("TRAKT_CLIENT_SECRET")
	traktMovieList    = os.Getenv("TRAKT_MOVIE_LIST")
	traktTvShowList   = os.Getenv("TRAKT_TVSHOW_LIST")
	traktUser         = os.Getenv("TRAKT_USER")
)

type media struct {
	ImdbId    string `json:"imdb_id"`
	MediaType string `json:"media_type"`
	TmdbId    string `json:"tmdb_id"`
	TvdbId    string `json:"tvdb_id"`
}

type webhookBody struct {
	Media    media  `json:"media"`
	Username string `json:"username"`
}

func init() {
	if httpPort == "" {
		httpPort = "8686"
	}

	if databaseDbName == "" {
		databaseDbName = "overtrakt"
	}

	if databaseHost == "" {
		databaseHost = "localhost"
	}

	if databaseUsername == "" {
		currentUser, err := user.Current()
		if err == nil {
			databaseUsername = currentUser.Username
		}
	}
}

func main() {
	if databaseUsername == "" || traktClientId == "" || traktClientSecret == "" || traktMovieList == "" || traktTvShowList == "" || traktUser == "" {
		missing := make([]string, 0)
		if databaseUsername == "" {
			missing = append(missing, "DATABASE_USERNAME")
		}
		if traktClientId == "" {
			missing = append(missing, "TRAKT_CLIENT_ID")
		}
		if traktClientSecret == "" {
			missing = append(missing, "TRAKT_CLIENT_SECRET")
		}
		if traktMovieList == "" {
			missing = append(missing, "TRAKT_MOVIE_LIST")
		}
		if traktTvShowList == "" {
			missing = append(missing, "TRAKT_TVSHOW_LIST")
		}
		if traktUser == "" {
			missing = append(missing, "TRAKT_USER")
		}

		log.Fatalf("One or more mandatory env values missing: %s", strings.Join(missing, ", "))
	}

	var err error
	database, err = db.Connect(databaseDbName, databaseHost, databasePassword, databaseUsername)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	client = trakt.NewClient(
		traktClientId,
		traktClientSecret,
		database,
	)

	args := os.Args[1:]

	if len(args) == 0 {
		http.HandleFunc("/health", health)
		http.HandleFunc("/webhook", webhook)
		log.Printf("Overtrakt: listening on port %s", httpPort)
		err = http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil)
		if err != nil {
			log.Fatalf("unable to start http server: %v", err)
		}
	}

	if args[0] == "unsynced" {
		unsynced()
	}
}

func health(response http.ResponseWriter, request *http.Request) {
	statusCode := 200

	clientHealth := client.Health()

	credentials, err := database.GetTraktAuth(traktClientId)
	databaseHealth := err == nil && credentials.AccessToken != ""

	if !clientHealth || !databaseHealth {
		statusCode = 500
	}

	response.WriteHeader(statusCode)

	if clientHealth && databaseHealth {
		_, err = response.Write([]byte(fmt.Sprintf("Hello %s, I'm OK\n\nDatabase: %v\nTrakt: %v", request.RemoteAddr, databaseHealth, clientHealth)))
	} else {
		_, err = response.Write([]byte(fmt.Sprintf("Hello %s, I'm not OK\n\nDatabase: %v\nTrakt: %v", request.RemoteAddr, databaseHealth, clientHealth)))
	}

	if err != nil {
		log.Printf("error serving health check: %v", err)
	}
}

func unsynced() {
	results, err := client.SyncUnsynced(traktMovieList, traktTvShowList, traktUser)
	if err != nil {
		log.Fatalf("unsynced: %v", err)
		return
	}

	for _, result := range results {
		var requestId string
		var requestType string
		if result.Request.TmdbId != "" {
			requestId = result.Request.TmdbId
			requestType = "tmdb"
		} else if result.Request.TvdbId != "" {
			requestId = result.Request.TvdbId
			requestType = "tvdb"
		} else {
			requestId = result.Request.ImdbId
			requestType = "imdb"
		}

		if result.Error != nil {
			log.Printf("Error adding %s using %s id %s, %v", result.Request.RequestType, requestType, requestId, err)
			continue
		}

		log.Printf("Successfully added %s using %s id %s", result.Request.RequestType, requestType, requestId)
	}
}

func webhook(response http.ResponseWriter, request *http.Request) {
	defer closeRequestBody(request.Body)

	var webhookRequest webhookBody
	err := json.NewDecoder(request.Body).Decode(&webhookRequest)
	if err != nil {
		log.Printf("webhook: %v", err)
		notify.Message(fmt.Sprintf("Error reading webhook body: %v", err))
		return
	}

	imdbId, _ := strconv.Atoi(webhookRequest.Media.ImdbId)
	if imdbId == 0 {
		webhookRequest.Media.ImdbId = ""
	}

	switch webhookRequest.Media.MediaType {
	case "movie":
		tmdbId, _ := strconv.Atoi(webhookRequest.Media.TmdbId)
		if tmdbId == 0 {
			webhookRequest.Media.TmdbId = ""
		}

		err = client.AddMovieToUserList(webhookRequest.Media.ImdbId, webhookRequest.Media.TmdbId, traktUser, traktMovieList)
		if err != nil {
			log.Printf("webhook: %v", err)
			response.WriteHeader(500)
			return
		}
		break

	case "tv":
		tvdbId, _ := strconv.Atoi(webhookRequest.Media.TvdbId)
		if tvdbId == 0 {
			webhookRequest.Media.TvdbId = ""
		}

		err = client.AddShowToUserList(webhookRequest.Media.ImdbId, webhookRequest.Media.TvdbId, traktUser, traktTvShowList)
		if err != nil {
			log.Printf("webhook: %v", err)
			response.WriteHeader(500)
			return
		}
		break
	}

	response.WriteHeader(201)
}

func closeRequestBody(body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		log.Printf("Error closing request body: %v", err)
	}
}
