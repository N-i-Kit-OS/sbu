package storage

import "context"

type MockStorage struct {
	data map[string][]byte
}

func (m *MockStorage) PutObject(ctx context.Context, bucketName, objectName string, data []byte) error {
	m.data[objectName] = data
	return nil
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}
