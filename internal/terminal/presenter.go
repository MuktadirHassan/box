package terminal

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/ui"
)

type Presenter struct{}

func NewPresenter() ui.Presenter { return Presenter{} }

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

func (Presenter) ConfigureInitial(definition box.Definition) (box.Definition, error) {
	if !interactiveTerminal() {
		return definition, nil
	}
	configuration := &definition.Configuration
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Base image").Description("The image used to create the development environment.").Value(&configuration.Image).Validate(nonEmpty("base image")),
		huh.NewInput().Title("Linux user").Description("The user account created inside the box.").Value(&configuration.User).Validate(box.ValidateUser),
		huh.NewSelect[string]().Title("Network policy").Options(huh.NewOption("Outbound network access", "outbound"), huh.NewOption("No network access", "none")).Value(&configuration.Network),
		huh.NewSelect[string]().Title("Environment template").Description("Optional tools installed when the box is created; requires Ubuntu, Debian, or Arch Linux.").Options(huh.NewOption("No template", ""), huh.NewOption("Terminal tools (fish, tmux, neovim, and more)", box.TerminalToolsTemplate)).Value(&configuration.Template),
		huh.NewConfirm().Title("Persist the home directory?").Value(&configuration.Home.Enabled),
		huh.NewConfirm().Title("Persist development caches?").Value(&configuration.Caches.Enabled),
		huh.NewConfirm().Title("Enable clipboard integration?").Value(&configuration.Integrations.Clipboard),
		huh.NewConfirm().Title("Enable SSH agent forwarding?").Value(&configuration.Integrations.SSHAgent),
	))
	if err := form.Run(); err != nil {
		return box.Definition{}, fmt.Errorf("configure box: %w", err)
	}
	return definition, nil
}

func (Presenter) ConfirmSetup() error {
	if !interactiveTerminal() {
		return fmt.Errorf("interactive confirmation is unavailable")
	}
	var confirmed bool
	if err := huh.NewConfirm().Title(render(titleStyle, "Apply this configuration?")).Affirmative("Apply").Negative("Cancel").Value(&confirmed).Run(); err != nil {
		return fmt.Errorf("confirm setup: %w", err)
	}
	if !confirmed {
		return fmt.Errorf("setup cancelled")
	}
	return nil
}

func (Presenter) ShowDefinition(writer io.Writer, definition box.Definition) error {
	if _, err := fmt.Fprintln(writer, render(titleStyle, "Box "+definition.Name)); err != nil {
		return err
	}
	fields := [][2]string{{"State", string(definition.State)}, {"Backend", string(definition.Backend)}, {"Version", strconv.Itoa(definition.Version)}}
	if definition.State == box.ReadyState {
		configuration := definition.Configuration
		fields = append(fields, [2]string{"Image", configuration.Image}, [2]string{"User", configuration.User}, [2]string{"Network", configuration.Network}, [2]string{"Template", configuration.Template}, [2]string{"Persistent home", strconv.FormatBool(configuration.Home.Enabled)}, [2]string{"Persistent caches", strconv.FormatBool(configuration.Caches.Enabled)}, [2]string{"Clipboard", strconv.FormatBool(configuration.Integrations.Clipboard)}, [2]string{"SSH agent", strconv.FormatBool(configuration.Integrations.SSHAgent)})
		for _, mount := range configuration.Mounts {
			fields = append(fields, [2]string{"Mount", mount.Source + ":" + mount.Destination})
		}
	}
	for _, field := range fields {
		if err := writeInspectField(writer, field[0], field[1]); err != nil {
			return err
		}
	}
	return nil
}

func (Presenter) ShowRuntime(writer io.Writer, state box.RuntimeState, id string) error {
	if err := writeInspectField(writer, "Runtime", string(state)); err != nil {
		return err
	}
	if id == "" {
		return nil
	}
	return writeInspectField(writer, "Runtime ID", id)
}

func writeInspectField(writer io.Writer, label, value string) error {
	_, err := fmt.Fprintf(writer, "  %s  %s\n", render(labelStyle, pad(label, len("Persistent caches"))), value)
	return err
}

func (Presenter) ShowList(writer io.Writer, definitions []box.Definition) error {
	if err := writeTitle(writer, "Boxes"); err != nil {
		return err
	}
	if len(definitions) == 0 {
		_, err := fmt.Fprintln(writer, "  No boxes yet. Create one with: box create <name>")
		return err
	}

	type row struct{ name, state, image string }
	rows := make([]row, 0, len(definitions))
	nameWidth, stateWidth := len("NAME"), len("STATE")
	for _, definition := range definitions {
		image := definition.Configuration.Image
		if image == "" {
			image = "-"
		}
		item := row{definition.Name, string(definition.State), image}
		rows = append(rows, item)
		nameWidth = max(nameWidth, len(item.name))
		stateWidth = max(stateWidth, len(item.state))
	}
	if _, err := fmt.Fprintf(writer, "  %s  %s  %s\n", render(labelStyle, pad("NAME", nameWidth)), render(labelStyle, pad("STATE", stateWidth)), render(labelStyle, "IMAGE")); err != nil {
		return err
	}
	for _, item := range rows {
		if _, err := fmt.Fprintf(writer, "  %s  %s  %s\n", pad(item.name, nameWidth), pad(item.state, stateWidth), item.image); err != nil {
			return err
		}
	}
	return nil
}

func pad(value string, width int) string {
	return fmt.Sprintf("%-*s", width, value)
}

func (Presenter) ShowWarning(writer io.Writer, message string) error {
	_, err := fmt.Fprintln(writer, render(warningStyle, message))
	return err
}

func (Presenter) ShowSuccess(writer io.Writer, message string) error {
	_, err := fmt.Fprintln(writer, render(lipgloss.NewStyle().Foreground(lipgloss.Color("10")), message))
	return err
}

func writeTitle(writer io.Writer, value string) error {
	_, err := fmt.Fprintln(writer, render(titleStyle, value))
	return err
}
func nonEmpty(name string) func(string) error {
	return func(value string) error {
		if value == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}
		return nil
	}
}
func interactiveTerminal() bool { return terminal(os.Stdin) && terminal(os.Stdout) }
func terminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
func render(style lipgloss.Style, value string) string {
	if !interactiveTerminal() {
		return value
	}
	return style.Render(value)
}
