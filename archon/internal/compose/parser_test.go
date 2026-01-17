package compose

import (
	"testing"
)

func TestParsePorts_ShortForm(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		wantPorts     int
		wantContainer int
		wantHost      int
		wantProtocol  string
	}{
		{
			name: "container port only",
			yaml: `
services:
  web:
    ports:
      - "3000"
`,
			wantPorts:     1,
			wantContainer: 3000,
			wantHost:      0,
			wantProtocol:  "tcp",
		},
		{
			name: "host:container port",
			yaml: `
services:
  web:
    ports:
      - "8080:80"
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      8080,
			wantProtocol:  "tcp",
		},
		{
			name: "ip:host:container port",
			yaml: `
services:
  web:
    ports:
      - "127.0.0.1:3000:80"
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      3000,
			wantProtocol:  "tcp",
		},
		{
			name: "udp protocol suffix",
			yaml: `
services:
  dns:
    ports:
      - "6060:6060/udp"
`,
			wantPorts:     1,
			wantContainer: 6060,
			wantHost:      6060,
			wantProtocol:  "udp",
		},
		{
			name: "tcp protocol suffix",
			yaml: `
services:
  web:
    ports:
      - "8080:80/tcp"
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      8080,
			wantProtocol:  "tcp",
		},
		{
			name: "port range extracts first port",
			yaml: `
services:
  web:
    ports:
      - "8000-8005:8000-8005"
`,
			wantPorts:     1,
			wantContainer: 8000,
			wantHost:      8000,
			wantProtocol:  "tcp",
		},
		{
			name: "integer port",
			yaml: `
services:
  web:
    ports:
      - 3000
`,
			wantPorts:     1,
			wantContainer: 3000,
			wantHost:      0,
			wantProtocol:  "tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := ParsePorts(tt.yaml)
			if err != nil {
				t.Fatalf("ParsePorts() error = %v", err)
			}
			if len(ports) != tt.wantPorts {
				t.Errorf("ParsePorts() got %d ports, want %d", len(ports), tt.wantPorts)
			}
			if len(ports) > 0 {
				if ports[0].ContainerPort != tt.wantContainer {
					t.Errorf("ContainerPort = %d, want %d", ports[0].ContainerPort, tt.wantContainer)
				}
				if ports[0].HostPort != tt.wantHost {
					t.Errorf("HostPort = %d, want %d", ports[0].HostPort, tt.wantHost)
				}
				if ports[0].Protocol != tt.wantProtocol {
					t.Errorf("Protocol = %s, want %s", ports[0].Protocol, tt.wantProtocol)
				}
			}
		})
	}
}

func TestParsePorts_LongForm(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		wantPorts     int
		wantContainer int
		wantHost      int
		wantProtocol  string
	}{
		{
			name: "target only",
			yaml: `
services:
  web:
    ports:
      - target: 80
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      0,
			wantProtocol:  "tcp",
		},
		{
			name: "target and published",
			yaml: `
services:
  web:
    ports:
      - target: 80
        published: 8080
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      8080,
			wantProtocol:  "tcp",
		},
		{
			name: "target with string published",
			yaml: `
services:
  web:
    ports:
      - target: 80
        published: "8080"
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      8080,
			wantProtocol:  "tcp",
		},
		{
			name: "with protocol",
			yaml: `
services:
  dns:
    ports:
      - target: 53
        published: 53
        protocol: udp
`,
			wantPorts:     1,
			wantContainer: 53,
			wantHost:      53,
			wantProtocol:  "udp",
		},
		{
			name: "full long form",
			yaml: `
services:
  web:
    ports:
      - target: 80
        published: 8080
        host_ip: 127.0.0.1
        protocol: tcp
`,
			wantPorts:     1,
			wantContainer: 80,
			wantHost:      8080,
			wantProtocol:  "tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := ParsePorts(tt.yaml)
			if err != nil {
				t.Fatalf("ParsePorts() error = %v", err)
			}
			if len(ports) != tt.wantPorts {
				t.Errorf("ParsePorts() got %d ports, want %d", len(ports), tt.wantPorts)
			}
			if len(ports) > 0 {
				if ports[0].ContainerPort != tt.wantContainer {
					t.Errorf("ContainerPort = %d, want %d", ports[0].ContainerPort, tt.wantContainer)
				}
				if ports[0].HostPort != tt.wantHost {
					t.Errorf("HostPort = %d, want %d", ports[0].HostPort, tt.wantHost)
				}
				if ports[0].Protocol != tt.wantProtocol {
					t.Errorf("Protocol = %s, want %s", ports[0].Protocol, tt.wantProtocol)
				}
			}
		})
	}
}

