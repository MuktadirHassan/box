package box

import (
	"fmt"
	"regexp"
)

var templatePattern = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)

func ValidateTemplate(name string) error {
	if name == "" {
		return nil
	}
	if !templatePattern.MatchString(name) {
		return fmt.Errorf("invalid template %q: use lowercase letters, numbers, and single hyphens or periods only", name)
	}
	return nil
}
