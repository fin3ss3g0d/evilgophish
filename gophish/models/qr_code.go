package models

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/skip2/go-qrcode"
)

// generateQRCode generates a QR code with the given content and size, saves it to a temporary file,
// encodes the image to a base64 string, and then deletes the temporary file.
func generateQRCode(content string, stringSize string) (string, string, error) {
	// Set the temporary directory
	tempDir := "./temp_qr_codes"

	// Ensure the temporary directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err := os.MkdirAll(tempDir, os.ModePerm)
		if err != nil {
			return "", "", fmt.Errorf("failed to create temporary directory: %v", err)
		}
	}

	// Generate the QR code
	qrCode, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %v", err)
	}
	qrCode.DisableBorder = true

	// Create a temporary file for the QR code. The pattern "qr-*.png" ensures a unique filename for each QR code.
	tempFile, err := ioutil.TempFile(tempDir, "*.png")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	tempFilePath := tempFile.Name()
	tempFileName := filepath.Base(tempFilePath) // Extract just the filename
	tempFile.Close()                            // Close the file so it can be removed after reading

	// Convert the size to an int
	size, err := strconv.Atoi(stringSize)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert QR code size to int: %v", err)
	}

	// Write the QR code to the file
	err = qrCode.WriteFile(size, tempFilePath)
	if err != nil {
		os.Remove(tempFilePath) // Attempt to remove the file in case of error
		return "", "", fmt.Errorf("failed to write QR code to file: %v", err)
	}

	// Read the file back to get the byte slice
	qrCodeBytes, err := ioutil.ReadFile(tempFilePath)
	if err != nil {
		os.Remove(tempFilePath) // Clean up
		return "", "", fmt.Errorf("failed to read temporary QR code file: %v", err)
	}

	// Delete the temporary file
	err = os.Remove(tempFilePath)
	if err != nil {
		// Log the error but proceed
		fmt.Println("Warning: Failed to delete temporary QR code file:", err)
	}

	// Encode the byte slice to a base64 string
	base64QRCode := base64.StdEncoding.EncodeToString(qrCodeBytes)

	return base64QRCode, tempFileName, nil
}
