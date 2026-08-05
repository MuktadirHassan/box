package box

import "testing"

func TestValidateBackend(t *testing.T) {
	for _, backend := range []Backend{PodmanBackend, LimaBackend} {
		if err := ValidateBackend(backend); err != nil {
			t.Errorf("ValidateBackend(%q) error = %v", backend, err)
		}
	}

	if err := ValidateBackend("future-runtime"); err != nil {
		t.Errorf("ValidateBackend(future-runtime) error = %v", err)
	}
	if err := ValidateBackend("invalid backend"); err == nil {
		t.Error("ValidateBackend(invalid backend) error = nil")
	}
}
