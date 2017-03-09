package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zhirsch/destinykioskstatus/api"
	"github.com/zhirsch/destinykioskstatus/db"
	"github.com/zhirsch/destinykioskstatus/server"
	"github.com/zhirsch/oauth2"
)

type VendorHandler struct {
	Server *server.Server
	Vendor api.Vendor
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
	vendorDefinition := h.Server.Manifest.GetDestinyVendorDefinition(h.Vendor)

	sources := &sources{
		client:         h.Server.API,
		token:          bungieUser.Token,
		membershipType: destinyUser.MembershipType,
		characterID:    characterID,
	}

	type Item struct {
		Description string
		Icon        string
		Missing     bool
		ForSale     bool
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
			itemDefinition := h.Server.Manifest.GetDestinyInventoryItemDefinition(saleItem.Item.ItemHash)
			item := Item{
				Description: getItemDescription(itemDefinition.ItemName, saleItem.FailureIndexes, vendorDefinition.FailureStrings),
				Icon:        fmt.Sprintf("https://www.bungie.net%s", itemDefinition.Icon),
			}
			for _, unlockStatus := range saleItem.UnlockStatuses {
				item.Missing = item.Missing || !unlockStatus.IsSet
			}
			if item.Missing {
				for _, sourceHash := range itemDefinition.SourceHashes {
					item.ForSale = item.ForSale || sources.isForSale(saleItem.Item.ItemHash, sourceHash)
				}
			}
			category.Items = append(category.Items, item)
		}
		data.Categories = append(data.Categories, category)
	}

	if err := h.Server.Template.Execute(w, data); err != nil {
		panic(err)
	}
}

func characterURL(u url.URL, destinyCharacter *db.DestinyCharacter) string {
	q := u.Query()
	q.Set("c", string(destinyCharacter.CharacterID))
	u.RawQuery = q.Encode()
	return u.String()
}

func getItemDescription(itemName string, failureIndexes []int, failureStrings []string) string {
	if len(failureIndexes) == 0 {
		return itemName
	}
	var description []string
	for _, failureIndex := range failureIndexes {
		description = append(description, failureStrings[failureIndex])
	}
	return itemName + "\n\n" + strings.Join(description, "\n")
}

var (
	sellingVendors = []api.Vendor{
		api.BountyTrackerVendor{},
		api.Cayde6Vendor{},
		api.CrucibleVendor{},
		api.CryptarchVendor{},
		api.DeadOrbitVendor{},
		api.ErisMornVendor{},
		api.EvaLevanteVendor{},
		api.EververseVendor{},
		api.FutureWarCultVendor{},
		api.GunsmithVendor{},
		api.IkoraReyVendor{},
		api.NewMonarchyVendor{},
		api.ShaxxVendor{},
		api.ShipwrightVendor{},
		api.TheSpeakerVendor{},
		api.VanguardVendor{},
		api.ZavalaVendor{},
	}
)

type cache struct {
	// Assumes that the vendors are selling the same thing for all users.
	entries map[uint32]*api.MyCharacterVendorDataResponse
	sync.RWMutex
}

func (c *cache) get(client *api.Client, token *oauth2.Token, membershipType db.DestinyMembershipType, characterID db.DestinyCharacterID, vendor api.Vendor) *api.MyCharacterVendorDataResponse {
	c.RLock()
	vendorResp, ok := c.entries[vendor.Hash()]
	if !ok || c.isExpired(vendorResp) {
		c.RUnlock()
		c.Lock()
		defer c.Unlock()
		vendorResp, ok = c.entries[vendor.Hash()]
		if !ok || c.isExpired(vendorResp) {
			log.Printf("getting vendor %v", vendor.Name())
			vendorResp = client.MyCharacterVendorData(token, membershipType, characterID, vendor.Hash())
			if c.entries == nil {
				c.entries = make(map[uint32]*api.MyCharacterVendorDataResponse)
			}
			c.entries[vendor.Hash()] = vendorResp
			t, err := time.Parse("2006-01-02T15:04:05Z", vendorResp.Response.Data.NextRefreshDate)
			if err != nil {
				panic(err)
			}
			log.Printf("vendor %v expires at %v", vendor.Name(), t)
		}
	} else {
		defer c.RUnlock()
	}
	return vendorResp
}

func (c *cache) isExpired(vendorResp *api.MyCharacterVendorDataResponse) bool {
	t, err := time.Parse("2006-01-02T15:04:05Z", vendorResp.Response.Data.NextRefreshDate)
	if err != nil {
		panic(err)
	}
	return time.Now().After(t)
}

var vendorCache cache

type sources struct {
	client         *api.Client
	token          *oauth2.Token
	membershipType db.DestinyMembershipType
	characterID    db.DestinyCharacterID
}

func (s *sources) isForSale(itemHash, sourceHash uint32) bool {
	forSale := make(map[uint32]bool)
	for _, vendor := range sellingVendors {
		vendorResp := vendorCache.get(s.client, s.token, s.membershipType, s.characterID, vendor)
		for _, saleItemCategory := range vendorResp.Response.Data.SaleItemCategories {
			for _, saleItem := range saleItemCategory.SaleItems {
				forSale[saleItem.Item.ItemHash] = true
			}
		}
	}
	return forSale[itemHash]
}
