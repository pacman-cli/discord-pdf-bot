package port

type StoragePort interface {
	Save(filename string, data []byte) (string, error)
	Delete(path string) error
	Read(path string) ([]byte, error)
	List(folder string) (map[string]string, error)
}
