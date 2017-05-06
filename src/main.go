package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/zhirsch/oauth2"
	"github.com/zhirsch/oauth2/bungie"

	"github.com/zhirsch/destinykioskstatus/src/handler"
	"github.com/zhirsch/destinykioskstatus/src/server"
)

var (
	addr           = flag.String("addr", ":443", "The address to listen on.")
	apiKey         = flag.String("apikey", "", "The Bungie API key.")
	authURL        = flag.String("authurl", "", "The Bungie auth URL.")
	manifestDBPath = flag.String("manifestdb", "", "The path to the manifest sqlite database.")
	userDBPath     = flag.String("userdb", "", "The path to the user sqlite database.")
	templatePath   = flag.String("template", "kiosk.html", "The path to the HTML template file.")
	tlsCertPath    = flag.String("tlscert", "server.crt", "The path to the  TLS certificate file.")
	tlsKeyPath     = flag.String("tlskey", "server.key", "The path to the TLS key file.")
	mediaPath      = flag.String("media", "", "The path to the media directory.")
)

func main() {
	flag.Parse()
	if *apiKey == "" {
		log.Fatal("need to provide --apikey")
	}
	if *authURL == "" {
		log.Fatal("need to provide --authurl")
	}
	if *manifestDBPath == "" {
		log.Fatal("need to provide --manifestdb")
	}
	if *userDBPath == "" {
		log.Fatal("need to provide --userdb")
	}
	if *templatePath == "" {
		log.Fatal("need to provide --template")
	}
	if *mediaPath == "" {
		log.Fatal("need to provide --media")
	}

	authConfig := &oauth2.Config{
		ClientID:  *apiKey,
		Endpoint:  bungie.Endpoint(*authURL),
		Exchanger: bungie.Exchanger{},
	}

	s, err := server.NewServer(authConfig, *manifestDBPath, *userDBPath, *templatePath)
	if err != nil {
		log.Fatal(err)
	}

	handlers := map[string]http.Handler{
		"/BungieAuthCallback": handler.BungieAuthCallbackHandler{s, authConfig},
		"/media/":             http.StripPrefix("/media/", http.FileServer(http.Dir(*mediaPath))),
	}
	authedHandlers := map[string]handler.Handler{
		"/emblems":  handler.VendorHandler{s, 3301500998},
		"/shaders":  handler.VendorHandler{s, 2420628997},
		"/ships":    handler.VendorHandler{s, 2244880194},
		"/sparrows": handler.VendorHandler{s, 44395194},
		"/emotes":   handler.VendorHandler{s, 614738178},
		"/weapons":  handler.VendorHandler{s, 1460182514},
		"/armor":    handler.VendorHandler{s, 3902439767},
	}
	for p, h := range authedHandlers {
		handlers[p] = handler.AuthenticationMiddlewareHandler{s, authConfig, h}
	}
	for p, h := range handlers {
		http.Handle(p, handler.StackTraceMiddlewareHandler{h})
	}

	if err := http.ListenAndServeTLS(*addr, *tlsCertPath, *tlsKeyPath, nil); err != nil {
		log.Fatal(err)
	}
}
