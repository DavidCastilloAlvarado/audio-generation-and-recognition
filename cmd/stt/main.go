package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type sttResponse struct {
	Text string `json:"text"`
}

func main() {
	inputFile := flag.String("in", "", "Input audio or video file path")
	inputURL := flag.String("url", "", "Public HTTPS audio or video URL passed directly to ElevenLabs")
	modelID := flag.String("model-id", "scribe_v2", "ElevenLabs STT model ID")
	raw := flag.Bool("raw", false, "Print raw JSON response")
	baseURL := flag.String("base-url", "https://api.elevenlabs.io", "ElevenLabs API base URL")
	flag.Parse()

	hasFile := strings.TrimSpace(*inputFile) != ""
	hasURL := strings.TrimSpace(*inputURL) != ""
	if hasFile == hasURL {
		exitWithError(errors.New("provide exactly one of -in or -url"))
	}

	if hasURL {
		if err := validateHTTPSURL(*inputURL); err != nil {
			exitWithError(err)
		}
	}

	envValues, err := readDotEnv(".env")
	if err != nil {
		exitWithError(err)
	}

	apiKey := getConfigValue("API_KEY", envValues)
	if apiKey == "" {
		exitWithError(errors.New("API_KEY is missing from environment and .env"))
	}

	responseBody, err := transcribe(*baseURL, apiKey, *modelID, *inputFile, *inputURL)
	if err != nil {
		exitWithError(err)
	}

	if *raw {
		fmt.Println(string(responseBody))
		return
	}

	var transcript sttResponse
	if err := json.Unmarshal(responseBody, &transcript); err == nil && strings.TrimSpace(transcript.Text) != "" {
		fmt.Println(transcript.Text)
		return
	}

	prettyJSON := &bytes.Buffer{}
	if err := json.Indent(prettyJSON, responseBody, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
		return
	}

	fmt.Println(string(responseBody))
}

func transcribe(baseURL, apiKey, modelID, inputFile, inputURL string) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("model_id", modelID); err != nil {
		return nil, fmt.Errorf("write model_id field: %w", err)
	}

	if strings.TrimSpace(inputFile) != "" {
		if err := addFileField(writer, inputFile); err != nil {
			return nil, err
		}
	} else {
		if err := writer.WriteField("cloud_storage_url", strings.TrimSpace(inputURL)); err != nil {
			return nil, fmt.Errorf("write cloud_storage_url field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/speech-to-text", body)
	if err != nil {
		return nil, fmt.Errorf("create STT request: %w", err)
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send STT request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read STT response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("STT request failed: %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}

	return responseBody, nil
}

func addFileField(writer *multipart.Writer, inputFile string) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(inputFile))
	if err != nil {
		return fmt.Errorf("create file form field: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy input file into request: %w", err)
	}

	return nil
}

func validateHTTPSURL(rawURL string) error {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if !parsedURL.IsAbs() || !strings.EqualFold(parsedURL.Scheme, "https") || parsedURL.Host == "" {
		return errors.New("-url must be a public HTTPS URL")
	}

	return nil
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
