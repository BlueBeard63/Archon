package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestStartHealthCheckTicker(t *testing.T) {
	tests := []struct {
		name         string
		intervalSecs int
		wantInterval time.Duration
	}{
		{
			name:         "custom interval",
			intervalSecs: 60,
			wantInterval: 60 * time.Second,
		},
		{
			name:         "default interval when zero",
			intervalSecs: 0,
			wantInterval: DefaultHealthCheckInterval,
		},
		{
			name:         "default interval when negative",
			intervalSecs: -1,
			wantInterval: DefaultHealthCheckInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := GetHealthCheckInterval(tt.intervalSecs)
			assert.Equal(t, tt.wantInterval, interval)
		})
	}
}

func TestTickMsg_IsTeaMsg(t *testing.T) {
	// Verify TickMsg implements tea.Msg interface
	var msg tea.Msg = TickMsg{}
	assert.NotNil(t, msg)
}

func TestHealthCheckResultMsg(t *testing.T) {
	// Verify NodeHealthCheckResultMsg can carry error information
	msg := NodeHealthCheckResultMsg{
		Error: nil,
	}
	assert.Nil(t, msg.Error)

	// With error
	msg = NodeHealthCheckResultMsg{
		Error: assert.AnError,
	}
	assert.NotNil(t, msg.Error)
}

func TestSiteStatusCheckResultMsg(t *testing.T) {
	// Verify SiteStatusCheckResultMsg structure
	msg := SiteStatusCheckResultMsg{
		Status: "running",
		Error:  nil,
	}
	assert.Equal(t, "running", msg.Status)
	assert.Nil(t, msg.Error)
}
