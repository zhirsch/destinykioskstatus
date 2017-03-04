package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"github.com/zhirsch/oauth2"

	"github.com/cenkalti/backoff"
)

type Client struct {
	AuthConfig *oauth2.Config
}

func (c *Client) GetBungieNetUser(tok *oauth2.Token) *GetBungieNetUserResponse {
	req := new(GetBungieNetUserRequest)
	resp := new(GetBungieNetUserResponse)
	c.get(tok, req, resp)
	return resp
}

func (c *Client) GetBungieAccount(tok *oauth2.Token, membershipID string) *GetBungieAccountResponse {
	req := &GetBungieAccountRequest{membershipID}
	resp := new(GetBungieAccountResponse)
	c.get(tok, req, resp)
	return resp
}

func (c *Client) MyCharacterVendorData(tok *oauth2.Token, characterHash, vendorHash string) *MyCharacterVendorDataResponse {
	req := &MyCharacterVendorDataRequest{characterHash, vendorHash}
	resp := new(MyCharacterVendorDataResponse)
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
				return fmt.Errorf("bad message: %+v", resp)
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
