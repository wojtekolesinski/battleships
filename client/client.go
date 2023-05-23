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

var ErrUnauthorized = fmt.Errorf("unauthorized")
var ErrForbidden = fmt.Errorf("forbidden")
var ErrServiceUnavailable = fmt.Errorf("service unavailable")

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

func (c *Client) InitGame(nick, desc, targetNick string, wpbot bool) error {
	payload := models.GamePayload{Desc: desc, Nick: nick, TargetNick: targetNick, Wpbot: wpbot}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	payloadReader := bytes.NewReader(payloadJson)

	req, err := c.newRequest(http.MethodPost, "/game", payloadReader)
	if err != nil {
		return fmt.Errorf("client.newRequest: %w", err)
	}

	log.Info("client [InitGame]", "payload", payload)

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	log.Info("client [InitGame]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		log.Error("client [InitGame]", "client", fmt.Sprintf("%v", c))
		return fmt.Errorf("checkStatus: %w", err)
	}
	c.token = res.Header.Get("X-Auth-Token")
	log.Info("client [InitGame]", "token", c.token)
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
		log.Error("client [GetStatus]", "client", fmt.Sprintf("%v", c))
		return models.StatusData{}, fmt.Errorf("checkStatus: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("io.ReadAll: %w", err)
	}
	var status models.StatusData
	err = json.Unmarshal(body, &status)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	log.Info("client [GetStatus]", "status", status)
	return status, nil
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
		log.Error("client [GetBoard]", "client", fmt.Sprintf("%v", c))
		return models.Board{}, fmt.Errorf("checkStatus: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Board{}, fmt.Errorf("io.ReadAll: %w", err)
	}
	var board models.Board
	err = json.Unmarshal(body, &board)
	if err != nil {
		return models.Board{}, fmt.Errorf("json.Unmarshal: %w", err)
	}

	log.Info("client [GetBoard]", "board", board.Board)
	return board, nil
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
		log.Error("client [GetDescription]", "client", fmt.Sprintf("%v", c))
		return models.StatusData{}, fmt.Errorf("checkStatus: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("io.ReadAll: %w", err)
	}

	var status models.StatusData
	err = json.Unmarshal(body, &status)
	if err != nil {
		return models.StatusData{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	log.Info("client [GetDescription]", "status", status)
	return status, nil
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

	log.Info("client [Fire]", "payload", payload)

	res, err := c.Do(req)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [Fire]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		log.Error("client [Fire]", "client", fmt.Sprintf("%v", c))
		return models.FireAnswer{}, fmt.Errorf("checkStatus: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("io.ReadAll: %w", err)
	}

	var answer models.FireAnswer
	err = json.Unmarshal(body, &answer)
	if err != nil {
		return models.FireAnswer{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	log.Info("client [Fire]", "answer", answer)

	return answer, nil
}

func (c *Client) GetPlayersList() (models.ListData, error) {
	req, err := c.newRequest(http.MethodGet, "/game/list", nil)
	if err != nil {
		return models.ListData{}, fmt.Errorf("client.newRequest: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return models.ListData{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	log.Info("client [GetPlayersList]", "statusCode", res.StatusCode)
	err = checkStatus(res.StatusCode)
	if err != nil {
		log.Error("client [GetPlayersList]", "client", fmt.Sprintf("%v", c))
		return models.ListData{}, fmt.Errorf("checkStatus: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.ListData{}, fmt.Errorf("io.ReadAll: %w", err)
	}
	var data models.ListData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return models.ListData{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	log.Info("client [GetPlayersList]", "data", data)
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
		log.Error("client [RefreshSession]", "client", fmt.Sprintf("%v", c))
		return fmt.Errorf("checkStatus: %w", err)
	}
	return nil
}

func checkStatus(status int) error {
	switch status {
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 503:
		return ErrServiceUnavailable
	default:
		return nil
	}
}
