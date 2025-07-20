package oidc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/strowk/foxy-contexts/pkg/auth"
	"github.com/strowk/foxy-contexts/pkg/streamable_http"
	"golang.org/x/oauth2"
)

// TODO: use cryptographic signatures to verify state instead of storing it in memory
// to avoid replay attacks while remaining scalable

var nonces = make(map[string]string)

type OIDCConfig struct {
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthStyle    string
}

func (c *OIDCConfig) toOauth2Config() *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       []string{"email", "openid", "groups", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   c.AuthURL,
			TokenURL:  c.TokenURL,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	switch c.AuthStyle {
	case "in_params":
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	case "in_header":
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	case "auto_detect":
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleAutoDetect
	default:
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInHeader // default to in_header
	}

	return cfg
}

func OidcEchoConfigurer(
	oidcCfg *OIDCConfig,
	path string,
) *streamable_http.EchoConfigurer {
	return &streamable_http.EchoConfigurer{
		Configure: func(e *echo.Echo) {
			e.GET(path, func(c echo.Context) error {
				r := c.Request()
				code := r.URL.Query().Get("code")

				if code != "" {
					state := r.URL.Query().Get("state")

					if state == "" {
						return errors.New("state is empty")
					}

					decodedState := make(map[string]string)
					err := json.Unmarshal([]byte(state), &decodedState)
					if err != nil {
						return err
					}
					nonce := decodedState["nonce"]
					authUrl := decodedState["auth_url"]
					if authUrl == "" {
						return errors.New("auth_url is empty")
					}
					if nonce == "" {
						return errors.New("nonce is empty")
					}
					if nonces[nonce] != authUrl {
						return errors.New("nonce is not valid")
					}
					delete(nonces, nonce)

					tok, err := oidcCfg.
						toOauth2Config().
						Exchange(r.Context(), code)
					if err != nil {
						return err
					}

					// tokens[tok] = struct{}{}

					// redirect back to finish the flow
					return c.Redirect(
						http.StatusFound,
						fmt.Sprintf(
							// TODO: make this configurable
							"http://localhost:8080%s&remote_token=%s",
							authUrl,
							tok.AccessToken,
						),
					)
				} else {
					c.Response().Writer.WriteHeader(http.StatusBadRequest)
				}
				return nil
			})
		},
	}
}

func AuthorizationHandler(oidcCfg *OIDCConfig) auth.AuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		// TODO: does it actually make sense to save remote token as user id?
		remoteToken := r.URL.Query().Get("remote_token")
		if remoteToken != "" {
			return remoteToken, nil
		}

		authUrl := r.URL.String()
		nonce := uuid.NewString()
		nonces[nonce] = authUrl

		state := struct {
			AuthUrl string `json:"auth_url"`
			Nonce   string `json:"nonce"`
		}{
			AuthUrl: authUrl,
			Nonce:   nonce,
		}

		marshalledState, err := json.Marshal(state)
		if err != nil {
			return "", err
		}

		url := oidcCfg.toOauth2Config().AuthCodeURL(string(marshalledState))

		http.Redirect(w, r, url, http.StatusFound)
		return "", nil // this will stop the request here as we already provided redirection
	}
}
