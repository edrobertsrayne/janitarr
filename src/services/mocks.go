package services

import (
	"context"
	"sync"

	"github.com/edrobertsrayne/janitarr/src/api"
)

// MockDetectorAPIClient is a mock API client for testing Detector.
type MockDetectorAPIClient struct {
	MissingItems []int
	CutoffItems  []int
	Err          error
}

func (m *MockDetectorAPIClient) TestConnection(ctx context.Context) (*api.SystemStatus, error) {
	return nil, nil
}
func (m *MockDetectorAPIClient) GetAllMissing(ctx context.Context) ([]api.MediaItem, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var items []api.MediaItem
	for _, id := range m.MissingItems {
		items = append(items, api.MediaItem{ID: id})
	}
	return items, nil
}
func (m *MockDetectorAPIClient) GetAllCutoffUnmet(ctx context.Context) ([]api.MediaItem, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var items []api.MediaItem
	for _, id := range m.CutoffItems {
		items = append(items, api.MediaItem{ID: id})
	}
	return items, nil
}
func (m *MockDetectorAPIClient) TriggerSearch(ctx context.Context, ids []int) error { return nil }

// MockTriggerAPIClient is a mock API client for testing SearchTrigger.
type MockTriggerAPIClient struct {
	Mu             sync.Mutex
	TriggerCalls   [][]int
	TriggerErr     error
	ServerType     string
	TestConnResult *api.SystemStatus
	TestConnErr    error
	MissingItems   []api.MediaItem
	CutoffItems    []api.MediaItem
}

func (m *MockTriggerAPIClient) TestConnection(ctx context.Context) (*api.SystemStatus, error) {
	if m.TestConnErr != nil {
		return nil, m.TestConnErr
	}
	return m.TestConnResult, nil
}

func (m *MockTriggerAPIClient) GetAllMissing(ctx context.Context) ([]api.MediaItem, error) {
	return m.MissingItems, nil
}

func (m *MockTriggerAPIClient) GetAllCutoffUnmet(ctx context.Context) ([]api.MediaItem, error) {
	return m.CutoffItems, nil
}

func (m *MockTriggerAPIClient) TriggerSearch(ctx context.Context, ids []int) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.TriggerCalls = append(m.TriggerCalls, ids)
	return m.TriggerErr
}

func (m *MockTriggerAPIClient) GetTriggerCalls() [][]int {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	return m.TriggerCalls
}
