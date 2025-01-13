package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Netflix/go-env"
)

type SlackClient struct {
	Token   string
	Channel string
	Enabled bool
}

type slackConfig struct {
	SlackChannel string `env:"AUTODEPLOY_SLACK_CHANNEL" default:""`
	SlackToken   string `env:"AUTODEPLOY_SLACK_TOKEN" default:""`
}

func New() *SlackClient {
	config := slackConfig{}
	if _, err := env.UnmarshalFromEnviron(&config); err != nil {
		return nil
	}
	enabled := true
	if config.SlackChannel == "" || config.SlackToken == "" {
		enabled = false
	}
	return &SlackClient{
		Token:   config.SlackToken,
		Channel: config.SlackChannel,
		Enabled: enabled,
	}
}

func (s *SlackClient) SendToSlack(initialMsg string, followUpMsg []string) error {
	if s == nil || !s.Enabled {
		return nil
	}
	payload := map[string]string{
		"channel": s.Channel,
		"text":    initialMsg,
	}
	response, err := s.send(payload)
	if err != nil || !response.Ok {
		return fmt.Errorf("failed to send initial message: %w: %+v", err, response)
	}
	for _, msg := range followUpMsg {
		payload = map[string]string{
			"channel":   s.Channel,
			"text":      msg,
			"thread_ts": response.Ts,
		}
		_, err := s.send(payload)
		if err != nil || !response.Ok {
			return fmt.Errorf("failed to send follow up message: %w: %+v", err, response)
		}
	}
	return nil
}

func (s *SlackClient) send(payload map[string]string) (*SlackResponse, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to post message")
	}
	var response SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse Slack response: %v", err)
	}
	return &response, nil
}

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	Ts    string `json:"ts"`
	Error string `json:"error,omitempty"`
}
