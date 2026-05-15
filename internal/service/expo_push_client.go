package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultExpoPushEndpoint = "https://exp.host/--/api/v2/push/send"

type ExpoPushClient struct {
	httpClient  *http.Client
	endpoint    string
	accessToken string
}

type ExpoPushMessage struct {
	To    string         `json:"to"`
	Title string         `json:"title"`
	Body  string         `json:"body,omitempty"`
	Sound string         `json:"sound,omitempty"`
	Data  map[string]any `json:"data,omitempty"`
}

type ExpoPushTicket struct {
	Token  string
	Status string
	Error  string
}

type expoPushTicketResponse struct {
	Data []struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Details struct {
			Error string `json:"error"`
		} `json:"details"`
	} `json:"data"`
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func NewExpoPushClientFromEnv() *ExpoPushClient {
	endpoint := strings.TrimSpace(os.Getenv("EXPO_PUSH_ENDPOINT"))
	if endpoint == "" {
		endpoint = defaultExpoPushEndpoint
	}

	return &ExpoPushClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint:    endpoint,
		accessToken: strings.TrimSpace(os.Getenv("EXPO_ACCESS_TOKEN")),
	}
}

func (c *ExpoPushClient) Send(ctx context.Context, messages []ExpoPushMessage) ([]ExpoPushTicket, error) {
	if c == nil || len(messages) == 0 {
		return nil, nil
	}

	tickets := make([]ExpoPushTicket, 0, len(messages))
	for start := 0; start < len(messages); start += 100 {
		end := start + 100
		if end > len(messages) {
			end = len(messages)
		}

		chunkTickets, err := c.sendChunk(ctx, messages[start:end])
		if err != nil {
			return tickets, err
		}
		tickets = append(tickets, chunkTickets...)
	}

	return tickets, nil
}

func (c *ExpoPushClient) sendChunk(ctx context.Context, messages []ExpoPushMessage) ([]ExpoPushTicket, error) {
	body, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var decoded expoPushTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("expo push retornou status %d", resp.StatusCode)
	}

	if len(decoded.Errors) > 0 {
		return nil, fmt.Errorf("expo push retornou erro: %s", decoded.Errors[0].Message)
	}

	tickets := make([]ExpoPushTicket, 0, len(messages))
	for i, message := range messages {
		ticket := ExpoPushTicket{
			Token:  message.To,
			Status: "unknown",
		}

		if i < len(decoded.Data) {
			ticket.Status = decoded.Data[i].Status
			ticket.Error = decoded.Data[i].Details.Error
			if ticket.Error == "" {
				ticket.Error = decoded.Data[i].Message
			}
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}
