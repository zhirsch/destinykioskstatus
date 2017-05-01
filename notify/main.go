package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/db"
	"github.com/zhirsch/destinykioskstatus/kiosk"
	"github.com/zhirsch/oauth2"
	"github.com/zhirsch/oauth2/bungie"
)

const (
	sendGridEndpoint = "/v3/mail/send"
)

var (
	vendorHashes = [...]uint32{
		3301500998,
		2420628997,
		2244880194,
		44395194,
		614738178,
		1460182514,
		3902439767,
	}
)

var (
	fromName     = flag.String("from_name", "Destiny Kiosk Status", "The from name.")
	fromAddr     = flag.String("from_addr", "noreply@destinykioskstatus.com", "The from email.")
	templatePath = flag.String("template", "notify.html", "The path to the HTML template file.")

	sendGridAPIKey = flag.String("sendgrid_apikey", "", "The SendGrid API key.")
	sendGridHost   = flag.String("sendgrid_host", "", "The SendGrid host.")

	bungieAPIKey         = flag.String("bungie_apikey", "", "The Bungie API key.")
	bungieAuthURL        = flag.String("bungie_authurl", "", "The Bungie auth URL.")
	bungieManifestDBPath = flag.String("bungie_manifestdb", "", "The path to the Bungie manifest db.")
	userDBPath           = flag.String("userdb", "", "The path to the user sqlite database.")
)

func main() {
	flag.Parse()
	if *sendGridAPIKey == "" {
		log.Fatal("need to provide --sendgrid_apikey")
	}
	if *bungieAPIKey == "" {
		log.Fatal("need to provide --bungie_apikey")
	}
	if *bungieAuthURL == "" {
		log.Fatal("need to provide --bungie_authurl")
	}
	if *bungieManifestDBPath == "" {
		log.Fatal("need to provide --bungie_manifestdb")
	}
	if *userDBPath == "" {
		log.Fatal("need to provide --userdb")
	}

	// Create the Bungie API client.
	authConfig := &oauth2.Config{
		ClientID:  *bungieAPIKey,
		Endpoint:  bungie.Endpoint(*bungieAuthURL),
		Exchanger: bungie.Exchanger{},
	}
	client := &api.Client{authConfig}

	// Load the Bungie manifest.
	manifest, err := api.NewManifest(*bungieManifestDBPath)
	if err != nil {
		panic(err)
	}

	// Load the user database.
	db, err := db.NewDB(*userDBPath)
	if err != nil {
		panic(err)
	}

	// Load the email template.
	templ, err := template.ParseFiles(*templatePath)
	if err != nil {
		panic(err)
	}

	// Select the Bungie user.
	//
	// TODO(zhirsch): Select all bungie users in a for loop.
	bungieUser, err := db.SelectBungieUser("12646688")
	if err != nil {
		panic(err)
	}

	// Get the Destiny user.
	//
	// TODO: Support multiple DestinyUsers on the same BungieUser.
	destinyUser := bungieUser.DestinyUsers[0]

	// Get the first character.  This assumes that the kiosk is the same for
	// all characters.  Probably not a bad assumption when considering
	// ships, shaders, sparrows, emblems, etc. for sale.
	characterID := destinyUser.DestinyCharacters[0].CharacterID

	// Iterate over all the vendor hashes.
	var data []kiosk.Data
	for _, vendorHash := range vendorHashes {
		data = append(data, kiosk.FetchKioskStatus(bungieUser, destinyUser, characterID, vendorHash, client, manifest))
	}

	from := mail.NewEmail(*fromName, *fromAddr)
	subject := "Destiny Kiosk Status Update"
	to := mail.NewEmail("", "zhirsch@umich.edu")

	buf := new(bytes.Buffer)
	if err := templ.Execute(buf, data); err != nil {
		panic(err)
	}
	content := mail.NewContent("text/html", buf.String())

	m := mail.NewV3MailInit(from, subject, to, content)

	request := sendgrid.GetRequest(*sendGridAPIKey, sendGridEndpoint, *sendGridHost)
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
