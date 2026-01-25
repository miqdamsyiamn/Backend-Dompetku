package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ImageKitResponse struct {
	FileID      string `json:"fileId"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
	Size        int    `json:"size"`
}

type ImageKitErrorResponse struct {
	Message string `json:"message"`
}

// UploadToImageKit uploads a file to ImageKit and returns the URL
func UploadToImageKit(file multipart.File, fileHeader *multipart.FileHeader, folder string) (string, error) {
	privateKey := os.Getenv("IMAGEKIT_PRIVATE_KEY")
	urlEndpoint := os.Getenv("IMAGEKIT_URL_ENDPOINT")
	
	if privateKey == "" || urlEndpoint == "" {
		return "", fmt.Errorf("ImageKit configuration not found")
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Create form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file as base64
	base64File := base64.StdEncoding.EncodeToString(fileBytes)
	writer.WriteField("file", base64File)

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("profile_%d%s", time.Now().UnixNano(), ext)
	writer.WriteField("fileName", filename)

	// Set folder
	if folder != "" {
		writer.WriteField("folder", folder)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://upload.imagekit.io/api/v1/files/upload", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Set basic auth with private key
	auth := base64.StdEncoding.EncodeToString([]byte(privateKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ImageKitErrorResponse
		json.Unmarshal(respBody, &errResp)
		return "", fmt.Errorf("upload failed: %s", errResp.Message)
	}

	var result ImageKitResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return result.URL, nil
}
