package mocks

import (
	"github.com/BlueBeard63/archon/internal/api"
	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockNodeClient is a mock implementation of the NodeClient interface
type MockNodeClient struct {
	mock.Mock
}

// DeploySite mocks the DeploySite method
func (m *MockNodeClient) DeploySite(endpoint, apiKey string, site *models.Site, domainName string) error {
	args := m.Called(endpoint, apiKey, site, domainName)
	return args.Error(0)
}

// DeploySiteWebSocket mocks the DeploySiteWebSocket method
func (m *MockNodeClient) DeploySiteWebSocket(endpoint, apiKey string, site *models.Site, domainName string, progressCallback api.DeploymentProgressCallback) error {
	args := m.Called(endpoint, apiKey, site, domainName, progressCallback)
	return args.Error(0)
}

// DeleteSite mocks the DeleteSite method
func (m *MockNodeClient) DeleteSite(endpoint, apiKey string, siteID uuid.UUID, domain, siteName string, siteType models.SiteType) error {
	args := m.Called(endpoint, apiKey, siteID, domain, siteName, siteType)
	return args.Error(0)
}

// GetSiteStatus mocks the GetSiteStatus method
func (m *MockNodeClient) GetSiteStatus(endpoint, apiKey string, siteID uuid.UUID, siteName string, siteType models.SiteType) (*models.SiteStatus, error) {
	args := m.Called(endpoint, apiKey, siteID, siteName, siteType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SiteStatus), args.Error(1)
}

// StopSite mocks the StopSite method
func (m *MockNodeClient) StopSite(endpoint, apiKey string, siteID uuid.UUID, siteName string, siteType models.SiteType) error {
	args := m.Called(endpoint, apiKey, siteID, siteName, siteType)
	return args.Error(0)
}

// RestartSite mocks the RestartSite method
func (m *MockNodeClient) RestartSite(endpoint, apiKey string, siteID uuid.UUID) error {
	args := m.Called(endpoint, apiKey, siteID)
	return args.Error(0)
}

// HealthCheck mocks the HealthCheck method
func (m *MockNodeClient) HealthCheck(endpoint, apiKey string) (*api.HealthResponse, error) {
	args := m.Called(endpoint, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.HealthResponse), args.Error(1)
}

// GetDockerInfo mocks the GetDockerInfo method
func (m *MockNodeClient) GetDockerInfo(endpoint, apiKey string) (*models.DockerInfo, error) {
	args := m.Called(endpoint, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DockerInfo), args.Error(1)
}

// GetTraefikInfo mocks the GetTraefikInfo method
func (m *MockNodeClient) GetTraefikInfo(endpoint, apiKey string) (*models.TraefikInfo, error) {
	args := m.Called(endpoint, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TraefikInfo), args.Error(1)
}

// GetContainerLogs mocks the GetContainerLogs method
func (m *MockNodeClient) GetContainerLogs(endpoint, apiKey string, siteID uuid.UUID, lines int) ([]string, error) {
	args := m.Called(endpoint, apiKey, siteID, lines)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetContainerMetrics mocks the GetContainerMetrics method
func (m *MockNodeClient) GetContainerMetrics(endpoint, apiKey string, siteID uuid.UUID) (*api.ContainerMetrics, error) {
	args := m.Called(endpoint, apiKey, siteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ContainerMetrics), args.Error(1)
}

// UpdateSite mocks the UpdateSite method
func (m *MockNodeClient) UpdateSite(endpoint, apiKey string, siteID uuid.UUID) error {
	args := m.Called(endpoint, apiKey, siteID)
	return args.Error(0)
}
