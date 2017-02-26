package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) (*Client, error) {
	return &Client{apiKey}, nil
}

func (c *Client) GetAccessTokensFromCode(code string) *GetAccessTokensFromCodeResponse {
	req := &GetAccessTokensFromCodeRequest{code}
	resp := new(GetAccessTokensFromCodeResponse)
	c.post(req, resp)
	return resp
}

func (c *Client) GetBungieNetUser(auth string) *GetBungieNetUserResponse {
	req := new(GetBungieNetUserRequest)
	resp := new(GetBungieNetUserResponse)
	c.get(req, resp, auth)
	return resp
}

func (c *Client) GetBungieAccount(auth, membershipID string) *GetBungieAccountResponse {
	req := &GetBungieAccountRequest{membershipID}
	resp := new(GetBungieAccountResponse)
	c.get(req, resp, auth)
	return resp
}

func (c *Client) MyCharacterVendorData(auth, characterHash, vendorHash string) *MyCharacterVendorDataResponse {
	req := &MyCharacterVendorDataRequest{characterHash, vendorHash}
	resp := new(MyCharacterVendorDataResponse)
	c.get(req, resp, auth)
	return resp
}

func (c *Client) get(req Request, resp Response, auth string) {
	httpReq, err := http.NewRequest("GET", req.URL(), nil)
	if err != nil {
		panic(err)
	}
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth))
	c.do(httpReq, resp)
}

func (c *Client) post(req Request, resp Response) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(req); err != nil {
		panic(err)
	}
	httpReq, err := http.NewRequest("POST", req.URL(), &body)
	if err != nil {
		panic(err)
	}
	httpReq.Header.Add("Content-Type", "application/json")
	c.do(httpReq, resp)
}

func (c *Client) do(httpReq *http.Request, resp Response) {
	httpReq.Header.Add("X-API-Key", c.apiKey)
	err := backoff.RetryNotify(
		func() error {
			httpResp, err := http.DefaultClient.Do(httpReq)
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