func TestParsePorts_MultipleServices(t *testing.T) {
	yaml := `
services:
  web:
    ports:
      - "8080:80"
  api:
    ports:
      - "3000:3000"
  db:
    ports:
      - "5432"
`
	ports, err := ParsePorts(yaml)
	if err != nil {
		t.Fatalf("ParsePorts() error = %v", err)
	}
	if len(ports) != 3 {
		t.Errorf("ParsePorts() got %d ports, want 3", len(ports))
	}

	// Check that we got ports from different services
	services := make(map[string]bool)
	for _, p := range ports {
		services[p.ServiceName] = true
	}
	if len(services) != 3 {
		t.Errorf("ParsePorts() got %d services, want 3", len(services))
	}
}

func TestParsePorts_MultiplePorts(t *testing.T) {
	yaml := `
services:
  web:
    ports:
      - "80"
      - "443"
      - "8080:8080"
`
	ports, err := ParsePorts(yaml)
	if err != nil {
		t.Fatalf("ParsePorts() error = %v", err)
	}
	if len(ports) != 3 {
		t.Errorf("ParsePorts() got %d ports, want 3", len(ports))
	}
}

func TestParsePorts_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantPorts int
		wantErr   bool
	}{
		{
			name:      "empty ports",
			yaml:      `services: { web: { ports: [] } }`,
			wantPorts: 0,
			wantErr:   false,
		},
		{
			name:      "no ports key",
			yaml:      `services: { web: { image: nginx } }`,
			wantPorts: 0,
			wantErr:   false,
		},
		{
			name:      "empty services",
			yaml:      `services: {}`,
			wantPorts: 0,
			wantErr:   false,
		},
		{
			name:      "invalid yaml",
			yaml:      `this is not valid yaml: [`,
			wantPorts: 0,
			wantErr:   true,
		},
		{
			name: "service without ports",
			yaml: `
services:
  web:
    image: nginx
  api:
    ports:
      - "3000"
`,
			wantPorts: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := ParsePorts(tt.yaml)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(ports) != tt.wantPorts {
				t.Errorf("ParsePorts() got %d ports, want %d", len(ports), tt.wantPorts)
			}
		})
	}
}

func TestParsePorts_LongFormMissingTarget(t *testing.T) {
	yaml := `
services:
  web:
    ports:
      - published: 8080
`
	// Should not error, but should skip the invalid port
	ports, err := ParsePorts(yaml)
	if err != nil {
		t.Fatalf("ParsePorts() should not error for missing target, got: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("ParsePorts() should return 0 ports when target is missing, got %d", len(ports))
	}
}

func TestGetFirstPort(t *testing.T) {
	tests := []struct {
		name  string
		ports []DetectedPort
		want  int
	}{
		{
			name:  "empty slice",
			ports: []DetectedPort{},
			want:  0,
		},
		{
			name:  "nil slice",
			ports: nil,
			want:  0,
		},
		{
			name: "single port",
			ports: []DetectedPort{
				{ContainerPort: 8080},
			},
			want: 8080,
		},
		{
			name: "multiple ports returns first",
			ports: []DetectedPort{
				{ContainerPort: 80},
				{ContainerPort: 443},
				{ContainerPort: 8080},
			},
			want: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFirstPort(tt.ports)
			if got != tt.want {
				t.Errorf("GetFirstPort() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParsePortNumber(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"80", 80, false},
		{"8080", 8080, false},
		{"  3000  ", 3000, false},
		{"8000-8005", 8000, false},
		{"65535", 65535, false},
		{"1", 1, false},
		{"0", 0, true},
		{"65536", 0, true},
		{"-1", 0, true},
		{"not-a-number", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parsePortNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePortNumber(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parsePortNumber(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePorts_RealWorldCompose(t *testing.T) {
	// Test with a realistic compose file
	yaml := `
version: "3.8"
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - app

  app:
    build: .
    ports:
      - "3000"
    environment:
      - NODE_ENV=production

  redis:
    image: redis:alpine
    ports:
      - target: 6379
        published: 6379

  postgres:
    image: postgres:14
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
`
	ports, err := ParsePorts(yaml)
	if err != nil {
		t.Fatalf("ParsePorts() error = %v", err)
	}

	if len(ports) != 5 {
		t.Errorf("ParsePorts() got %d ports, want 5", len(ports))
	}

	// Verify specific ports exist (order is non-deterministic due to map iteration)
	containerPorts := make(map[int]bool)
	for _, p := range ports {
		containerPorts[p.ContainerPort] = true
	}

	expectedPorts := []int{80, 443, 3000, 6379, 5432}
	for _, expected := range expectedPorts {
		if !containerPorts[expected] {
			t.Errorf("Expected port %d not found in parsed ports", expected)
		}
	}
}
