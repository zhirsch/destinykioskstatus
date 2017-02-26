package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/zhirsch/destinykioskstatus/api"
)

const (
	USER_COOKIE_NAME = "X-DestinyKioskStatus-User"
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

type server struct {
	api      *api.Client
	template *template.Template
	db       *DB
}

func newServer(apiKey, authURL, templatePath string) (*server, error) {
	s := &server{}

	api, err := api.NewClient(apiKey, authURL)
	if err != nil {
		return nil, err
	} else {
		s.api = api
	}

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	} else {
		s.template = t
	}

	db, err := NewDB(*dbPath)
	if err != nil {
		return nil, err
	} else {
		s.db = db
	}

	return s, nil
}

type kioskHandler struct {
	server *server
	vendor api.Vendor
}

var ErrNeedAuth = errors.New("ErrNeedAuth")

func (h kioskHandler) getUserFromAuth(w http.ResponseWriter, r *http.Request) (*User, error) {
	if !h.server.api.Authenticate(w, r) {
		return nil, ErrNeedAuth
	}

	// Get the user info.
	userResp, err := h.server.api.GetBungieNetUser()
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           userResp.Response.User.MembershipID,
		Name:         userResp.Response.User.DisplayName,
		AuthToken:    h.server.api.AuthToken,
		RefreshToken: h.server.api.RefreshToken,
	}

	// Insert the bungie auth into the database and set the cookie.
	if err := h.server.db.InsertUser(user); err != nil {
		log.Printf("unable to write bungie auth to db: %v", err)
	}

	http.SetCookie(w, &http.Cookie{Name: "X-DestinyKioskStatus-User", Value: user.ID})
	return user, nil
}

func (h kioskHandler) getUser(w http.ResponseWriter, r *http.Request) (*User, error) {
	// Get the cookie.
	cookie, err := r.Cookie(USER_COOKIE_NAME)
	if err == http.ErrNoCookie {
		log.Print("user cookie not set; doing authentication")
		return h.getUserFromAuth(w, r)
	} else if err != nil {
		return nil, err
	}

	// Get the bungie tokens from the db.
	user, err := h.server.db.SelectUser(cookie.Value)
	if err != nil {
		log.Printf("no stored user for %v; doing authentication", cookie.Value)
		return h.getUserFromAuth(w, r)
	}
	h.server.api.AuthToken = user.AuthToken
	h.server.api.RefreshToken = user.RefreshToken
	return user, nil
}

func (h kioskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(w, r)
	if err != nil {
		if err != ErrNeedAuth {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get the account info.
	accountResp, err := h.server.api.GetBungieAccount(user.ID)
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
	vendorResp, err := h.server.api.MyCharacterVendorData(characterID, h.vendor.Hash())
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
		User:  user.Name,
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
	if *dbPath == "" {
		log.Fatal("need to provide --db")
	}

	s, err := newServer(*apiKey, *authURL, *templatePath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/BungieAuthCallback", s.api.HandleBungieAuthCallback)
	http.Handle("/emblems", kioskHandler{s, api.EmblemKioskVendor{}})
	http.Handle("/shaders", kioskHandler{s, api.ShaderKioskVendor{}})
	http.Handle("/ships", kioskHandler{s, api.ShipKioskVendor{}})
	http.Handle("/sparrows", kioskHandler{s, api.SparrowKioskVendor{}})
	http.Handle("/emotes", kioskHandler{s, api.EmoteKioskVendor{}})
	http.Handle("/weapons", kioskHandler{s, api.ExoticWeaponKioskVendor{}})
	http.Handle("/armor", kioskHandler{s, api.ExoticArmorKioskVendor{}})
	if err := http.ListenAndServeTLS(*addr, *tlsCertPath, *tlsKeyPath, nil); err != nil {
		log.Fatal(err)
	}
}
