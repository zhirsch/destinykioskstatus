package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/db"
	"github.com/zhirsch/destinykioskstatus/server"
)

type VendorHandler struct {
	Server *server.Server
	Vendor api.Vendor
}

func makeURL(u *url.URL, args ...string) *url.URL {
	if len(args)%2 != 0 {
		panic(fmt.Errorf("len(args) must be a multiple of two"))
	}

	var v url.URL = *u

	q := v.Query()
	for i := 0; i < len(args); i += 2 {
		q.Set(args[i], args[i+1])
	}
	v.RawQuery = q.Encode()

	return &v
}

func (h VendorHandler) ServeHTTP(u *db.User, w http.ResponseWriter, r *http.Request) {
	// Get the account info.
	accountResp := h.Server.API.GetBungieAccount(u.AuthToken.Value, u.ID)

	// Get the character to display info for.  If there isn't a character,
	// redirect to the first character.
	characterID := r.URL.Query().Get("c")
	if characterID == "" {
		u := makeURL(r.URL, "c", accountResp.Response.DestinyAccounts[0].Characters[0].CharacterID)
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}

	// Get the vendor info.
	vendorResp := h.Server.API.MyCharacterVendorData(u.AuthToken.Value, characterID, h.Vendor.Hash())
	failureStrings := vendorResp.Response.Definitions.VendorDetails[h.Vendor.Hash()].FailureStrings

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
		URL     string
	}
	type Data struct {
		Title            string
		User             string
		Characters       []Character
		CurrentCharacter string
		Categories       []Category
	}
	data := Data{
		Title:            h.Vendor.Name(),
		User:             u.Name,
		CurrentCharacter: characterID,
	}
	for _, account := range accountResp.Response.DestinyAccounts {
		for _, character := range account.Characters {
			data.Characters = append(data.Characters, Character{
				ID:      character.CharacterID,
				Class:   character.CharacterClass.ClassName,
				Current: character.CharacterID == characterID,
				URL:     makeURL(r.URL, "c", character.CharacterID).String(),
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

	if err := h.Server.Template.Execute(w, data); err != nil {
		panic(err)
	}
}
