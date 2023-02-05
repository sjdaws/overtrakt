package trakt

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sjdaws/overtrakt/notify"
	"log"
	"net/http"
	"time"

	db "github.com/sjdaws/overtrakt/database"
)

type accessTokenRequest struct {
	Code         string `json:"code"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type accessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	CreatedAt    int    `json:"created_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type credentials struct {
	accessToken  string
	clientId     string
	clientSecret string
	deviceCode   string
	expiresAt    time.Time
	refreshToken string
	tokenType    string
}

type deviceCodeRequest struct {
	ClientId string `json:"client_id"`
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationUrl string `json:"verification_url"`
}

type refreshTokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	RedirectUri  string `json:"redirect_uri"`
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) authenticate() error {
	if c.credentials.accessToken == "" {
		traktCredentials, err := c.database.GetTraktAuth(c.credentials.clientId)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("auth: %v", err)
		}

		if traktCredentials != nil {
			c.credentials.accessToken = traktCredentials.AccessToken
			c.credentials.expiresAt = traktCredentials.ExpiresAt
			c.credentials.refreshToken = traktCredentials.RefreshToken
			c.credentials.tokenType = traktCredentials.TokenType
		}
	}

	if c.credentials.expiresAt.After(time.Now()) {
		return nil
	}

	if c.credentials.refreshToken != "" {
		log.Print("auth: trakt access token has expired, requesting refreshed token")
		response, err := c.refreshAccessToken(c.credentials.refreshToken)
		if err == nil {
			return c.saveAccessToken(response)
		}
	}

	response, err := c.createAccessToken()
	if err != nil {
		return fmt.Errorf("auth: %v", err)
	}

	return c.saveAccessToken(response)
}

func (c *Client) createAccessToken() (*accessTokenResponse, error) {
	codeResponse, err := c.getDeviceCode()
	if err != nil {
		return nil, err
	}

	c.credentials.deviceCode = codeResponse.DeviceCode
	expiresAt := time.Now().Add(time.Duration(codeResponse.ExpiresIn) * time.Second)

	notify.Message("Action required: authentication requires intervention.")
	log.Print("******************************** ACTION REQUIRED ********************************")
	log.Print("*                                                                               *")
	log.Printf("* Please go to %s and enter the following code: %s *", codeResponse.VerificationUrl, codeResponse.UserCode)
	log.Print("*                                                                               *")
	log.Printf("* Code will expire at %s                           *", expiresAt.Format(time.RFC822))
	log.Print("*                                                                               *")
	log.Print("*********************************************************************************")

	for {
		if time.Now().After(expiresAt) {
			return nil, fmt.Errorf("auth: unable to fetch trakt access token within allowed time limit")
		}

		// log.Printf("auth: waiting for authorisation for code: %s", codeResponse.UserCode)

		tokenResponse, err := c.getAccessToken()
		if err != nil {
			return nil, err
		}

		if len(tokenResponse.AccessToken) == 0 {
			time.Sleep(time.Duration(codeResponse.Interval) * time.Second)
			continue
		}

		log.Print("auth: authorised")

		return tokenResponse, nil
	}
}

func (c *Client) getAccessToken() (*accessTokenResponse, error) {
	httpResponse, err := c.doRequest(requestParameters{
		auth: false,
		body: accessTokenRequest{
			Code:         c.credentials.deviceCode,
			ClientId:     c.credentials.clientId,
			ClientSecret: c.credentials.clientSecret,
		},
		method: http.MethodPost,
		path:   "/oauth/device/token",
	})
	if err != nil {
		return nil, err
	}

	defer c.close(httpResponse.Body)

	var response accessTokenResponse
	if httpResponse.StatusCode == 200 {
		err = json.NewDecoder(httpResponse.Body).Decode(&response)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

func (c *Client) getDeviceCode() (*deviceCodeResponse, error) {
	httpResponse, err := c.doRequest(requestParameters{
		auth: false,
		body: deviceCodeRequest{
			ClientId: c.credentials.clientId,
		},
		method: http.MethodPost,
		path:   "/oauth/device/code",
	})
	if err != nil {
		return nil, err
	}

	defer c.close(httpResponse.Body)

	var response deviceCodeResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) refreshAccessToken(refreshToken string) (*accessTokenResponse, error) {
	httpResponse, err := c.doRequest(requestParameters{
		auth: false,
		body: refreshTokenRequest{
			ClientId:     c.credentials.clientId,
			ClientSecret: c.credentials.clientSecret,
			GrantType:    "refresh_token",
			RedirectUri:  "urn:ietf:wg:oauth:2.0:oob",
			RefreshToken: refreshToken,
		},
		method: http.MethodPost,
		path:   "/oauth/token",
	})
	if err != nil {
		return nil, err
	}

	defer c.close(httpResponse.Body)

	var response accessTokenResponse
	if httpResponse.StatusCode == 200 {
		err = json.NewDecoder(httpResponse.Body).Decode(&response)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

func (c *Client) saveAccessToken(response *accessTokenResponse) error {
	c.credentials.accessToken = response.AccessToken
	c.credentials.tokenType = response.TokenType

	return c.database.SetTraktAuth(&db.TraktCredentials{
		ClientId:     c.credentials.clientId,
		AccessToken:  c.credentials.accessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    response.TokenType,
		ExpiresAt:    time.Unix(int64(response.CreatedAt)+int64(response.ExpiresIn), 0),
	})
}
