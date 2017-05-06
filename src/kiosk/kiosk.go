package kiosk

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/zhirsch/destinykioskstatus/src/api"
	"github.com/zhirsch/destinykioskstatus/src/db"
	"github.com/zhirsch/oauth2"
)

var vendorIdentifierBlacklist = map[string]bool{
	"VENDOR_BOUNTY_TRACKER":      true,
	"VENDOR_KIOSK_EMBLEMS":       true,
	"VENDOR_KIOSK_EMOTES":        true,
	"VENDOR_KIOSK_EXOTIC_ARMOR":  true,
	"VENDOR_KIOSK_EXOTIC_WEAPON": true,
	"VENDOR_KIOSK_HOLIDAY":       true,
	"VENDOR_KIOSK_SHADERS":       true,
	"VENDOR_KIOSK_SHIPS":         true,
	"VENDOR_KIOSK_VEHICLES":      true,
	"VENDOR_POSTMASTER":          true,
	"VENDOR_REEF_POSTMASTER":     true,
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

type Data struct {
	Title      string
	User       string
	Categories []Category
}

func (d Data) MissingAndForSale() bool {
	for _, category := range d.Categories {
		if category.MissingAndForSale() {
			return true
		}
	}
	return false
}

func (c Category) MissingAndForSale() bool {
	for _, item := range c.Items {
		if item.Missing && item.ForSale {
			return true
		}
	}
	return false
}

func FetchKioskStatus(bungieUser *db.BungieUser, destinyUser *db.DestinyUser, characterID db.DestinyCharacterID, vendorHash uint32, client *api.Client, manifest *api.Manifest) Data {
	// Get the vendor info.
	vendorResp := client.MyCharacterVendorData(bungieUser.Token, destinyUser.MembershipType, characterID, vendorHash)
	vendorDefinition := manifest.GetDestinyVendorDefinition(vendorHash)

	// Get the items that are for sale for this user.
	itemsForSale := getItemsForSale(destinyUser.MembershipType, characterID, client, manifest, bungieUser.Token)

	data := Data{
		Title: vendorDefinition.Summary.VendorName,
		User:  destinyUser.DisplayName,
	}
	for _, saleItemCategory := range vendorResp.Response.Data.SaleItemCategories {
		category := Category{Title: saleItemCategory.CategoryTitle}
		for _, saleItem := range saleItemCategory.SaleItems {
			itemDefinition := manifest.GetDestinyInventoryItemDefinition(saleItem.Item.ItemHash)
			item := Item{
				Description: getItemDescription(itemDefinition.ItemName, saleItem.FailureIndexes, vendorDefinition.FailureStrings),
				Icon:        fmt.Sprintf("https://www.bungie.net%s", itemDefinition.Icon),
			}
			for _, unlockStatus := range saleItem.UnlockStatuses {
				item.Missing = item.Missing || !unlockStatus.IsSet
			}
			if item.Missing {
				_, itemForSale := itemsForSale[saleItem.Item.ItemHash]
				item.ForSale = item.ForSale || itemForSale
			}
			category.Items = append(category.Items, item)
		}
		data.Categories = append(data.Categories, category)
	}
	return data
}

func getItemsForSale(membershipType db.DestinyMembershipType, characterID db.DestinyCharacterID, client *api.Client, manifest *api.Manifest, token *oauth2.Token) map[uint32]bool {
	allVendorsResp := client.GetAllVendorsForCurrentCharacter(token, membershipType, characterID)

	forSale := make(map[uint32]bool)
	for _, vendor := range allVendorsResp.Response.Data.Vendors {
		if !vendor.Enabled {
			continue
		}
		vendorDefinition := manifest.GetDestinyVendorDefinition(vendor.VendorHash)
		if _, ok := vendorIdentifierBlacklist[vendorDefinition.Summary.VendorIdentifier]; ok {
			continue
		}
		vendorResp := vendorCache.get(client, manifest, token, membershipType, characterID, vendorDefinition)
		for _, saleItemCategory := range vendorResp.Response.Data.SaleItemCategories {
			for _, saleItem := range saleItemCategory.SaleItems {
				forSale[saleItem.Item.ItemHash] = true
			}
		}
	}
	return forSale
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

type cache struct {
	// Assumes that the vendors are selling the same thing for all users.
	entries map[uint32]*api.MyCharacterVendorDataResponse
	sync.RWMutex
}

func (c *cache) get(client *api.Client, manifest *api.Manifest, token *oauth2.Token, membershipType db.DestinyMembershipType, characterID db.DestinyCharacterID, vendorDefinition *api.DestinyVendorDefinition) *api.MyCharacterVendorDataResponse {
	c.RLock()
	vendorResp, ok := c.entries[vendorDefinition.Hash]
	if !ok || c.isExpired(vendorResp) {
		c.RUnlock()
		c.Lock()
		defer c.Unlock()
		vendorResp, ok = c.entries[vendorDefinition.Hash]
		if !ok || c.isExpired(vendorResp) {
			log.Printf("getting vendor %v (%v)", vendorDefinition.Summary.VendorName, vendorDefinition.Summary.VendorIdentifier)
			vendorResp = client.MyCharacterVendorData(token, membershipType, characterID, vendorDefinition.Hash)
			if c.entries == nil {
				c.entries = make(map[uint32]*api.MyCharacterVendorDataResponse)
			}
			c.entries[vendorDefinition.Hash] = vendorResp
			t, err := time.Parse("2006-01-02T15:04:05Z", vendorResp.Response.Data.NextRefreshDate)
			if err != nil {
				panic(err)
			}
			log.Printf("vendor %v expires at %v", vendorDefinition.Summary.VendorName, t)
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
