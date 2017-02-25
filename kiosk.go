package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/zhirsch/destinyapi"
)

var (
	apiKey       = flag.String("apikey", "84fb2e09fafe4573bbc45e5c013c9029", "The Bungie API key.")
	authURL      = flag.String("authurl", "https://www.bungie.net/en/Application/Authorize/11251", "The Bungie auth URL.")
	templatePath = flag.String("template", "kiosk.html", "The path to the HTML template file.")
	tlsCertPath  = flag.String("tlscert", "server.crt", "The path to the  TLS certificate file.")
	tlsKeyPath   = flag.String("tlskey", "server.key", "The path to the TLS key file.")
)

type server struct {
	client   *destinyapi.Client
	template *template.Template
}

func newServer(apiKey, authURL, templatePath string) (*server, error) {
	c, err := destinyapi.NewClient(apiKey, authURL)
	if err != nil {
		return nil, err
	}

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	return &server{client: c, template: t}, nil
}

func (s *server) handleEmblems(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.EmblemKioskVendorHash, "Emblems")
}

func (s *server) handleShaders(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.ShaderKioskVendorHash, "Shaders")
}

func (s *server) handleShips(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.ShipKioskVendorHash, "Ships")
}

func (s *server) handleSparrows(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.SparrowKioskVendorHash, "Sparrows")
}

func (s *server) handleEmotes(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.EmoteKioskVendorHash, "Emotes")
}

func (s *server) handleWeapons(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.ExoticWeaponKioskVendorHash, "Exotic Weapons")
}

func (s *server) handleArmor(w http.ResponseWriter, r *http.Request) {
	if !s.client.Authenticate(w, r) {
		return
	}
	s.handleVendor(w, r, destinyapi.ExoticArmorKioskVendorHash, "Exotic Armor")
}

func (s *server) handleVendor(w http.ResponseWriter, r *http.Request, vendorHash int, title string) {
	// Get the user info.
	userResp, err := s.client.GetBungieNetUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the account info.
	accountResp, err := s.client.GetBungieAccount(userResp.Response.User.MembershipID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the character to display info for.  If there isn't a character,
	// redirect to the first character.
	characterID := r.URL.Query().Get("c")
	if characterID == "" {
		var u url.URL = *r.URL
		q := u.Query()
		q.Set("c", accountResp.Response.DestinyAccounts[0].Characters[0].CharacterID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}

	// Get the vendor info.
	vendorHashString := fmt.Sprintf("%v", vendorHash)
	vendorResp, err := s.client.MyCharacterVendorData(characterID, vendorHashString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	failureStrings := vendorResp.Response.Definitions.VendorDetails[vendorHashString].FailureStrings

	type Item struct {
		Description string
		Icon        string
		Owned       bool
	}
	type Category struct {
		Title string
		Items []Item
	}
	type Character struct {
		ID      string
		Class   string
		Current bool
	}
	type Data struct {
		Title      string
		User       string
		Characters []Character
		Categories []Category
	}
	data := Data{
		Title: title,
		User:  userResp.Response.User.DisplayName,
	}
	for _, account := range accountResp.Response.DestinyAccounts {
		for _, character := range account.Characters {
			data.Characters = append(data.Characters, Character{
				ID:      character.CharacterID,
				Class:   character.CharacterClass.ClassName,
				Current: character.CharacterID == characterID,
			})
		}
	}
	for _, saleItemCategory := range vendorResp.Response.Data.SaleItemCategories {
		category := Category{Title: saleItemCategory.CategoryTitle}
		for _, saleItem := range saleItemCategory.SaleItems {
			itemHashString := fmt.Sprintf("%v", saleItem.Item.ItemHash)
			definition := vendorResp.Response.Definitions.Items[itemHashString]

			item := Item{
				Icon: fmt.Sprintf("http://www.bungie.net%s", definition.Icon),
			}
			if len(saleItem.FailureIndexes) == 0 {
				item.Description = definition.ItemName
				item.Owned = true
			} else {
				item.Description = fmt.Sprintf("%s\n\n%s",
					definition.ItemName,
					failureStrings[saleItem.FailureIndexes[0]],
				)
				item.Owned = false
			}
			category.Items = append(category.Items, item)
		}
		data.Categories = append(data.Categories, category)
	}

	if err := s.template.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.Parse()

	s, err := newServer(*apiKey, *authURL, *templatePath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/BungieAuthCallback", s.client.HandleBungieAuthCallback)
	http.HandleFunc("/emblems", s.handleEmblems)
	http.HandleFunc("/shaders", s.handleShaders)
	http.HandleFunc("/ships", s.handleShips)
	http.HandleFunc("/sparrows", s.handleSparrows)
	http.HandleFunc("/emotes", s.handleEmotes)
	http.HandleFunc("/weapons", s.handleWeapons)
	http.HandleFunc("/armor", s.handleArmor)
	if err := http.ListenAndServeTLS(":8080", *tlsCertPath, *tlsKeyPath, nil); err != nil {
		log.Fatal(err)
	}
}
