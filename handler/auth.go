package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/zhirsch/destinykioskstatus/db"
	"github.com/zhirsch/destinykioskstatus/server"
)

const userCookieName = "X-DestinyKioskStatus-User"

var errNeedAuth = errors.New("ErrNeedAuth")

type AuthenticationMiddlewareHandler struct {
	Server  *server.Server
	Handler Handler

	AuthURL *url.URL
}

func (h AuthenticationMiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := h.getUser(w, r)
	if err != nil {
		u := r.URL
		u.Scheme = "https"
		u.Host = r.Host
		http.Redirect(w, r, h.getAuthURL(u), http.StatusSeeOther)
		return
	}
	h.Handler.ServeHTTP(u, w, r)
}

func (h AuthenticationMiddlewareHandler) getAuthURL(u *url.URL) string {
	var au url.URL = *h.AuthURL
	q := au.Query()
	q.Set("state", u.String())
	au.RawQuery = q.Encode()
	return au.String()
}

func (h AuthenticationMiddlewareHandler) getUser(w http.ResponseWriter, r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie(userCookieName)
	if err != nil {
		return nil, err
	}
	return h.Server.DB.SelectUser(cookie.Value)
}

type BungieAuthCallbackHandler struct {
	Server *server.Server
}

func (h BungieAuthCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	resp := h.Server.API.GetAccessTokensFromCode(query["code"][0])

	// Create the tokens.
	now := time.Now()
	authToken := db.Token{
		Value:   resp.Response.AccessToken.Value,
		Ready:   now.Add(time.Duration(resp.Response.AccessToken.ReadyIn) * time.Second),
		Expires: now.Add(time.Duration(resp.Response.AccessToken.Expires) * time.Second),
	}
	refreshToken := db.Token{
		Value:   resp.Response.RefreshToken.Value,
		Ready:   now.Add(time.Duration(resp.Response.RefreshToken.ReadyIn) * time.Second),
		Expires: now.Add(time.Duration(resp.Response.RefreshToken.Expires) * time.Second),
	}

	// Get the user info.
	userResp := h.Server.API.GetBungieNetUser(authToken.Value)

	// Create and insert the user into the database.
	user := &db.User{
		ID:           userResp.Response.User.MembershipID,
		Name:         userResp.Response.User.DisplayName,
		AuthToken:    authToken,
		RefreshToken: refreshToken,
	}
	if err := h.Server.DB.InsertUser(user); err != nil {
		log.Printf("unable to write bungie auth to db: %v", err)
	}

	// Set the cookie and redirecto to the original URL.
	http.SetCookie(w, &http.Cookie{Name: "X-DestinyKioskStatus-User", Value: user.ID})
	http.Redirect(w, r, query["state"][0], http.StatusSeeOther)
}
