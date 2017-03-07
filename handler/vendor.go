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

func characterURL(u url.URL, destinyCharacter *db.DestinyCharacter) string {
	q := u.Query()
	q.Set("c", string(destinyCharacter.CharacterID))
	u.RawQuery = q.Encode()
	return u.String()
}

func (h VendorHandler) ServeHTTP(bungieUser *db.BungieUser, w http.ResponseWriter, r *http.Request) {
	// TODO: Support multiple DestinyUsers on the same BungieUser.
	destinyUser := bungieUser.DestinyUsers[0]

	// Get the character to display info for.  If there isn't a character,
	// redirect to the first character.
	characterID := db.DestinyCharacterID(r.URL.Query().Get("c"))
	if characterID == "" {
		http.Redirect(w, r, characterURL(*r.URL, destinyUser.DestinyCharacters[0]), http.StatusFound)
		return
	}

	// Get the vendor info.
	vendorResp := h.Server.API.MyCharacterVendorData(bungieUser.Token, destinyUser.MembershipType, characterID, h.Vendor.Hash())
	failureStrings := vendorResp.Response.Definitions.VendorDetails[h.Vendor.Hash()].FailureStrings

	type Item struct {
		Description string
		Icon        string
		Missing     bool
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
		User:             destinyUser.DisplayName,
		CurrentCharacter: string(characterID),
	}
	for _, character := range destinyUser.DestinyCharacters {
		data.Characters = append(data.Characters, Character{
			ID:      string(character.CharacterID),
			Class:   character.ClassName,
			Current: character.CharacterID == characterID,
			URL:     characterURL(*r.URL, character),
		})
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
				item.Missing = false
			} else {
				item.Description = fmt.Sprintf("%s\n\n%s",
					definition.ItemName,
					failureStrings[saleItem.FailureIndexes[0]],
				)
				item.Missing = true
			}
			category.Items = append(category.Items, item)
		}
		data.Categories = append(data.Categories, category)
	}

	if err := h.Server.Template.Execute(w, data); err != nil {
		panic(err)
	}
}
