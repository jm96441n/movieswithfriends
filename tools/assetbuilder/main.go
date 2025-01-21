package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// AssetManifest stores the mapping between original and fingerprinted filenames
type AssetManifest struct {
	Assets map[string]string `json:"assets"`
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}
	uiPath := getUIPath(wd)
	// Configuration - could be moved to flags or config file
	sourceDir := filepath.Join(uiPath, "/static")
	outputDir := filepath.Join(uiPath, "/dist")

	manifestPath := filepath.Join(uiPath, "/dist/manifest.json")

	manifest := AssetManifest{
		Assets: make(map[string]string),
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Walk through all files in source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Calculate file hash
		hash, err := calculateFileHash(path)
		if err != nil {
			return fmt.Errorf("failed to calculate hash for %s: %v", path, err)
		}

		// Create fingerprinted filename
		ext := filepath.Ext(relPath)
		baseWithoutExt := strings.TrimSuffix(relPath, ext)
		fingerprintedName := fmt.Sprintf("%s-%s%s", baseWithoutExt, hash[:8], ext)

		// Create output path
		outputPath := filepath.Join(outputDir, fingerprintedName)
		outputDir := filepath.Dir(outputPath)

		// Ensure output directory exists
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", outputDir, err)
		}

		// Copy file to output directory
		if err := copyFile(path, outputPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %v", path, err)
		}

		// Add to manifest
		manifest.Assets[relPath] = fingerprintedName
		log.Printf("Processed: %s -> %s", relPath, fingerprintedName)

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to process assets: %v", err)
	}

	// Write manifest file
	manifestFile, err := os.Create(manifestPath)
	if err != nil {
		log.Fatalf("Failed to create manifest file: %v", err)
	}
	defer manifestFile.Close()

	encoder := json.NewEncoder(manifestFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		log.Fatalf("Failed to write manifest: %v", err)
	}

	log.Printf("Asset manifest written to: %s", manifestPath)
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// recurse up the current path until you hit the root of the module which is where the migrations are stored
func getUIPath(path string) string {
	_, err := os.Stat(filepath.Join(path, "go.mod"))

	if err == nil {
		return filepath.Join(path, "ui")
	}

	if path == "/" {
		panic("went too far!")
	}

	return getUIPath(filepath.Dir(path))
}
