package client

import (
	"bytes"
	"encoding/json"
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

	slog.Info("client [InitGame]", slog.Any("payload", payload))

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	slog.Info("client [InitGame]", slog.Int("statusCode", res.StatusCode))
	c.token = res.Header.Get("X-Auth-Token")
	slog.Info("client [InitGame]", slog.String("token", c.token))
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
	slog.Info("client [GetStatus]", slog.Int("statusCode", res.StatusCode))
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &status)
	if err != nil {
		return
	}
	slog.Info("client [GetStatus]", slog.Any("status", status))
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

	slog.Info("client [GetBoard]", slog.Int("statusCode", res.StatusCode))
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &board)
	slog.Info("client [GetBoard]", slog.Any("board", board.Board))
	return
}

func (c *Client) GetDescription() (status StatusData, err error) {
	path, err := url.JoinPath(c.baseUrl, "/game/desc")
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

	slog.Info("client [GetDescription]", slog.Int("statusCode", res.StatusCode))
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &status)
	if err != nil {
		return
	}
	slog.Info("client [GetDescription]", slog.Any("status", status))
	return
}

func (c *Client) Fire(coord string) (answer FireAnswer, err error) {
	payload := FirePayload{Coord: coord}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return
	}

	payloadReader := bytes.NewReader(payloadJson)

	path, err := url.JoinPath(c.baseUrl, "/game/fire")
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, path, payloadReader)
	req.Header.Set("X-Auth-Token", c.token)
	if err != nil {
		return
	}

	slog.Info("client [Fire]", slog.Any("payload", payload))

	res, err := c.client.Do(req)
	if err != nil {
		return
	}

	slog.Info("client [Fire]", slog.Int("statusCode", res.StatusCode))
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &answer)
	if err != nil {
		return
	}
	slog.Info("client [Fire]", slog.Any("answer", answer))

	return
}
