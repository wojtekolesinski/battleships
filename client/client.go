package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/wojtekolesinski/battleships/models"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrUnauthorized       = fmt.Errorf("unauthorized")
	ErrForbidden          = fmt.Errorf("forbidden")
	ErrServiceUnavailable = fmt.Errorf("service unavailable")
	ErrNotFound           = fmt.Errorf("not found")
	ErrBadRequest         = fmt.Errorf("bad request")
)

type Client struct {
	*http.Client
	baseUrl string
	token   string
}

func NewClient(baseUrl string, timeout time.Duration) *Client {
	return &Client{
		baseUrl: baseUrl,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) newRequestWithToken(method string, path string, body io.Reader) (*http.Request, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("client.newRequest: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.token)

	return req, nil
}

func (c *Client) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
	path, err := url.JoinPath(c.baseUrl, path)
	if err != nil {
		return nil, fmt.Errorf("url.JoinPath: %w", err)
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("http.newRequest: %w", err)
	}

	return req, nil
}

func (c *Client) InitGame(payload models.GamePayload) error {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	payloadReader := bytes.NewReader(payloadJson)

	req, err := c.newRequest(http.MethodPost, "/game", payloadReader)
	if err != nil {
		return fmt.Errorf("client.newRequest: %w", err)
	}

	log.Debug("client [InitGame]", "payload", payload)

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	log.Info("client [InitGame]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return fmt.Errorf("checkStatus: %w", err)
	}
	c.token = res.Header.Get("X-Auth-Token")
	log.Debug("client [InitGame]", "token", c.token)
	return nil
}

func (c *Client) GetStatus() (models.StatusData, error) {
	req, err := c.newRequestWithToken(http.MethodGet, "/game", nil)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("client.newRequestWithToken: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetStatus]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data models.StatusData
	err = getFromResponseBody(res, &data)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}

	log.Debug("client [GetStatus]", "status", data)
	return data, nil
}

func (c *Client) GetBoard() (models.Board, error) {
	req, err := c.newRequestWithToken(http.MethodGet, "/game/board", nil)
	if err != nil {
		return models.Board{}, fmt.Errorf("client.newRequestWithToken: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return models.Board{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetBoard]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return models.Board{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data models.Board
	err = getFromResponseBody(res, &data)
	if err != nil {
		return models.Board{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}

	log.Debug("client [GetBoard]", "board", data.Board)
	return data, nil
}

func (c *Client) GetDescription() (models.StatusData, error) {
	req, err := c.newRequestWithToken(http.MethodGet, "/game/desc", nil)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("client.newRequestWithToken: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetDescription]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data models.StatusData
	err = getFromResponseBody(res, &data)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}
	log.Debug("client [GetDescription]", "status", data)
	return data, nil
}

func (c *Client) Fire(coord string) (models.FireAnswer, error) {
	payload := models.FirePayload{Coord: coord}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("json.Marshal: %w", err)
	}

	payloadReader := bytes.NewReader(payloadJson)

	req, err := c.newRequestWithToken(http.MethodPost, "/game/fire", payloadReader)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("client.newRequestWithToken: %w", err)
	}

	log.Debug("client [Fire]", "payload", payload)
	res, err := c.Do(req)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [Fire]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data models.FireAnswer
	err = getFromResponseBody(res, &data)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}

	log.Debug("client [Fire]", "data", data)
	return data, nil
}

func (c *Client) GetPlayersList() ([]models.ListData, error) {
	req, err := c.newRequest(http.MethodGet, "/game/list", nil)
	if err != nil {
		return []models.ListData{}, fmt.Errorf("client.newRequest: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return []models.ListData{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetPlayersList]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return []models.ListData{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data []models.ListData
	err = getFromResponseBody(res, &data)
	if err != nil {
		return []models.ListData{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}

	log.Debug("client [GetPlayersList]", "data", data)
	return data, nil
}

func (c *Client) RefreshSession() error {
	req, err := c.newRequestWithToken(http.MethodGet, "/game/refresh", nil)
	if err != nil {
		return fmt.Errorf("client.newRequestWithToken: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [RefreshSession]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return fmt.Errorf("checkStatus: %w", err)
	}
	return nil
}

func (c *Client) GetStats() (models.StatsList, error) {
	req, err := c.newRequest(http.MethodGet, "/stats", nil)
	if err != nil {
		return models.StatsList{}, fmt.Errorf("client.newRequest: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return models.StatsList{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetStats]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return models.StatsList{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data models.StatsList
	err = getFromResponseBody(res, &data)
	if err != nil {
		return models.StatsList{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}

	log.Debug("client [GetStats]", "data", data)
	return data, nil
}

func (c *Client) GetPlayerStats(nick string) (models.StatsNick, error) {
	req, err := c.newRequest(http.MethodGet, "/stats/"+nick, nil)
	if err != nil {
		return models.StatsNick{}, fmt.Errorf("client.newRequest: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return models.StatsNick{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetPlayerStats]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return models.StatsNick{}, fmt.Errorf("checkStatus: %w", err)
	}

	var data models.StatsNick
	err = getFromResponseBody(res, &data)
	if err != nil {
		return models.StatsNick{}, fmt.Errorf("client.getFromResponseBody: %w", err)
	}

	log.Debug("client [GetPlayerStats]", "data", data)
	return data, nil
}

func (c *Client) AbandonGame() error {
	req, err := c.newRequestWithToken(http.MethodDelete, "/game/abandon", nil)
	if err != nil {
		return fmt.Errorf("client.newRequest: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [AbandonGame]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		return fmt.Errorf("checkStatus: %w", err)
	}

	return nil
}

func getFromResponseBody[T any](res *http.Response, target *T) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	return nil
}

func checkStatus(status int) error {
	switch status {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 503:
		return ErrServiceUnavailable
	default:
		return nil
	}
}
