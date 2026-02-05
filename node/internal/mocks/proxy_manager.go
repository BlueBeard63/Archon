package mocks

import (
	"context"

	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockProxyManager is a mock implementation of the ProxyManager interface
type MockProxyManager struct {
	mock.Mock
}

// ConfigureForValidation mocks the ConfigureForValidation method
func (m *MockProxyManager) ConfigureForValidation(ctx context.Context, site *models.DeployRequest) error {
	args := m.Called(ctx, site)
	return args.Error(0)
}

// Configure mocks the Configure method
func (m *MockProxyManager) Configure(ctx context.Context, site *models.DeployRequest, certPath, keyPath string) error {
	args := m.Called(ctx, site, certPath, keyPath)
	return args.Error(0)
}

// Remove mocks the Remove method
func (m *MockProxyManager) Remove(ctx context.Context, siteID uuid.UUID, domain string) error {
	args := m.Called(ctx, siteID, domain)
	return args.Error(0)
}

// Reload mocks the Reload method
func (m *MockProxyManager) Reload(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// GetInfo mocks the GetInfo method
func (m *MockProxyManager) GetInfo(ctx context.Context) (*models.TraefikInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TraefikInfo), args.Error(1)
}
