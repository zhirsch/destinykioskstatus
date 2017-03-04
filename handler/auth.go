package handler

import (
	"errors"
	"log"
	"net/http"

	"golang.org/x/net/context"
	"github.com/zhirsch/oauth2"

	"github.com/zhirsch/destinykioskstatus/db"
	"github.com/zhirsch/destinykioskstatus/server"
)

const userCookieName = "X-DestinyKioskStatus-User"

var errNeedAuth = errors.New("ErrNeedAuth")

type AuthenticationMiddlewareHandler struct {
	Server     *server.Server
	AuthConfig *oauth2.Config
	Handler    Handler
}

func (h AuthenticationMiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := h.getUser(w, r)
	if err != nil {
		u := *r.URL
		u.Scheme = "https"
		u.Host = r.Host
		http.Redirect(w, r, h.AuthConfig.AuthCodeURL(u.String()), http.StatusSeeOther)
		return
	}
	h.Handler.ServeHTTP(u, w, r)
}

func (h AuthenticationMiddlewareHandler) getUser(w http.ResponseWriter, r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie(userCookieName)
	if err != nil {
		return nil, err
	}
	return h.Server.DB.SelectUser(cookie.Value)
}

type BungieAuthCallbackHandler struct {
	Server     *server.Server
	AuthConfig *oauth2.Config
}

func (h BungieAuthCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate the incoming query.
	q := r.URL.Query()
	token, err := h.AuthConfig.Exchange(context.TODO(), q["code"][0])
	if err != nil {
		panic(err)
	}

	// Get the user info.
	userResp := h.Server.API.GetBungieNetUser(token)

	// Create and insert the user into the database.
	user := &db.User{
		ID:    userResp.Response.User.MembershipID,
		Name:  userResp.Response.User.DisplayName,
		Token: token,
	}
	if err := h.Server.DB.InsertUser(user); err != nil {
		log.Printf("unable to write user to db: %v", err)
	}

	// Set the cookie and redirecto to the original URL.
	http.SetCookie(w, &http.Cookie{Name: "X-DestinyKioskStatus-User", Value: user.ID})
	http.Redirect(w, r, q["state"][0], http.StatusSeeOther)
}
