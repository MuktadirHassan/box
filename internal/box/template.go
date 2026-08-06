package box

import "fmt"

func ValidateTemplate(name string) error {
	if name == "" {
		return nil
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid template %q: use lowercase letters, numbers, and single hyphens only", name)
	}
	return nil
}
