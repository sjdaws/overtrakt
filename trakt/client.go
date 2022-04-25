package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	db "github.com/sjdaws/overtrakt/database"
)

type Client struct {
	credentials credentials
	database    *db.Database
	httpClient  *http.Client
}

type requestParameters struct {
	auth   bool
	body   interface{}
	method string
	path   string
}

const apiUrl = "https://api.trakt.tv"

func NewClient(clientId string, clientSecret string, database *db.Database) *Client {
	client := &Client{
		credentials: credentials{
			clientId:     clientId,
			clientSecret: clientSecret,
		},
		database: database,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout: 5 * time.Second,
			},
		},
	}

	err := client.authenticate()
	if err != nil {
		log.Print(err)
	}

	return client
}

func (c *Client) Health() bool {
	return c.credentials.accessToken != ""
}

func (c *Client) doRequest(parameters requestParameters) (*http.Response, error) {
	request, err := http.NewRequest(parameters.method, fmt.Sprintf("%s%s", apiUrl, parameters.path), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "application/json")

	if parameters.body != nil {
		request.Header.Add("Content-Type", "application/json")
		body, err := json.Marshal(parameters.body)
		if err != nil {
			return nil, err
		}
		request.Body = io.NopCloser(bytes.NewReader(body))
	}

	request.URL.RawQuery = request.URL.Query().Encode()

	if parameters.auth {
		request.Header.Add("Authorization", fmt.Sprintf("%s %s", c.credentials.tokenType, c.credentials.accessToken))
		request.Header.Add("trakt-api-key", c.credentials.clientId)
		request.Header.Add("trakt-api-version", "2")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) close(body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		log.Printf("Unable to close http response body: %v", err)
	}
}

func (c *Client) queryApi(parameters requestParameters) (*http.Response, error) {
	err := c.authenticate()
	if err != nil {
		return nil, err
	}

	parameters.auth = true
	if parameters.method == "" {
		parameters.method = http.MethodPost
	}

	return c.doRequest(parameters)
}
