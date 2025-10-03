package bridge

import (
	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/types/ipfs"
)

// MockIPFSClient provides a test implementation of IPFSClient
type MockIPFSClient struct{}

func (m *MockIPFSClient) Add(data []byte) (string, error) {
	return "mock-cid", nil
}

func (m *MockIPFSClient) AddFile(file ipfs.File) (string, error) {
	return "mock-file-cid", nil
}

func (m *MockIPFSClient) AddFolder(folder ipfs.Folder) (string, error) {
	return "mock-folder-cid", nil
}

func (m *MockIPFSClient) Get(cid string) ([]byte, error) {
	return []byte("mock-ipfs-data"), nil
}

func (m *MockIPFSClient) GetFile(cid string) (ipfs.File, error) {
	return nil, nil
}

func (m *MockIPFSClient) GetFolder(cid string) (ipfs.Folder, error) {
	return nil, nil
}

func (m *MockIPFSClient) Pin(cid string, name string) error {
	return nil
}

func (m *MockIPFSClient) Unpin(cid string) error {
	return nil
}

func (m *MockIPFSClient) Exists(cid string) (bool, error) {
	return true, nil
}

func (m *MockIPFSClient) IsPinned(ipns string) (bool, error) {
	return true, nil
}

func (m *MockIPFSClient) Ls(cid string) ([]string, error) {
	return []string{"mock-file1", "mock-file2"}, nil
}

func (m *MockIPFSClient) NodeStatus() (*ipfs.NodeStatus, error) {
	return &ipfs.NodeStatus{
		PeerID:         "mock-peer-id",
		Version:        "mock-version",
		PeerType:       "kubo",
		ConnectedPeers: 5,
	}, nil
}

// AsynqClientInterface defines the interface we need for testing
type AsynqClientInterface interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	Close() error
}

// MockAsynqClient provides a test double for asynq.Client
type MockAsynqClient struct {
	enqueuedTasks []MockTask
}

type MockTask struct {
	Type    string
	Payload []byte
	Queue   string
}

func (m *MockAsynqClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	mockTask := MockTask{
		Type:    task.Type(),
		Payload: task.Payload(),
		Queue:   "default", // Default queue
	}

	// For testing purposes, we'll just use the default queue
	// In real implementation, options would be parsed properly
	m.enqueuedTasks = append(m.enqueuedTasks, mockTask)

	return &asynq.TaskInfo{
		ID:    "test-task-id",
		Type:  task.Type(),
		Queue: mockTask.Queue,
	}, nil
}

func (m *MockAsynqClient) Close() error {
	return nil
}
