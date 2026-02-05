package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSite_ResolveDockerCredentials(t *testing.T) {
	credID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name             string
		site             Site
		credentialLookup func(uuid.UUID) (string, string)
		wantUsername     string
		wantToken        string
	}{
		{
			name: "site with DockerCredentialID - returns credential from lookup",
			site: Site{
				DockerCredentialID: &credID,
				DockerUsername:     "legacy_user", // Should be ignored
				DockerToken:        "legacy_token",
			},
			credentialLookup: func(id uuid.UUID) (string, string) {
				if id == credID {
					return "cred_user", "cred_token"
				}
				return "", ""
			},
			wantUsername: "cred_user",
			wantToken:    "cred_token",
		},
		{
			name: "site with legacy DockerUsername/Token - returns those",
			site: Site{
				DockerCredentialID: nil,
				DockerUsername:     "legacy_user",
				DockerToken:        "legacy_token",
			},
			credentialLookup: func(id uuid.UUID) (string, string) {
				return "should_not_be_called", "should_not_be_called"
			},
			wantUsername: "legacy_user",
			wantToken:    "legacy_token",
		},
		{
			name: "site with nil DockerCredentialID and empty legacy - returns empty",
			site: Site{
				DockerCredentialID: nil,
				DockerUsername:     "",
				DockerToken:        "",
			},
			credentialLookup: func(id uuid.UUID) (string, string) {
				return "", ""
			},
			wantUsername: "",
			wantToken:    "",
		},
		{
			name: "site with invalid DockerCredentialID - returns empty gracefully",
			site: Site{
				DockerCredentialID: func() *uuid.UUID {
					id := uuid.MustParse("99999999-9999-9999-9999-999999999999")
					return &id
				}(),
				DockerUsername: "",
				DockerToken:    "",
			},
			credentialLookup: func(id uuid.UUID) (string, string) {
				// Credential not found - return empty
				return "", ""
			},
			wantUsername: "",
			wantToken:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, token := tt.site.ResolveDockerCredentials(tt.credentialLookup)
			assert.Equal(t, tt.wantUsername, username)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

func TestParsePortMapping(t *testing.T) {
	tests := []struct {
		name          string
		portStr       string
		wantContainer int
		wantHost      int
		wantErr       bool
	}{
		{
			name:          "single port",
			portStr:       "3000",
			wantContainer: 3000,
			wantHost:      3000,
			wantErr:       false,
		},
		{
			name:          "container:host format",
			portStr:       "3000:3001",
			wantContainer: 3000,
			wantHost:      3001,
			wantErr:       false,
		},
		{
			name:          "with whitespace",
			portStr:       " 8080 ",
			wantContainer: 8080,
			wantHost:      8080,
			wantErr:       false,
		},
		{
			name:    "empty string",
			portStr: "",
			wantErr: true,
		},
		{
			name:    "invalid port",
			portStr: "abc",
			wantErr: true,
		},
		{
			name:    "port out of range high",
			portStr: "99999",
			wantErr: true,
		},
		{
			name:    "port out of range low",
			portStr: "0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, host, err := ParsePortMapping(tt.portStr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantContainer, container)
			assert.Equal(t, tt.wantHost, host)
		})
	}
}

func TestFormatPortMapping(t *testing.T) {
	tests := []struct {
		name          string
		containerPort int
		hostPort      int
		want          string
	}{
		{
			name:          "same port",
			containerPort: 3000,
			hostPort:      3000,
			want:          "3000",
		},
		{
			name:          "different ports",
			containerPort: 3000,
			hostPort:      3001,
			want:          "3000:3001",
		},
		{
			name:          "host port zero",
			containerPort: 8080,
			hostPort:      0,
			want:          "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPortMapping(tt.containerPort, tt.hostPort)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDomainMapping_GetEffectiveHostPort(t *testing.T) {
	tests := []struct {
		name     string
		mapping  DomainMapping
		wantPort int
	}{
		{
			name: "host port set",
			mapping: DomainMapping{
				Port:     3000,
				HostPort: 8080,
			},
			wantPort: 8080,
		},
		{
			name: "host port zero - falls back to Port",
			mapping: DomainMapping{
				Port:     3000,
				HostPort: 0,
			},
			wantPort: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mapping.GetEffectiveHostPort()
			assert.Equal(t, tt.wantPort, result)
		})
	}
}
