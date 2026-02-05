package config

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerCredential_Add(t *testing.T) {
	tests := []struct {
		name       string
		credential DockerCredential
		wantErr    bool
	}{
		{
			name: "valid credential",
			credential: DockerCredential{
				Name:     "GitHub Container Registry",
				Registry: "ghcr.io",
				Username: "myuser",
				Token:    "ghp_xxxxxxxxxxxx",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			credential: DockerCredential{
				Name:     "",
				Registry: "ghcr.io",
				Username: "myuser",
				Token:    "ghp_xxxxxxxxxxxx",
			},
			wantErr: true,
		},
		{
			name: "empty registry",
			credential: DockerCredential{
				Name:     "Test",
				Registry: "",
				Username: "myuser",
				Token:    "ghp_xxxxxxxxxxxx",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &Settings{
				DockerCredentials: []DockerCredential{},
			}

			err := settings.AddDockerCredential(tt.credential)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, settings.DockerCredentials, 1)
			assert.NotEqual(t, uuid.Nil, settings.DockerCredentials[0].ID)
			assert.Equal(t, tt.credential.Name, settings.DockerCredentials[0].Name)
			assert.Equal(t, tt.credential.Registry, settings.DockerCredentials[0].Registry)
		})
	}
}

func TestDockerCredential_GetByID(t *testing.T) {
	settings := &Settings{
		DockerCredentials: []DockerCredential{
			{
				ID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Name:     "GHCR",
				Registry: "ghcr.io",
				Username: "user1",
				Token:    "token1",
			},
			{
				ID:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Name:     "DockerHub",
				Registry: "docker.io",
				Username: "user2",
				Token:    "token2",
			},
		},
	}

	tests := []struct {
		name     string
		id       uuid.UUID
		wantName string
		wantNil  bool
	}{
		{
			name:     "existing credential",
			id:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			wantName: "GHCR",
			wantNil:  false,
		},
		{
			name:     "second credential",
			id:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			wantName: "DockerHub",
			wantNil:  false,
		},
		{
			name:    "non-existent credential",
			id:      uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			wantNil: true,
		},
		{
			name:    "nil UUID",
			id:      uuid.Nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := settings.GetDockerCredentialByID(tt.id)
			if tt.wantNil {
				assert.Nil(t, cred)
			} else {
				require.NotNil(t, cred)
				assert.Equal(t, tt.wantName, cred.Name)
			}
		})
	}
}

func TestDockerCredential_Update(t *testing.T) {
	credID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	settings := &Settings{
		DockerCredentials: []DockerCredential{
			{
				ID:       credID,
				Name:     "Old Name",
				Registry: "old.registry.io",
				Username: "olduser",
				Token:    "oldtoken",
			},
		},
	}

	tests := []struct {
		name       string
		id         uuid.UUID
		credential DockerCredential
		wantErr    bool
	}{
		{
			name: "update existing credential",
			id:   credID,
			credential: DockerCredential{
				Name:     "New Name",
				Registry: "new.registry.io",
				Username: "newuser",
				Token:    "newtoken",
			},
			wantErr: false,
		},
		{
			name: "update non-existent credential",
			id:   uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			credential: DockerCredential{
				Name:     "Test",
				Registry: "test.io",
				Username: "test",
				Token:    "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := settings.UpdateDockerCredential(tt.id, tt.credential)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			cred := settings.GetDockerCredentialByID(tt.id)
			require.NotNil(t, cred)
			assert.Equal(t, tt.credential.Name, cred.Name)
			assert.Equal(t, tt.credential.Registry, cred.Registry)
			assert.Equal(t, tt.credential.Username, cred.Username)
			assert.Equal(t, tt.credential.Token, cred.Token)
			// ID should remain unchanged
			assert.Equal(t, tt.id, cred.ID)
		})
	}
}

func TestDockerCredential_Delete(t *testing.T) {
	tests := []struct {
		name        string
		initialCred []DockerCredential
		deleteID    uuid.UUID
		wantErr     bool
		wantCount   int
	}{
		{
			name: "delete existing credential",
			initialCred: []DockerCredential{
				{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Cred1"},
				{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "Cred2"},
			},
			deleteID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "delete non-existent credential",
			initialCred: []DockerCredential{
				{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Cred1"},
			},
			deleteID:  uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			wantErr:   true,
			wantCount: 1,
		},
		{
			name: "delete last credential",
			initialCred: []DockerCredential{
				{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Cred1"},
			},
			deleteID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &Settings{
				DockerCredentials: make([]DockerCredential, len(tt.initialCred)),
			}
			copy(settings.DockerCredentials, tt.initialCred)

			err := settings.DeleteDockerCredential(tt.deleteID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, settings.DockerCredentials, tt.wantCount)
		})
	}
}

func TestDockerCredential_ListAll(t *testing.T) {
	tests := []struct {
		name        string
		credentials []DockerCredential
		wantCount   int
	}{
		{
			name:        "empty list",
			credentials: []DockerCredential{},
			wantCount:   0,
		},
		{
			name:        "nil list",
			credentials: nil,
			wantCount:   0,
		},
		{
			name: "multiple credentials",
			credentials: []DockerCredential{
				{ID: uuid.New(), Name: "Cred1"},
				{ID: uuid.New(), Name: "Cred2"},
				{ID: uuid.New(), Name: "Cred3"},
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &Settings{
				DockerCredentials: tt.credentials,
			}
			list := settings.ListDockerCredentials()
			assert.Len(t, list, tt.wantCount)
		})
	}
}
