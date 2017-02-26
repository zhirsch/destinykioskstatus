package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Token struct {
	Value   string
	Ready   time.Time
	Expires time.Time
}

func (t Token) IsReady() bool {
	return time.Now().After(t.Ready)
}

func (t Token) IsExpired() bool {
	return time.Now().After(t.Expires)
}

func (c *Client) Authenticate(w http.ResponseWriter, r *http.Request) bool {
	if !c.AuthToken.IsExpired() {
		return true
	}

	u := r.URL
	u.Scheme = "https"
	u.Host = r.Host

	var au url.URL = *c.authURL
	q := au.Query()
	q.Set("state", u.String())
	au.RawQuery = q.Encode()

	http.Redirect(w, r, au.String(), http.StatusSeeOther)
	return false
}

func (c *Client) HandleBungieAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Validate the incoming query.
	query := r.URL.Query()
	if _, ok := query["code"]; !ok {
		panic(fmt.Sprintf("no 'code' in request: %s", r.URL))
	} else if len(query["code"]) != 1 {
		panic(fmt.Sprintf("multiple 'code' in request: %s", r.URL))
	} else if _, ok := query["state"]; !ok {
		panic(fmt.Sprintf("no `state` in request: %s", r.URL))
	} else if len(query["state"]) != 1 {
		panic(fmt.Sprintf("too many `state` in request: %s", r.URL))
	}

	// Prepare the request body.
	req := GetAccessTokensFromCodeRequest{Code: query["code"][0]}
	body, err := encode(req)
	if err != nil {
		panic(err)
	}

	// Send the HTTP request.
	httpReq, err := http.NewRequest("POST", req.URL(), body)
	if err != nil {
		panic(err)
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("X-API-Key", c.apiKey)
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		panic(err)
	}
	if httpResp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("bad response: %v", httpResp.StatusCode))
	}

	// Parse the HTTP response.
	var resp GetAccessTokensFromCodeResponse
	if err := decode(&resp, httpResp.Body); err != nil {
		panic(err)
	}
	if resp.ErrorCode != 1 {
		panic(fmt.Sprintf("bad message: %+v", resp))
	}

	// Create the tokens.
	now := time.Now()
	c.AuthToken = &Token{
		Value:   resp.Response.AccessToken.Value,
		Ready:   now.Add(time.Duration(resp.Response.AccessToken.ReadyIn) * time.Second),
		Expires: now.Add(time.Duration(resp.Response.AccessToken.Expires) * time.Second),
	}
	c.RefreshToken = &Token{
		Value:   resp.Response.RefreshToken.Value,
		Ready:   now.Add(time.Duration(resp.Response.RefreshToken.ReadyIn) * time.Second),
		Expires: now.Add(time.Duration(resp.Response.RefreshToken.Expires) * time.Second),
	}

	// Redirect to the original URL.
	http.Redirect(w, r, query["state"][0], http.StatusSeeOther)
}
