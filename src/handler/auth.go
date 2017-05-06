package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/zhirsch/oauth2"
	"golang.org/x/net/context"

	"github.com/zhirsch/destinykioskstatus/src/db"
	"github.com/zhirsch/destinykioskstatus/src/server"
)

const (
	cookieBungieMembershipID = "X-DestinyKioskStatus-BungieMembershipID"
)

var errNeedAuth = errors.New("ErrNeedAuth")

type AuthenticationMiddlewareHandler struct {
	Server     *server.Server
	AuthConfig *oauth2.Config
	Handler    Handler
}

func (h AuthenticationMiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bungieUser, err := h.getBungieUser(w, r)
	if err != nil {
		log.Printf("%v", err)
		u := *r.URL
		u.Scheme = "https"
		u.Host = r.Host
		http.Redirect(w, r, h.AuthConfig.AuthCodeURL(u.String()), http.StatusSeeOther)
		return
	}
	h.Handler.ServeHTTP(bungieUser, w, r)
}

func (h AuthenticationMiddlewareHandler) getBungieUser(w http.ResponseWriter, r *http.Request) (*db.BungieUser, error) {
	cookie, err := r.Cookie(cookieBungieMembershipID)
	if err != nil {
		return nil, err
	}
	return h.Server.DB.SelectBungieUser(db.BungieMembershipID(cookie.Value))
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

	// Get the account info.
	bungieAccountResp := h.Server.API.GetCurrentBungieAccount(token)

	bungieUser := &db.BungieUser{
		MembershipID: db.BungieMembershipID(bungieAccountResp.Response.BungieNetUser.MembershipID),
		DisplayName:  bungieAccountResp.Response.BungieNetUser.DisplayName,
		Token:        token,
	}
	for _, destinyAccountResp := range bungieAccountResp.Response.DestinyAccounts {
		destinyUser := &db.DestinyUser{
			MembershipType: db.DestinyMembershipType(destinyAccountResp.UserInfo.MembershipType),
			MembershipID:   db.DestinyMembershipID(destinyAccountResp.UserInfo.MembershipID),
			DisplayName:    destinyAccountResp.UserInfo.DisplayName,
		}
		for _, destinyCharacterResp := range destinyAccountResp.Characters {
			destinyCharacter := &db.DestinyCharacter{
				CharacterID: db.DestinyCharacterID(destinyCharacterResp.CharacterID),
				ClassName:   destinyCharacterResp.CharacterClass.ClassName,
			}
			destinyUser.DestinyCharacters = append(destinyUser.DestinyCharacters, destinyCharacter)
		}
		bungieUser.DestinyUsers = append(bungieUser.DestinyUsers, destinyUser)
	}

	// Create and insert the user into the database.
	if err := h.Server.DB.InsertBungieUser(bungieUser); err != nil {
		log.Printf("unable to write BungieUser to db: %v", err)
	}

	// Set the cookie.
	cookie := &http.Cookie{
		Name:  cookieBungieMembershipID,
		Value: string(bungieUser.MembershipID),
	}
	http.SetCookie(w, cookie)

	// Redirect to the original URL.
	http.Redirect(w, r, q["state"][0], http.StatusSeeOther)
}
