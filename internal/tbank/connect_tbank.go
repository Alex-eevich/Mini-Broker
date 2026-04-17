package tbank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Token   string
	BaseURL string
}

type Client struct {
	token   string
	baseURL string
	client  *http.Client
}

func Load(token string) *Config {
	return &Config{
		Token:   token,
		BaseURL: "https://sandbox-invest-public-api.tbank.ru/rest",
	}
}

func NewClient(token, baseURL string) *Client {
	return &Client{
		token:   token,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Config) ConnectTbank(token string) *Client {
	connect := Load(token)
	if connect.Token == "" {
		log.Println("Отсутствует token для подключения к песочнице")
	}
	Client := NewClient(connect.Token, connect.BaseURL)
	return Client
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
		fmt.Sprintf("%s/%s", c.baseURL, path),
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
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
