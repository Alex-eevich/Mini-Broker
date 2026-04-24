package tbank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Token   string
	BaseURL string
}

type Client struct {
	Token   string
	BaseURL string
	Client  *http.Client
}

func Load(token string) *Config {
	return &Config{
		Token:   token,
		BaseURL: "https://sandbox-invest-public-api.tbank.ru/rest",
	}
}

func NewClient(token, baseURL string) *Client {
	return &Client{
		Token:   token,
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func NewClientFromToken(token string) *Client {
	cfg := Load(token)
	return NewClient(cfg.Token, cfg.BaseURL)
}

func (c *Client) do(method, path string, body any, out any) error {
	var buf *bytes.Buffer

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s", c.BaseURL, path),
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	// Для отладки новых запросов
	//log.Println(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

func ConnectTbankWithTokenCheck(token string) bool {
	url := "https://sandbox-invest-public-api.tbank.ru/rest/tinkoff.public.invest.api.contract.v1.UsersService/GetAccounts"

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer([]byte(`{}`)),
	)
	if err != nil {
		fmt.Println("request error:", err)
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http error:", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}
