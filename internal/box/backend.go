package box

import (
	"fmt"
	"regexp"
)

type Backend string

const (
	PodmanBackend Backend = "podman"
	LimaBackend   Backend = "lima"
)

type RuntimeState string

const (
	RuntimeMissing RuntimeState = "missing"
	RuntimeCreated RuntimeState = "created"
	RuntimeRunning RuntimeState = "running"
	RuntimeStopped RuntimeState = "stopped"
)

type RuntimeMetadata struct {
	Backend Backend      `json:"backend"`
	ID      string       `json:"id"`
	State   RuntimeState `json:"state"`
}

type RuntimeStatus struct {
	State        RuntimeState `json:"state"`
	StorageBytes int64        `json:"storage_bytes"`
}

var backendPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateBackend(backend Backend) error {
	if !backendPattern.MatchString(string(backend)) {
		return fmt.Errorf("invalid backend %q: use lowercase letters, numbers, and single hyphens only", backend)
	}

	return nil
}
