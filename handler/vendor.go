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

func (h VendorHandler) ServeHTTP(u *db.User, w http.ResponseWriter, r *http.Request) {
	// Get the account info.
	accountResp, err := h.Server.API.GetBungieAccount(u.ID)
	if err != nil {
		panic(err)
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
	vendorResp, err := h.Server.API.MyCharacterVendorData(characterID, h.Vendor.Hash())
	if err != nil {
		panic(err)
	}
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
	}
	type Data struct {
		Title      string
		User       string
		Characters []Character
		Categories []Category
	}
	data := Data{
		Title: h.Vendor.Name(),
		User:  u.Name,
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

	if err := h.Server.Template.Execute(w, data); err != nil {
		panic(err)
	}
}
