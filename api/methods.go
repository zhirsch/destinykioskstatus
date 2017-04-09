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
	VendorHash     uint32
}

func (r *MyCharacterVendorDataRequest) URL() string {
	return fmt.Sprintf("https://www.bungie.net/Platform/Destiny/%v/MyAccount/Character/%v/Vendor/%v/", r.MembershipType, r.CharacterHash, r.VendorHash)
}

type MyCharacterVendorDataResponse struct {
	Header
	Response struct {
		Data struct {
			VendorHash         uint32 `json:"vendorHash"`
			NextRefreshDate    string `json:"nextRefreshDate"`
			SaleItemCategories []struct {
				CategoryTitle string `json:"categoryTitle"`
				SaleItems     []struct {
					FailureIndexes []int `json:"failureIndexes"`
					Item           struct {
						ItemHash uint32 `json:"itemHash"`
					} `json:"item"`
					UnlockStatuses []struct {
						IsSet          bool   `json:"isSet"`
						UnlockFlagHash uint32 `json:"unlockFlagHash"`
					} `json:"unlockStatuses"`
				} `json:"saleItems"`
			} `json:"saleItemCategories"`
		} `json:"data"`
	} `json:"Response"`
}

func (r *MyCharacterVendorDataResponse) GetHeader() *Header {
	return &r.Header
}

type GetAllVendorsForCurrentCharacterRequest struct {
	MembershipType int64
	CharacterHash  string
}

func (r *GetAllVendorsForCurrentCharacterRequest) URL() string {
	return fmt.Sprintf("https://www.bungie.net/Platform/Destiny/%v/MyAccount/Character/%v/Vendors/Summaries/", r.MembershipType, r.CharacterHash)
}

type GetAllVendorsForCurrentCharacterResponse struct {
	Header
	Response struct {
		Data struct {
			Vendors []struct {
				VendorHash      uint32 `json:"vendorHash"`
				NextRefreshDate string `json:"nextRefreshDate"`
				Enabled         bool   `json:"enabled"`
			} `json:"vendors"`
		} `json:"data"`
	} `json:"Response"`
}

func (r *GetAllVendorsForCurrentCharacterResponse) GetHeader() *Header {
	return &r.Header
}
