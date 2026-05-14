package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ttsRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id,omitempty"`
}

func main() {
	text := flag.String("text", "", "Text to convert to speech")
	modelID := flag.String("model-id", "eleven_multilingual_v2", "ElevenLabs TTS model ID")
	outputFile := flag.String("out", "output.mp3", "Output audio file path")
	outputFormat := flag.String("format", "mp3_44100_128", "ElevenLabs output format")
	baseURL := flag.String("base-url", "https://api.elevenlabs.io", "ElevenLabs API base URL")
	flag.Parse()

	if strings.TrimSpace(*text) == "" {
		exitWithError(errors.New("-text is required"))
	}

	envValues, err := readDotEnv(".env")
	if err != nil {
		exitWithError(err)
	}

	apiKey := getConfigValue("API_KEY", envValues)
	if apiKey == "" {
		exitWithError(errors.New("API_KEY is missing from environment and .env"))
	}

	voiceID := getConfigValue("VOICE_ID", envValues)
	if voiceID == "" {
		exitWithError(errors.New("VOICE_ID is missing from environment and .env"))
	}

	audio, err := synthesizeSpeech(*baseURL, apiKey, voiceID, *modelID, *text, *outputFormat)
	if err != nil {
		exitWithError(err)
	}

	if err := os.WriteFile(*outputFile, audio, 0o644); err != nil {
		exitWithError(fmt.Errorf("write output file: %w", err))
	}

	fmt.Printf("Saved audio to %s\n", *outputFile)
	if absPath, err := filepath.Abs(*outputFile); err == nil {
		fmt.Printf("Absolute path: %s\n", absPath)
	}
}

func synthesizeSpeech(baseURL, apiKey, voiceID, modelID, text, outputFormat string) ([]byte, error) {
	endpoint, err := url.Parse(strings.TrimRight(baseURL, "/") + "/v1/text-to-speech/" + voiceID)
	if err != nil {
		return nil, fmt.Errorf("build TTS URL: %w", err)
	}

	query := endpoint.Query()
	query.Set("output_format", outputFormat)
	endpoint.RawQuery = query.Encode()

	payload, err := json.Marshal(ttsRequest{
		Text:    text,
		ModelID: modelID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal TTS payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create TTS request: %w", err)
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg, application/octet-stream")

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send TTS request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read TTS response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS request failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func readDotEnv(path string) (map[string]string, error) {
	values := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return values, nil
}

func getConfigValue(key string, envValues map[string]string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return strings.TrimSpace(envValues[key])
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
