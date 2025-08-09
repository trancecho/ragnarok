package fastgpt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type FastGPTService struct {
	apiKey     string
	apiBaseURL string
	httpClient *http.Client
}

// NewFastGPTService creates a new FastGPT service client
func NewFastGPTService(apiKey string, apiBaseURL string) *FastGPTService {
	return &FastGPTService{
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// FastGPTChat sends a chat request to FastGPT and returns the response
func (s *FastGPTService) FastGPTChat(req Request) ([]byte, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", s.apiBaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from FastGPT API: %s, status: %d", string(bodyBytes), resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return bodyBytes, nil
}

// FastGPTStreamChat sends a streaming chat request to FastGPT and processes chunks with the callback
func (s *FastGPTService) FastGPTStreamChat(req Request, callback func(data []byte)) error {
	// Ensure stream is enabled
	req.Stream = true

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", s.apiBaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error response from FastGPT API: %s, status: %d", string(bodyBytes), resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		// Skip empty lines
		if len(line) <= 2 {
			continue
		}

		// Check for data prefix
		data := bytes.TrimSpace(line)
		if !bytes.HasPrefix(data, []byte("data: ")) {
			continue
		}

		// Extract the data part
		data = bytes.TrimPrefix(data, []byte("data: "))

		// Check for stream end
		if string(data) == "[DONE]" {
			break
		}

		// Process the data chunk
		callback(data)
	}

	return nil
}
