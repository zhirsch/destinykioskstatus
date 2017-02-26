package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/handlers"
	"github.com/zhirsch/destinykioskstatus/server"
)

var (
	addr         = flag.String("addr", ":443", "The address to listen on.")
	apiKey       = flag.String("apikey", "", "The Bungie API key.")
	authURL      = flag.String("authurl", "", "The Bungie auth URL.")
	dbPath       = flag.String("db", "", "The path to the sqlite database.")
	templatePath = flag.String("template", "kiosk.html", "The path to the HTML template file.")
	tlsCertPath  = flag.String("tlscert", "server.crt", "The path to the  TLS certificate file.")
	tlsKeyPath   = flag.String("tlskey", "server.key", "The path to the TLS key file.")
)

type StackTraceMiddlewareHandler struct {
	handler http.Handler
}

func (h StackTraceMiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			trace := debug.Stack()
			http.Error(w, fmt.Sprintf("%v\n\n%v", p, string(trace)), http.StatusInternalServerError)
		}
	}()
	h.handler.ServeHTTP(w, r)
}

func main() {
	flag.Parse()
	if *apiKey == "" {
		log.Fatal("need to provide --apikey")
	}
	if *authURL == "" {
		log.Fatal("need to provide --authurl")
	}
	if *dbPath == "" {
		log.Fatal("need to provide --db")
	}

	s, err := server.NewServer(*apiKey, *authURL, *templatePath, *dbPath)
	if err != nil {
		log.Fatal(err)
	}

	handlers := map[string]http.Handler{
		"/emblems":  handlers.VendorHandler{s, api.EmblemKioskVendor{}},
		"/shaders":  handlers.VendorHandler{s, api.ShaderKioskVendor{}},
		"/ships":    handlers.VendorHandler{s, api.ShipKioskVendor{}},
		"/sparrows": handlers.VendorHandler{s, api.SparrowKioskVendor{}},
		"/emotes":   handlers.VendorHandler{s, api.EmoteKioskVendor{}},
		"/weapons":  handlers.VendorHandler{s, api.ExoticWeaponKioskVendor{}},
		"/armor":    handlers.VendorHandler{s, api.ExoticArmorKioskVendor{}},

		"/BungieAuthCallback": http.HandlerFunc(s.API.HandleBungieAuthCallback),
	}
	for pattern, handler := range handlers {
		http.Handle(pattern, StackTraceMiddlewareHandler{handler})
	}

	if err := http.ListenAndServeTLS(*addr, *tlsCertPath, *tlsKeyPath, nil); err != nil {
		log.Fatal(err)
	}
}
