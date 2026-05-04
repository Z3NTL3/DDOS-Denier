package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	Token      string
	HttpClient *http.Client
	Headers    http.Header
}

type (
	ZoneInfo struct {
		ID string `json:"id"`
	}
	Zone struct {
		Success bool       `json:"success"`
		Result  []ZoneInfo `json:"result"`
	}
)

func New(token string) *Client {
	return &Client{
		Token:      token,
		HttpClient: &http.Client{},
		Headers: http.Header{
			"Authorization": []string{"Bearer " + token},
			"Content-Type":  []string{"application/json"},
		},
	}
}

func (c *Client) VerifyToken() error {
	var b io.ReadCloser

	req, err := http.NewRequest(http.MethodGet, VerifyEndpoint, b)
	if err != nil {
		return err
	}

	req.Header = c.Headers

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("[Token Not Valid] %s", string(b))
	}

	return nil
}

func (c *Client) GetZone(domain string) (*Zone, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/?name=%s", ListZoneEndpoint, domain), nil)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Cloudflare API returned with failure response: %s", resp.Status)
	}

	model := &Zone{}

	b, err := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, model); err != nil {
		return nil, err
	}

	if !model.Success {
		return nil, fmt.Errorf("Cloudflare API returned with failure response: %s", resp.Status)
	}

	return model, err
}

func (c *Client) UpdateZone(mode string, zone_id string) error {
	body := struct {
		Value string `json:"value"`
	}{Value: mode}

	jsonRaw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	rd := bytes.NewReader(jsonRaw)

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s/settings/security_level", ListZoneEndpoint, zone_id), rd)
	if err != nil {
		return err
	}

	req.Header = c.Headers
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyraw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	m := struct {
		Success bool `json:"success"`
	}{}
	if err := json.Unmarshal(bodyraw, &m); err != nil {
		return err
	}

	if !m.Success {
		return errors.New("Cloudflare API returned with failure")
	}

	return nil
}
