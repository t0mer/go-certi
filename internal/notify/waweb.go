package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func dispatchWaWeb(ctx context.Context, configJSON, msg string) error {
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return fmt.Errorf("waweb: parse config: %w", err)
	}
	baseURL := cfg["base_url"]
	phone := cfg["phone"]
	authHeader := cfg["auth"]
	if baseURL == "" || phone == "" {
		return fmt.Errorf("waweb: missing base_url or phone in config")
	}
	url := fmt.Sprintf("%s/api/send/message", baseURL)
	payload := map[string]any{"phone": phone, "message": msg}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("waweb: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("waweb: status %d", resp.StatusCode)
	}
	return nil
}
