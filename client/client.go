package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client  http.Client
	baseUrl string
	token   string
}

func NewClient(baseUrl string, timeout time.Duration) *Client {
	return &Client{
		baseUrl: baseUrl,
		client: http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) InitGame(nick, desc, targetNick string, wpbot bool) error {
	payload := GamePayload{Desc: desc, Nick: nick, TargetNick: targetNick, Wpbot: wpbot}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	payloadReader := bytes.NewReader(payloadJson)

	path, err := url.JoinPath(c.baseUrl, "/game")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, path, payloadReader)
	if err != nil {
		return err
	}

	slog.Info("client", slog.String("payload", string(payloadJson)))

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	slog.Info("client", slog.String("status", res.Status))
	c.token = res.Header.Get("X-Auth-Token")
	slog.Info("client", slog.String("token", c.token))
	return nil
}

func (c *Client) GetStatus() (status StatusData, err error) {
	path, err := url.JoinPath(c.baseUrl, "/game")
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Auth-Token", c.token)
	if err != nil {
		return
	}

	res, err := c.client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &status)
	if err != nil {
		return
	}
	return
}

func (c *Client) GetBoard() (board Board, err error) {
	path, err := url.JoinPath(c.baseUrl, "/game/board")
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Auth-Token", c.token)
	if err != nil {
		return
	}

	res, err := c.client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	fmt.Println("BODY")
	fmt.Println(string(body))

	err = json.Unmarshal(body, &board)
	return
}
