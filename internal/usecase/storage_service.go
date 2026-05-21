package usecase

import "discord-pdf-bot/internal/domain/port"

type StorageService struct {
	storage port.StoragePort
}

func NewStorageService(storage port.StoragePort) *StorageService {
	return &StorageService{storage: storage}
}

func (s *StorageService) Save(filename string, data []byte) (string, error) {
	return s.storage.Save(filename, data)
}

func (s *StorageService) Delete(path string) error {
	return s.storage.Delete(path)
}

func (s *StorageService) Read(path string) ([]byte, error) {
	return s.storage.Read(path)
}

func (s *StorageService) List(folder string) (map[string]string, error) {
	return s.storage.List(folder)
}
