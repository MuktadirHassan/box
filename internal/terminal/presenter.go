package terminal

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/templates"
	"github.com/MuktadirHassan/box/internal/ui"
	assets "github.com/MuktadirHassan/box/templates"
)

type Presenter struct {
	catalog      templates.Catalog
	isTerminal   func(io.Writer) bool
	stepInterval time.Duration
}

func NewPresenter(catalog ...templates.Catalog) ui.Presenter {
	var c templates.Catalog
	if len(catalog) > 0 {
		c = catalog[0]
	}
	if c == nil {
		c = templates.NewEmbeddedCatalog(assets.FS())
	}
	return Presenter{catalog: c, isTerminal: terminalWriter, stepInterval: 80 * time.Millisecond}
}

type step struct {
	writer      io.Writer
	label       string
	interactive bool
	interval    time.Duration
	done        chan struct{}
	wait        sync.WaitGroup
	mutex       sync.Mutex
	frame       int
	writeErr    error
	finished    bool
}

var spinnerFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (p Presenter) StartStep(writer io.Writer, label string) (ui.Step, error) {
	interactive := p.isTerminal != nil && p.isTerminal(writer)
	status := &step{writer: writer, label: label, interactive: interactive}
	if !interactive {
		if _, err := fmt.Fprintf(writer, "%s...\n", label); err != nil {
			return nil, err
		}
		return status, nil
	}

	status.interval = p.stepInterval
	if status.interval <= 0 {
		status.interval = 80 * time.Millisecond
	}
	status.done = make(chan struct{})
	if err := status.writeFrame(); err != nil {
		return nil, err
	}
	status.wait.Add(1)
	go status.animate()
	return status, nil
}

func (s *step) animate() {
	defer s.wait.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mutex.Lock()
			if s.writeErr == nil {
				s.frame = (s.frame + 1) % len(spinnerFrames)
				s.writeErr = s.writeFrameLocked()
			}
			s.mutex.Unlock()
		case <-s.done:
			return
		}
	}
}

func (s *step) writeFrame() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.writeFrameLocked()
}

func (s *step) writeFrameLocked() error {
	_, err := fmt.Fprintf(s.writer, "\r\x1b[2K%s %s", spinnerFrames[s.frame], s.label)
	return err
}

func (s *step) Success() { s.finish("✓", false) }
func (s *step) Fail()    { s.finish("✗", true) }

func (s *step) finish(symbol string, failed bool) {
	s.mutex.Lock()
	if s.finished {
		s.mutex.Unlock()
		return
	}
	s.finished = true
	if s.interactive {
		close(s.done)
	}
	s.mutex.Unlock()

	if !s.interactive {
		if failed {
			_, _ = fmt.Fprintf(s.writer, "Failed: %s.\n", s.label)
		}
		return
	}

	s.wait.Wait()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.writeErr == nil {
		_, s.writeErr = fmt.Fprintf(s.writer, "\r\x1b[2K%s %s\n", symbol, s.label)
	}
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

func (p Presenter) ConfigureInitial(definition box.Definition) (box.Definition, error) {
	if !interactiveTerminal() {
		return definition, nil
	}
	configuration := &definition.Configuration
	templateOptions, err := environmentTemplateOptions(p.catalog)
	if err != nil {
		return box.Definition{}, err
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Base image").Description("The image used to create the development environment.").Value(&configuration.Image).Validate(nonEmpty("base image")),
		huh.NewInput().Title("Linux user").Description("The user account created inside the box.").Value(&configuration.User).Validate(box.ValidateUser),
		huh.NewSelect[string]().Title("Network policy").Options(huh.NewOption("Outbound network access", "outbound"), huh.NewOption("No network access", "none")).Value(&configuration.Network),
		huh.NewSelect[string]().Title("Environment template").Description("Optional tools installed when the box is created; currently supports Ubuntu images.").Options(templateOptions...).Value(&configuration.Template),
		huh.NewSelect[string]().Title("Interactive shell").Options(huh.NewOption("POSIX shell", "sh"), huh.NewOption("Bash", "bash"), huh.NewOption("Fish", "fish"), huh.NewOption("Zsh", "zsh")).Value(&configuration.Shell),
		huh.NewSelect[string]().Title("Prompt").Options(huh.NewOption("No prompt customization", "none"), huh.NewOption("Starship", "starship")).Value(&configuration.Prompt),
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

func environmentTemplateOptions(catalog templates.Catalog) ([]huh.Option[string], error) {
	available, err := catalog.List()
	if err != nil {
		return nil, fmt.Errorf("load environment templates: %w", err)
	}
	options := make([]huh.Option[string], 0, len(available)+1)
	options = append(options, huh.NewOption("No template", ""))
	for _, template := range available {
		options = append(options, huh.NewOption(template.Description, template.Name))
	}
	return options, nil
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
		fields = append(fields, [2]string{"Image", configuration.Image}, [2]string{"User", configuration.User}, [2]string{"Network", configuration.Network}, [2]string{"Template", configuration.Template}, [2]string{"Shell", configuration.Shell}, [2]string{"Prompt", configuration.Prompt}, [2]string{"Persistent home", strconv.FormatBool(configuration.Home.Enabled)}, [2]string{"Persistent caches", strconv.FormatBool(configuration.Caches.Enabled)}, [2]string{"Clipboard", strconv.FormatBool(configuration.Integrations.Clipboard)}, [2]string{"SSH agent", strconv.FormatBool(configuration.Integrations.SSHAgent)})
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
func terminalWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	return ok && terminal(file)
}
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
