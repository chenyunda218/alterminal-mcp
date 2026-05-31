// Package client provides types and HTTP access for the Alterminal Actor API.
// See swagger.yaml for the OpenAPI specification.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseUrl   string
	ActorId   string
	SecretKey string
}

func New(baseUrl, actorId, secretKey string) *Client {
	return &Client{
		BaseUrl:   baseUrl,
		ActorId:   actorId,
		SecretKey: secretKey,
	}
}

// GetMessages returns the actor's conversation messages ordered by position.
func (c *Client) GetMessages() (*MessagesResponse, error) {
	url := strings.TrimRight(c.BaseUrl, "/") + "/api/actors/" + c.ActorId + "/messages"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("client: get messages: %w", err)
	}
	req.Header.Set("secret-key", c.SecretKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: get messages: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("client: get messages: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("client: get messages: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("client: get messages: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out MessagesResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("client: get messages: decode response: %w", err)
	}
	return &out, nil
}

// UpdateTools replaces all tool definitions for the actor.
func (c *Client) UpdateTools(tools ToolsUpdateBody) error {
	url := strings.TrimRight(c.BaseUrl, "/") + "/api/actors/" + c.ActorId + "/tools"

	payload, err := json.Marshal(tools)
	if err != nil {
		return fmt.Errorf("client: update tools: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("client: update tools: %w", err)
	}
	req.Header.Set("secret-key", c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("client: update tools: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("client: update tools: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return fmt.Errorf("client: update tools: %s", errResp.Error.Message)
		}
		return fmt.Errorf("client: update tools: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var out SuccessResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("client: update tools: decode response: %w", err)
	}
	if !out.Success {
		return fmt.Errorf("client: update tools: success=false")
	}
	return nil
}
