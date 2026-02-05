package mocks

import (
	"context"

	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDockerClient is a mock implementation for Docker client operations
type MockDockerClient struct {
	mock.Mock
}

// EnsureNetwork mocks the EnsureNetwork method
func (m *MockDockerClient) EnsureNetwork(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// DeploySite mocks the DeploySite method
func (m *MockDockerClient) DeploySite(ctx context.Context, req *models.DeployRequest, dataDir string) (*models.DeployResponse, error) {
	args := m.Called(ctx, req, dataDir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeployResponse), args.Error(1)
}

// GetSiteStatus mocks the GetSiteStatus method
func (m *MockDockerClient) GetSiteStatus(ctx context.Context, siteID uuid.UUID) (*models.SiteStatusResponse, error) {
	args := m.Called(ctx, siteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SiteStatusResponse), args.Error(1)
}

// StopSite mocks the StopSite method
func (m *MockDockerClient) StopSite(ctx context.Context, siteID uuid.UUID) error {
	args := m.Called(ctx, siteID)
	return args.Error(0)
}

// RestartSite mocks the RestartSite method
func (m *MockDockerClient) RestartSite(ctx context.Context, siteID uuid.UUID) error {
	args := m.Called(ctx, siteID)
	return args.Error(0)
}

// DeleteSite mocks the DeleteSite method
func (m *MockDockerClient) DeleteSite(ctx context.Context, siteID uuid.UUID) error {
	args := m.Called(ctx, siteID)
	return args.Error(0)
}

// GetContainerLogs mocks the GetContainerLogs method
func (m *MockDockerClient) GetContainerLogs(ctx context.Context, siteID uuid.UUID, lines int) ([]string, error) {
	args := m.Called(ctx, siteID, lines)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetDockerInfo mocks the GetDockerInfo method
func (m *MockDockerClient) GetDockerInfo(ctx context.Context) (*models.DockerInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DockerInfo), args.Error(1)
}

// CheckPortConflicts mocks the CheckPortConflicts method
func (m *MockDockerClient) CheckPortConflicts(ctx context.Context, hostPorts []int, excludeSiteID uuid.UUID) error {
	args := m.Called(ctx, hostPorts, excludeSiteID)
	return args.Error(0)
}

// PullLatestImage mocks the PullLatestImage method
func (m *MockDockerClient) PullLatestImage(ctx context.Context, image string, credentials models.DockerCredentials) error {
	args := m.Called(ctx, image, credentials)
	return args.Error(0)
}

// RecreateContainer mocks the RecreateContainer method
func (m *MockDockerClient) RecreateContainer(ctx context.Context, siteID uuid.UUID) error {
	args := m.Called(ctx, siteID)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockDockerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
