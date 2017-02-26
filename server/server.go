package server

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/db"
)

const userCookieName = "X-DestinyKioskStatus-User"

var ErrNeedAuth = errors.New("ErrNeedAuth")

type Server struct {
	API      *api.Client
	Template *template.Template
	DB       *db.DB
}

func NewServer(apiKey, authURL, templatePath, dbPath string) (*Server, error) {
	s := &Server{}

	api, err := api.NewClient(apiKey, authURL)
	if err != nil {
		panic(err)
	}
	s.API = api

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		panic(err)
	}
	s.Template = t

	db, err := db.NewDB(dbPath)
	if err != nil {
		panic(err)
	}
	s.DB = db

	return s, nil
}

func (s *Server) GetUserFromAuth(w http.ResponseWriter, r *http.Request) (*db.User, error) {
	if !s.API.Authenticate(w, r) {
		return nil, ErrNeedAuth
	}

	// Get the user info.
	userResp, err := s.API.GetBungieNetUser()
	if err != nil {
		return nil, err
	}

	user := &db.User{
		ID:           userResp.Response.User.MembershipID,
		Name:         userResp.Response.User.DisplayName,
		AuthToken:    s.API.AuthToken,
		RefreshToken: s.API.RefreshToken,
	}

	// Insert the bungie auth into the database and set the cookie.
	if err := s.DB.InsertUser(user); err != nil {
		log.Printf("unable to write bungie auth to db: %v", err)
	}

	http.SetCookie(w, &http.Cookie{Name: "X-DestinyKioskStatus-User", Value: user.ID})
	return user, nil
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) (*db.User, error) {
	// Get the cookie.
	cookie, err := r.Cookie(userCookieName)
	if err == http.ErrNoCookie {
		log.Print("user cookie not set; doing authentication")
		return s.GetUserFromAuth(w, r)
	} else if err != nil {
		return nil, err
	}

	// Get the bungie tokens from the db.
	user, err := s.DB.SelectUser(cookie.Value)
	if err != nil {
		log.Printf("no stored user for %v; doing authentication", cookie.Value)
		return s.GetUserFromAuth(w, r)
	}
	return user, nil
}
