package server

import (
	"html/template"

	"github.com/zhirsch/oauth2"

	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/db"
)

type Server struct {
	API      *api.Client
	Manifest *api.Manifest
	Template *template.Template
	DB       *db.DB
}

func NewServer(authConfig *oauth2.Config, manifestDBPath, userDBPath, templatePath string) (*Server, error) {
	s := &Server{
		API: &api.Client{authConfig},
	}

	if m, err := api.NewManifest(manifestDBPath); err != nil {
		panic(err)
	} else {
		s.Manifest = m
	}

	if db, err := db.NewDB(userDBPath); err != nil {
		panic(err)
	} else {
		s.DB = db
	}

	if t, err := template.ParseFiles(templatePath); err != nil {
		panic(err)
	} else {
		s.Template = t
	}

	return s, nil
}
