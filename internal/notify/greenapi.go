package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func dispatchGreenAPI(ctx context.Context, configJSON, msg string) error {
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return fmt.Errorf("greenapi: parse config: %w", err)
	}
	instanceID := cfg["instance_id"]
	apiToken := cfg["api_token_instance"]
	chatID := cfg["chat_id"]
	apiURL := cfg["api_url"]
	if apiURL == "" {
		apiURL = "https://api.green-api.com"
	}
	if instanceID == "" || apiToken == "" || chatID == "" {
		return fmt.Errorf("greenapi: missing instance_id, api_token_instance, or chat_id")
	}
	url := fmt.Sprintf("%s/waInstance%s/sendMessage/%s", apiURL, instanceID, apiToken)
	payload := map[string]any{"chatId": chatID, "message": msg}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("greenapi: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("greenapi: status %d", resp.StatusCode)
	}
	return nil
}
