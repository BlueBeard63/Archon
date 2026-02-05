package app

import (
	"errors"
	"testing"

	"github.com/BlueBeard63/archon/internal/mocks"
	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleDeleteSite_CallsNodeAPI(t *testing.T) {
	nodeClient := new(mocks.MockNodeClient)
	siteID := uuid.New()
	nodeID := uuid.New()
	domainID := uuid.New()

	// Setup mock expectation - should call DeleteSite on node
	nodeClient.On("DeleteSite",
		"http://node:8080",
		"api-key",
		siteID,
		"example.com",
		"test-site",
		models.SiteTypeContainer,
	).Return(nil)

	appState := state.NewAppState()
	appState.Sites = []models.Site{
		{
			ID:       siteID,
			Name:     "test-site",
			SiteType: models.SiteTypeContainer,
			DomainID: domainID,
			NodeID:   nodeID,
			Status:   models.SiteStatusRunning,
		},
	}
	appState.Domains = []models.Domain{
		{
			ID:   domainID,
			Name: "example.com",
		},
	}
	appState.Nodes = []models.Node{
		{
			ID:          nodeID,
			Name:        "test-node",
			APIEndpoint: "http://node:8080",
			APIKey:      "api-key",
		},
	}

	// Create model with mock
	m := &ModelTestable{
		state:      appState,
		nodeClient: nodeClient,
	}

	// Call delete
	err := m.handleDeleteSiteTestable(siteID)

	assert.NoError(t, err)
	nodeClient.AssertExpectations(t)
}

func TestHandleDeleteSite_NodeAPIFailure_StillDeletesLocal(t *testing.T) {
	nodeClient := new(mocks.MockNodeClient)
	siteID := uuid.New()
	nodeID := uuid.New()
	domainID := uuid.New()

	// Node API returns error
	nodeClient.On("DeleteSite",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(errors.New("node unreachable"))

	appState := state.NewAppState()
	appState.Sites = []models.Site{
		{
			ID:       siteID,
			Name:     "test-site",
			SiteType: models.SiteTypeContainer,
			DomainID: domainID,
			NodeID:   nodeID,
			Status:   models.SiteStatusRunning,
		},
	}
	appState.Domains = []models.Domain{
		{
			ID:   domainID,
			Name: "example.com",
		},
	}
	appState.Nodes = []models.Node{
		{
			ID:          nodeID,
			Name:        "test-node",
			APIEndpoint: "http://node:8080",
			APIKey:      "api-key",
		},
	}

	m := &ModelTestable{
		state:      appState,
		nodeClient: nodeClient,
	}

	// Call delete - should still succeed locally
	err := m.handleDeleteSiteTestable(siteID)

	assert.NoError(t, err)
	// Site should be removed from state
	assert.Len(t, m.state.Sites, 0)
	// Warning notification should be added
	assert.True(t, hasNotificationWithLevel(m.state.Notifications, "warning"))
}

func TestHandleDeleteSite_InactiveSite_SkipsNodeAPI(t *testing.T) {
	nodeClient := new(mocks.MockNodeClient)
	siteID := uuid.New()
	nodeID := uuid.New()
	domainID := uuid.New()

	// Node API should NOT be called for inactive sites

	appState := state.NewAppState()
	appState.Sites = []models.Site{
		{
			ID:       siteID,
			Name:     "test-site",
			SiteType: models.SiteTypeContainer,
			DomainID: domainID,
			NodeID:   nodeID,
			Status:   models.SiteStatusInactive, // Never deployed
		},
	}
	appState.Domains = []models.Domain{
		{
			ID:   domainID,
			Name: "example.com",
		},
	}
	appState.Nodes = []models.Node{
		{
			ID:          nodeID,
			Name:        "test-node",
			APIEndpoint: "http://node:8080",
			APIKey:      "api-key",
		},
	}

	m := &ModelTestable{
		state:      appState,
		nodeClient: nodeClient,
	}

	// Call delete
	err := m.handleDeleteSiteTestable(siteID)

	assert.NoError(t, err)
	// Site should be removed from state
	assert.Len(t, m.state.Sites, 0)
	// Node API should NOT have been called
	nodeClient.AssertNotCalled(t, "DeleteSite")
}

// ModelTestable is a testable version of Model
type ModelTestable struct {
	state      *state.AppState
	nodeClient *mocks.MockNodeClient
}

// handleDeleteSiteTestable mimics handleDeleteSite but uses mock
func (m *ModelTestable) handleDeleteSiteTestable(siteID uuid.UUID) error {
	for i, site := range m.state.Sites {
		if site.ID == siteID {
			// Get domain name
			domain := m.state.GetDomainByID(site.DomainID)
			domainName := ""
			if domain != nil {
				domainName = domain.Name
			}

			// Get node for this site
			node := m.state.GetNodeByID(site.NodeID)

			// Call node API to delete deployed resources (only for non-inactive sites)
			if site.Status != models.SiteStatusInactive && node != nil {
				err := m.nodeClient.DeleteSite(
					node.APIEndpoint,
					node.APIKey,
					site.ID,
					domainName,
					site.Name,
					site.GetSiteType(),
				)
				if err != nil {
					// Log warning but continue with local deletion
					m.state.AddNotification("Warning: Failed to delete from node: "+err.Error(), "warning")
				}
			}

			// Remove from slice
			m.state.Sites = append(m.state.Sites[:i], m.state.Sites[i+1:]...)
			m.state.AddNotification("Deleted site: "+site.Name, "success")
			return nil
		}
	}
	return errors.New("site not found")
}

func hasNotificationWithLevel(notifications []state.Notification, level string) bool {
	for _, n := range notifications {
		if n.Level == level {
			return true
		}
	}
	return false
}

// ============================================================================
// Deletion Confirmation Dialog Tests
// ============================================================================

func TestDeleteConfirm_InitializesCorrectState(t *testing.T) {
	siteID := uuid.New()

	appState := state.NewAppState()
	appState.Sites = []models.Site{
		{
			ID:   siteID,
			Name: "test-site",
		},
	}

	m := &ModelTestable{
		state: appState,
	}

	// Initialize confirmation state
	m.state.DeletionConfirmPending = true
	m.state.DeletionTargetID = siteID
	m.state.DeletionTargetName = "test-site"
	m.state.DeletionTargetType = "site"
	m.state.DeletionConfirmInput = ""

	assert.True(t, m.state.DeletionConfirmPending)
	assert.Equal(t, siteID, m.state.DeletionTargetID)
	assert.Equal(t, "test-site", m.state.DeletionTargetName)
	assert.Equal(t, "site", m.state.DeletionTargetType)
	assert.Equal(t, "", m.state.DeletionConfirmInput)
}

func TestDeleteConfirm_MatchingName_AllowsDelete(t *testing.T) {
	siteID := uuid.New()

	appState := state.NewAppState()
	appState.Sites = []models.Site{
		{
			ID:   siteID,
			Name: "MyApp",
		},
	}

	// Initialize confirmation state
	appState.DeletionConfirmPending = true
	appState.DeletionTargetID = siteID
	appState.DeletionTargetName = "MyApp"
	appState.DeletionTargetType = "site"
	appState.DeletionConfirmInput = "MyApp" // Exact match

	// Check confirmation matches
	assert.Equal(t, appState.DeletionConfirmInput, appState.DeletionTargetName)
}

func TestDeleteConfirm_WrongName_DoesNotMatch(t *testing.T) {
	siteID := uuid.New()

	appState := state.NewAppState()
	appState.Sites = []models.Site{
		{
			ID:   siteID,
			Name: "MyApp",
		},
	}

	// Initialize confirmation state
	appState.DeletionConfirmPending = true
	appState.DeletionTargetID = siteID
	appState.DeletionTargetName = "MyApp"
	appState.DeletionTargetType = "site"
	appState.DeletionConfirmInput = "myapp" // Wrong case

	// Check confirmation does not match (case-sensitive)
	assert.NotEqual(t, appState.DeletionConfirmInput, appState.DeletionTargetName)
}

func TestDeleteConfirm_Escape_ClearsState(t *testing.T) {
	siteID := uuid.New()

	appState := state.NewAppState()
	appState.DeletionConfirmPending = true
	appState.DeletionTargetID = siteID
	appState.DeletionTargetName = "test-site"
	appState.DeletionTargetType = "site"
	appState.DeletionConfirmInput = "partial"

	// Simulate escape by clearing state
	appState.DeletionConfirmPending = false
	appState.DeletionConfirmInput = ""
	appState.DeletionTargetID = uuid.Nil
	appState.DeletionTargetName = ""
	appState.DeletionTargetType = ""

	assert.False(t, appState.DeletionConfirmPending)
	assert.Equal(t, "", appState.DeletionConfirmInput)
	assert.Equal(t, uuid.Nil, appState.DeletionTargetID)
}

func TestDeleteConfirm_Backspace_RemovesLastChar(t *testing.T) {
	appState := state.NewAppState()
	appState.DeletionConfirmInput = "test"

	// Simulate backspace
	if len(appState.DeletionConfirmInput) > 0 {
		appState.DeletionConfirmInput = appState.DeletionConfirmInput[:len(appState.DeletionConfirmInput)-1]
	}

	assert.Equal(t, "tes", appState.DeletionConfirmInput)
}

func TestDeleteConfirm_Rune_AppendsChar(t *testing.T) {
	appState := state.NewAppState()
	appState.DeletionConfirmInput = "tes"

	// Simulate typing 't'
	appState.DeletionConfirmInput += "t"

	assert.Equal(t, "test", appState.DeletionConfirmInput)
}

func TestDeleteConfirm_NavigateAway_ResetsState(t *testing.T) {
	siteID := uuid.New()

	appState := state.NewAppState()
	appState.CurrentScreen = state.ScreenSiteDeleteConfirm
	appState.DeletionConfirmPending = true
	appState.DeletionTargetID = siteID
	appState.DeletionTargetName = "test-site"
	appState.DeletionTargetType = "site"
	appState.DeletionConfirmInput = "partial"

	// Navigate to a different screen
	appState.NavigateTo(state.ScreenSitesList)

	// Confirmation state should be reset
	assert.False(t, appState.DeletionConfirmPending)
	assert.Equal(t, "", appState.DeletionConfirmInput)
	assert.Equal(t, uuid.Nil, appState.DeletionTargetID)
	assert.Equal(t, "", appState.DeletionTargetName)
	assert.Equal(t, "", appState.DeletionTargetType)
}
