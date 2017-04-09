package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zhirsch/destinykioskstatus/db"
	"github.com/zhirsch/oauth2"
	"golang.org/x/net/context"

	"github.com/cenkalti/backoff"
)

type Client struct {
	AuthConfig *oauth2.Config
}

func (c *Client) GetCurrentBungieAccount(tok *oauth2.Token) *GetCurrentBungieAccountResponse {
	req := &GetCurrentBungieAccountRequest{}
	resp := new(GetCurrentBungieAccountResponse)
	c.get(tok, req, resp)
	return resp
}

func (c *Client) MyCharacterVendorData(tok *oauth2.Token, membershipType db.DestinyMembershipType, characterID db.DestinyCharacterID, vendorHash uint32) *MyCharacterVendorDataResponse {
	req := &MyCharacterVendorDataRequest{int64(membershipType), string(characterID), vendorHash}
	resp := new(MyCharacterVendorDataResponse)
	c.get(tok, req, resp)
	return resp
}

func (c *Client) GetAllVendorsForCurrentCharacter(tok *oauth2.Token, membershipType db.DestinyMembershipType, characterID db.DestinyCharacterID) *GetAllVendorsForCurrentCharacterResponse {
	req := &GetAllVendorsForCurrentCharacterRequest{int64(membershipType), string(characterID)}
	resp := new(GetAllVendorsForCurrentCharacterResponse)
	c.get(tok, req, resp)
	return resp
}

func (c *Client) get(tok *oauth2.Token, req Request, resp Response) {
	httpReq, err := http.NewRequest("GET", req.URL(), nil)
	if err != nil {
		panic(err)
	}
	httpReq.Header.Add("X-API-Key", c.AuthConfig.ClientID)

	client := c.AuthConfig.Client(context.TODO(), tok)
	err = backoff.RetryNotify(
		func() error {
			httpResp, err := client.Do(httpReq)
			if err != nil {
				return err
			}
			if httpResp.StatusCode != http.StatusOK {
				return fmt.Errorf("bad response: %v", httpResp.StatusCode)
			}
			if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
				return err
			}
			if resp.GetHeader().ErrorCode != 1 {
				return fmt.Errorf("bad message for %v: %+v", req.URL(), resp)
			}
			return nil
		},
		backoff.NewExponentialBackOff(),
		func(err error, dur time.Duration) {
			log.Printf("retrying: %v", err)
		},
	)
	if err != nil {
		panic(err)
	}
}
