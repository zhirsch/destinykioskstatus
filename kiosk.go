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
	addr         = flag.String("addr", ":443", "The address to listen on.")
	apiKey       = flag.String("apikey", "", "The Bungie API key.")
	authURL      = flag.String("authurl", "", "The Bungie auth URL.")
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

type kioskHandler struct {
	server *server
	vendor destinyapi.Vendor
}

func (h kioskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.server.client.Authenticate(w, r) {
		return
	}

	// Get the user info.
	userResp, err := h.server.client.GetBungieNetUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the account info.
	accountResp, err := h.server.client.GetBungieAccount(userResp.Response.User.MembershipID)
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
	vendorResp, err := h.server.client.MyCharacterVendorData(characterID, h.vendor.Hash())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	failureStrings := vendorResp.Response.Definitions.VendorDetails[h.vendor.Hash()].FailureStrings

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
		Title: h.vendor.Name(),
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
				Icon: fmt.Sprintf("https://www.bungie.net%s", definition.Icon),
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

	if err := h.server.template.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.Parse()
	if *apiKey == "" {
		log.Fatal("need to provide --apikey")
	}
	if *authURL == "" {
		log.Fatal("need to provide --authurl")
	}

	s, err := newServer(*apiKey, *authURL, *templatePath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/BungieAuthCallback", s.client.HandleBungieAuthCallback)
	http.Handle("/emblems", kioskHandler{s, destinyapi.EmblemKioskVendor{}})
	http.Handle("/shaders", kioskHandler{s, destinyapi.ShaderKioskVendor{}})
	http.Handle("/ships", kioskHandler{s, destinyapi.ShipKioskVendor{}})
	http.Handle("/sparrows", kioskHandler{s, destinyapi.SparrowKioskVendor{}})
	http.Handle("/emotes", kioskHandler{s, destinyapi.EmoteKioskVendor{}})
	http.Handle("/weapons", kioskHandler{s, destinyapi.ExoticWeaponKioskVendor{}})
	http.Handle("/armor", kioskHandler{s, destinyapi.ExoticArmorKioskVendor{}})
	if err := http.ListenAndServeTLS(*addr, *tlsCertPath, *tlsKeyPath, nil); err != nil {
		log.Fatal(err)
	}
}
