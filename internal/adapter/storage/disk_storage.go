package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DiskStorage struct {
	basePath string
}

func NewDiskStorage(basePath string) *DiskStorage {
	return &DiskStorage{basePath: basePath}
}

func (ds *DiskStorage) Save(filename string, data []byte) (string, error) {
	path := filepath.Join(ds.basePath, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	return path, nil
}

func (ds *DiskStorage) Delete(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (ds *DiskStorage) Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

func (ds *DiskStorage) List(folder string) (map[string]string, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	result := make(map[string]string)
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pdf") {
			name := strings.TrimSuffix(f.Name(), ".pdf")
			result[name] = filepath.Join(folder, f.Name())
		}
	}

	return result, nil
}
