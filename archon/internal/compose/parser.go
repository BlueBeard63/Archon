package compose

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComposeFile represents a minimal docker-compose.yml structure
type ComposeFile struct {
	Services map[string]Service `yaml:"services"`
}

// Service represents a service in a compose file
type Service struct {
	Ports []interface{} `yaml:"ports"` // Can be string or map (short/long form)
}

// DetectedPort represents a parsed port from compose
type DetectedPort struct {
	ServiceName   string
	ContainerPort int
	HostPort      int    // 0 if not specified
	Protocol      string // "tcp" or "udp", defaults to "tcp"
}

// ParsePorts extracts exposed ports from compose YAML content
func ParsePorts(content string) ([]DetectedPort, error) {
	var compose ComposeFile
	if err := yaml.Unmarshal([]byte(content), &compose); err != nil {
		return nil, fmt.Errorf("failed to parse compose YAML: %w", err)
	}

	var ports []DetectedPort

	// Iterate services in map order (Go maps are unordered, but we take what we get)
	for serviceName, service := range compose.Services {
		for _, portDef := range service.Ports {
			detected, err := parsePortDefinition(serviceName, portDef)
			if err != nil {
				// Log warning but continue parsing other ports
				continue
			}
			if detected != nil {
				ports = append(ports, *detected)
			}
		}
	}

	return ports, nil
}

// parsePortDefinition handles both short form (string) and long form (map) port definitions
func parsePortDefinition(serviceName string, portDef interface{}) (*DetectedPort, error) {
	switch v := portDef.(type) {
	case string:
		return parseShortForm(serviceName, v)
	case int:
		// Just a port number
		return &DetectedPort{
			ServiceName:   serviceName,
			ContainerPort: v,
			Protocol:      "tcp",
		}, nil
	case map[string]interface{}:
		return parseLongForm(serviceName, v)
	default:
		return nil, fmt.Errorf("unknown port definition type: %T", portDef)
	}
}

// parseShortForm parses short form port definitions like:
// - "3000"
// - "8000:8000"
// - "127.0.0.1:8001:8001"
// - "6060:6060/udp"
// - "9090-9091:8080-8081"
func parseShortForm(serviceName, portStr string) (*DetectedPort, error) {
	detected := &DetectedPort{
		ServiceName: serviceName,
		Protocol:    "tcp",
	}

	// Handle protocol suffix
	portStr = strings.TrimSpace(portStr)
	if strings.HasSuffix(portStr, "/udp") {
		detected.Protocol = "udp"
		portStr = strings.TrimSuffix(portStr, "/udp")
	} else if strings.HasSuffix(portStr, "/tcp") {
		portStr = strings.TrimSuffix(portStr, "/tcp")
	}

	// Split by colon to determine format
	parts := strings.Split(portStr, ":")

	var containerPortStr string
	var hostPortStr string

	switch len(parts) {
	case 1:
		// Just container port: "3000"
		containerPortStr = parts[0]
	case 2:
		// HOST:CONTAINER: "8000:8000"
		hostPortStr = parts[0]
		containerPortStr = parts[1]
	case 3:
		// IP:HOST:CONTAINER: "127.0.0.1:8001:8001"
		// parts[0] is IP, parts[1] is host port, parts[2] is container port
		hostPortStr = parts[1]
		containerPortStr = parts[2]
	default:
		return nil, fmt.Errorf("invalid port format: %s", portStr)
	}

	// Parse container port (handle ranges like "8080-8081")
	containerPort, err := parsePortNumber(containerPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid container port: %w", err)
	}
	detected.ContainerPort = containerPort

	// Parse host port if present
	if hostPortStr != "" {
		hostPort, err := parsePortNumber(hostPortStr)
		if err != nil {
			// Host port parsing failed, but we have container port
			detected.HostPort = 0
		} else {
			detected.HostPort = hostPort
		}
	}

	return detected, nil
}

// parsePortNumber extracts a port number from a string, handling ranges (takes first port)
func parsePortNumber(s string) (int, error) {
	s = strings.TrimSpace(s)

	// Handle port ranges like "8000-8005" - take first port
	if idx := strings.Index(s, "-"); idx > 0 {
		s = s[:idx]
	}

	port, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not a valid port number: %s", s)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port out of range: %d", port)
	}

	return port, nil
}

// parseLongForm parses long form port definitions like:
// - target: 80
//   published: "8080"
//   host_ip: 127.0.0.1
//   protocol: tcp
func parseLongForm(serviceName string, portMap map[string]interface{}) (*DetectedPort, error) {
	detected := &DetectedPort{
		ServiceName: serviceName,
		Protocol:    "tcp",
	}

	// Extract target (container port) - required
	if target, ok := portMap["target"]; ok {
		switch v := target.(type) {
		case int:
			detected.ContainerPort = v
		case string:
			port, err := parsePortNumber(v)
			if err != nil {
				return nil, err
			}
			detected.ContainerPort = port
		default:
			return nil, fmt.Errorf("invalid target port type: %T", target)
		}
	} else {
		return nil, fmt.Errorf("long form port missing 'target' field")
	}

	// Extract published (host port) - optional
	if published, ok := portMap["published"]; ok {
		switch v := published.(type) {
		case int:
			detected.HostPort = v
		case string:
			port, err := parsePortNumber(v)
			if err == nil {
				detected.HostPort = port
			}
		}
	}

	// Extract protocol - optional
	if protocol, ok := portMap["protocol"]; ok {
		if p, ok := protocol.(string); ok {
			detected.Protocol = p
		}
	}

	return detected, nil
}

// GetFirstPort returns the first detected port, or 0 if none found
func GetFirstPort(ports []DetectedPort) int {
	if len(ports) == 0 {
		return 0
	}
	return ports[0].ContainerPort
}
