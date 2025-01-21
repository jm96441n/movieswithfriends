package web

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"sync"
)

// AssetManifest stores the mapping between original and fingerprinted filenames
type AssetManifest struct {
	Assets map[string]string `json:"assets"`
}

// Loader handles loading and resolving fingerprinted asset paths
type Loader struct {
	manifest AssetManifest
	mu       sync.RWMutex
}

// NewLoader creates a new asset loader from a manifest file
func NewLoader(filesys fs.FS, manifestPath string) (*Loader, error) {
	file, err := filesys.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var manifest AssetManifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, err
	}

	return &Loader{
		manifest: manifest,
	}, nil
}

// Path returns the fingerprinted path for a given asset
func (l *Loader) Path(assetPath string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if fingerprinted, ok := l.manifest.Assets[assetPath]; ok {
		return filepath.Join("/dist", fingerprinted)
	}
	return filepath.Join("/dist", assetPath)
}
