package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPExecutor struct {
	Client *http.Client
}

func NewHTTPExecutor(timeout time.Duration) *HTTPExecutor {
	return &HTTPExecutor{Client: &http.Client{Timeout: timeout}}
}

func (e *HTTPExecutor) Execute(ip, command, hash string) (*VMResponse, error) {

	url := fmt.Sprintf("http://%s:8080/run-command", ip)

	payload := map[string]string{
		"command": command,
		"hash":    hash,
	}

	body, _ := json.Marshal(payload)
	resp, err := e.Client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	return &VMResponse{Output: string(b)}, nil
}

type VMResponse struct {
	Output string
}
