package api

import (
	"fmt"
)

type Request interface {
	URL() string
}

type Response interface {
	GetHeader() *Header
}

// A Header contains comment fields in a response from the Bungie API.
type Header struct {
	ErrorCode       int         `json:"ErrorCode"`
	ErrorStatus     string      `json:"ErrorStatus"`
	Message         string      `json:"Message"`
	MessageData     interface{} `json:"MessageData"`
	ThrottleSeconds int         `json:"ThrottleSeconds"`
}

type GetCurrentBungieAccountRequest struct{}

func (*GetCurrentBungieAccountRequest) URL() string {
	return "https://www.bungie.net/Platform/User/GetCurrentBungieAccount/"
}

type GetCurrentBungieAccountResponse struct {
	Header
	Response struct {
		BungieNetUser struct {
			DisplayName  string `json:"displayName"`
			MembershipID string `json:"membershipId"`
		} `json:"bungieNetUser"`
		DestinyAccounts []struct {
			Characters []struct {
				CharacterClass struct {
					ClassName string `json:"className"`
				} `json:"characterClass"`
				CharacterID string `json:"characterId"`
			} `json:"characters"`
			UserInfo struct {
				DisplayName    string `json:"displayName"`
				MembershipID   string `json:"membershipId"`
				MembershipType int    `json:"membershipType"`
			} `json:"userInfo"`
		} `json:"destinyAccounts"`
	} `json:"Response"`
}

func (r *GetCurrentBungieAccountResponse) GetHeader() *Header {
	return &r.Header
}

type MyCharacterVendorDataRequest struct {
	MembershipType int64
	CharacterHash  string
	VendorHash     string
}

func (r *MyCharacterVendorDataRequest) URL() string {
	return fmt.Sprintf("https://www.bungie.net/Platform/Destiny/%v/MyAccount/Character/%v/Vendor/%v/?definitions=true", r.MembershipType, r.CharacterHash, r.VendorHash)
}

type MyCharacterVendorDataResponse struct {
	Header
	Response struct {
		Data struct {
			VendorHash         int `json:"vendorHash"`
			SaleItemCategories []struct {
				CategoryTitle string `json:"categoryTitle"`
				SaleItems     []struct {
					FailureIndexes []int `json:"failureIndexes"`
					Item           struct {
						ItemHash int `json:"itemHash"`
					} `json:"item"`
				} `json:"saleItems"`
			} `json:"saleItemCategories"`
		} `json:"data"`
		Definitions struct {
			Items map[string]struct {
				ItemHash      int    `json:"itemHash"`
				ItemName      string `json:"itemName"`
				Icon          string `json:"icon"`
				SecondaryIcon string `json:"secondaryIcon"`
			} `json:"items"`
			VendorDetails map[string]struct {
				FailureStrings []string `json:"failureStrings"`
			} `json:"vendorDetails"`
		} `json:"definitions"`
	} `json:"Response"`
}

func (r *MyCharacterVendorDataResponse) GetHeader() *Header {
	return &r.Header
}
