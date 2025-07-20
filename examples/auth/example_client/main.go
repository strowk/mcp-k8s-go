package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/grokify/go-pkce"
	"github.com/labstack/echo/v4"
)

func main() {
	redirectUri := "http://localhost:8081/callback"

	body := `{
		"client_name": "test-client-name",
		"redirect_uris": ["%s"],
		"token_endpoint_auth_method": "client_secret_post",
		"grant_types": ["authorization_code"],
		"response_types": ["code"]
	}`
	body = fmt.Sprintf(body, redirectUri)

	reqBody := []byte(body)

	req, err := http.NewRequest("POST", "http://localhost:8080/register", bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		panic(err)
	}
	bodyString := respBody.String()

	if resp.StatusCode != http.StatusOK {
		panic("Error: " + http.StatusText(resp.StatusCode) + " " + bodyString)
	} else {
		println("Response: " + bodyString)
	}

	var registeredClient struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}

	err = json.Unmarshal(respBody.Bytes(), &registeredClient)
	if err != nil {
		panic(err)
	}

	codeVerifier, err := pkce.NewCodeVerifier(-1)
	if err != nil {
		panic(err)
	}
	codeChallenge := pkce.CodeChallengeS256(codeVerifier)

	resultingUrl := fmt.Sprintf(`http://localhost:8080/authorize?client_id=%s&response_type=code&code_challenge=%s&redirect_uri=%s&code_challenge_method=S256`, registeredClient.ClientId, codeChallenge, redirectUri)
	println("Go to URL: " + resultingUrl)

	e := echo.New()
	e.GET("/callback", func(c echo.Context) error {
		code := c.QueryParam("code")
		if code == "" {
			return c.String(http.StatusBadRequest, "code is empty")
		}
		tokenResponse, err := http.PostForm("http://localhost:8080/token", url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {redirectUri},
			"code_verifier": {codeVerifier},
			"client_id":     {registeredClient.ClientId},
			"client_secret": {registeredClient.ClientSecret},
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error sending request: "+err.Error())
		}

		defer tokenResponse.Body.Close()
		tokenRespBody := new(bytes.Buffer)
		_, err = tokenRespBody.ReadFrom(tokenResponse.Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading response: "+err.Error())
		}
		tokenBody := tokenRespBody.String()
		if tokenResponse.StatusCode != http.StatusOK {
			return c.String(http.StatusInternalServerError, "Error: "+http.StatusText(tokenResponse.StatusCode)+" "+tokenBody)
		} else {
			println("Token Response: " + tokenBody)

			parsedToken := new(struct {
				AccessToken  string `json:"access_token"`
				TokenType    string `json:"token_type"`
				ExpiresIn    int    `json:"expires_in"`
				RefreshToken string `json:"refresh_token"`
			})

			err = json.Unmarshal(tokenRespBody.Bytes(), parsedToken)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error unmarshalling response: "+err.Error())
			}

			// requesting MCP ping with the access token
			pingReq, err := http.NewRequest(
				"POST",
				"http://localhost:8080/mcp",
				bytes.NewBuffer([]byte(`{"method":"ping","params":{},"id":0, "jsonrpc":"2.0"}`)),
			)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error creating request: "+err.Error())
			}
			pingReq.Header.Set("Authorization", "Bearer "+parsedToken.AccessToken)
			pingReq.Header.Set("Content-Type", "application/json")
			pingReq.Header.Set("Accept", "application/json, text/event-stream")

			pingResp, err := client.Do(pingReq)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error sending request: "+err.Error())
			}
			defer pingResp.Body.Close()
			pingRespBody := new(bytes.Buffer)
			_, err = pingRespBody.ReadFrom(pingResp.Body)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error reading response: "+err.Error())
			}
			pingBodyString := pingRespBody.String()
			if pingResp.StatusCode != http.StatusOK {
				return c.String(http.StatusInternalServerError, "Error: "+http.StatusText(pingResp.StatusCode)+" "+pingBodyString)
			} else {
				println("MCP Ping Response: " + pingBodyString)
			}

			// try same request without authorization header
			// to check that authorization is required

			ping2Req, err := http.NewRequest(
				"POST",
				"http://localhost:8080/mcp",
				bytes.NewBuffer([]byte(`{"method":"ping","params":{},"id":1, "jsonrpc":"2.0"}`)),
			)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error creating request: "+err.Error())
			}
			ping2Req.Header.Set("Content-Type", "application/json")
			ping2Req.Header.Set("Accept", "application/json, text/event-stream")
			ping2Resp, err := client.Do(ping2Req)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error sending request: "+err.Error())
			}
			defer ping2Resp.Body.Close()
			if ping2Resp.StatusCode != http.StatusUnauthorized {
				println("Expected 401 Unauthorized, got: " + http.StatusText(ping2Resp.StatusCode))
			} else {
				println("MCP Ping Response without auth: " + http.StatusText(ping2Resp.StatusCode))
			}

			// finally try to call a tool that would list pods in kube-sytem namespace
			listPodsReq, err := http.NewRequest(
				"POST",
				"http://localhost:8080/mcp",
				bytes.NewBuffer([]byte(
					`{"method":"tools/call","params":{"name":"list-k8s-resources","arguments":{"context":"k3d-mcp-test-auth","namespace":"kube-system","kind":"pod"}},"id":2, "jsonrpc":"2.0"}`,
				)),
			)
			listPodsReq.Header.Set("Authorization", "Bearer "+parsedToken.AccessToken)
			listPodsReq.Header.Set("Content-Type", "application/json")
			listPodsReq.Header.Set("Accept", "application/json, text/event-stream")

			listPodsResp, err := client.Do(listPodsReq)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error sending request: "+err.Error())
			}
			defer listPodsResp.Body.Close()
			listPodsRespBody := new(bytes.Buffer)
			_, err = listPodsRespBody.ReadFrom(listPodsResp.Body)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error reading response: "+err.Error())
			}
			listPodsBodyString := listPodsRespBody.String()
			if listPodsResp.StatusCode != http.StatusOK {
				return c.String(http.StatusInternalServerError, "Error: "+http.StatusText(listPodsResp.StatusCode)+" "+listPodsBodyString)
			} else {
				println("MCP Tool Call Response: " + listPodsBodyString)
			}

		}
		return c.String(http.StatusOK, "OK")
	})

	e.Start(":8081")
}
