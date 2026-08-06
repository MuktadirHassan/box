package box

import "fmt"

const TerminalToolsTemplate = "terminal-tools"

var templatePackages = map[string][]string{
	TerminalToolsTemplate: {"fish", "jq", "neovim", "tmux", "ripgrep", "starship", "wl-clipboard"},
}

func ValidateTemplate(name string) error {
	if name == "" {
		return nil
	}
	if _, ok := templatePackages[name]; !ok {
		return fmt.Errorf("template %q is not supported; use %s", name, TerminalToolsTemplate)
	}
	return nil
}

func TemplatePackages(name string) []string {
	packages := templatePackages[name]
	return append([]string(nil), packages...)
}
