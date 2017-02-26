package server

import (
	"html/template"

	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/db"
)

type Server struct {
	API      *api.Client
	Template *template.Template
	DB       *db.DB
}

func NewServer(apiKey, templatePath, dbPath string) (*Server, error) {
	s := &Server{}

	api, err := api.NewClient(apiKey)
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
