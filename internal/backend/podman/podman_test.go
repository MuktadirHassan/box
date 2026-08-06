package podman

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
)

type commandCall struct{ arguments []string }
type outputResult struct {
	output string
	err    error
}
type fakeRunner struct {
	outputs               []outputResult
	outputCalls, runCalls []commandCall
	runErr                error
}

func (r *fakeRunner) Output(_ context.Context, arguments ...string) (string, error) {
	r.outputCalls = append(r.outputCalls, commandCall{arguments})
	result := r.outputs[0]
	r.outputs = r.outputs[1:]
	return result.output, result.err
}
func (r *fakeRunner) Run(_ context.Context, arguments ...string) error {
	r.runCalls = append(r.runCalls, commandCall{arguments})
	return r.runErr
}

func TestValidateRequiresRootlessPodman(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "true\n"}}}
	if err := New(Options{Runner: runner}).Validate(context.Background()); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestCreateUsesConfiguredIdentityNotHostIdentity(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "container-id\n"}}}
	mount := t.TempDir()
	backend := New(Options{Runner: runner})
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: "ubuntu:24.04", User: "dev", Home: box.Persistence{Enabled: true}, Caches: box.Persistence{Enabled: true}, Network: "none", Mounts: []box.Mount{{Source: mount, Destination: "/workspace"}}}
	metadata, err := backend.Create(context.Background(), definition)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if metadata.ID != "container-id" {
		t.Errorf("ID = %q", metadata.ID)
	}
	arguments := runner.outputCalls[0].arguments
	joined := strings.Join(arguments, "\x00")
	for _, want := range []string{fmt.Sprintf("--user\x00%d:%d", os.Getuid(), os.Getgid()), fmt.Sprintf("dev:x:%d:%d::/home/dev:/bin/sh", os.Getuid(), os.Getgid()), "HOME=/home/dev", "type=volume,src=box-demo-cache,dst=/home/dev/.cache,rw,U=true", "type=bind,src=" + mount + ",dst=/workspace,rw,nosuid,nodev"} {
		if !strings.Contains(joined, want) {
			t.Errorf("arguments missing %q: %#v", want, arguments)
		}
	}
	if !strings.Contains(joined, "--userns\x00keep-id") {
		t.Errorf("arguments do not preserve the invoking user's identity: %#v", arguments)
	}
}

func TestCreateBuildsSelectedTemplate(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "image-id\n"}, {output: "container-id\n"}}}
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: "ubuntu:24.04", User: "dev", Network: "outbound", Template: box.TerminalToolsTemplate}

	if _, err := New(Options{Runner: runner}).Create(context.Background(), definition); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(runner.outputCalls) != 2 {
		t.Fatalf("Output calls = %d, want 2", len(runner.outputCalls))
	}
	build := runner.outputCalls[0].arguments
	if len(build) < 6 || build[0] != "build" || build[1] != "--quiet" || build[4] != "--tag" || build[5] != "box-demo-template" {
		t.Errorf("build arguments = %#v", build)
	}
	create := strings.Join(runner.outputCalls[1].arguments, "\x00")
	if !strings.Contains(create, "box-demo-template\x00/bin/sh") {
		t.Errorf("create arguments do not use template image: %#v", runner.outputCalls[1].arguments)
	}
}

func TestTerminalToolsContainerfileInstallsRequestedPackages(t *testing.T) {
	for _, image := range []string{"ubuntu:24.04", "archlinux:latest"} {
		contents := templateContainerfile(image, box.TerminalToolsTemplate)
		for _, packageName := range box.TemplatePackages(box.TerminalToolsTemplate) {
			if !strings.Contains(contents, packageName) {
				t.Errorf("Containerfile for %s does not install %q: %s", image, packageName, contents)
			}
		}
	}
	if !strings.Contains(templateContainerfile("archlinux:latest", box.TerminalToolsTemplate), "pacman -Syu --noconfirm --needed") {
		t.Error("Arch Containerfile does not use pacman")
	}
	if !templateSupportedImage("debian:bookworm") || !templateSupportedImage("archlinux:latest") || templateSupportedImage("fedora:latest") {
		t.Error("templateSupportedImage() does not identify supported base images")
	}
}

func TestCreateRejectsUnsafeConfiguration(t *testing.T) {
	backend := New(Options{Runner: &fakeRunner{}})
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: " image", User: "dev", Network: "outbound"}
	if _, err := backend.Create(context.Background(), definition); err == nil {
		t.Error("Create() error = nil for invalid image")
	}
	definition.Configuration.Image = "ubuntu:24.04"
	definition.Configuration.Network = "accidentally-open"
	if _, err := backend.Create(context.Background(), definition); err == nil {
		t.Error("Create() error = nil for invalid network")
	}
	definition.Configuration.Network = "none"
	definition.Configuration.Mounts = []box.Mount{{Source: "relative", Destination: "/workspace"}}
	if _, err := backend.Create(context.Background(), definition); err == nil {
		t.Error("Create() error = nil for invalid mount")
	}
}

func TestCreateRejectsInvalidRuntimeID(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "\n"}}}
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: "ubuntu:24.04", User: "dev", Network: "none"}
	if _, err := New(Options{Runner: runner}).Create(context.Background(), definition); err == nil {
		t.Error("Create() error = nil for empty runtime ID")
	}
}

func TestIntegrationsRequireRealSockets(t *testing.T) {
	directory := t.TempDir()
	socket := filepath.Join(directory, "agent.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	backend := New(Options{Env: func(name string) string {
		return map[string]string{"XDG_RUNTIME_DIR": directory, "WAYLAND_DISPLAY": "agent.sock", "SSH_AUTH_SOCK": socket}[name]
	}})
	if _, err := backend.withClipboard(nil); err != nil {
		t.Fatalf("withClipboard() error = %v", err)
	}
	if _, err := backend.withSSHAgent(nil); err != nil {
		t.Fatalf("withSSHAgent() error = %v", err)
	}
	invalid := New(Options{Env: func(string) string { return "/tmp/not-a-socket" }})
	if _, err := invalid.withSSHAgent(nil); err == nil {
		t.Error("withSSHAgent() error = nil for regular path")
	}
}

func TestEnterOnlyStartsKnownStoppedStates(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "exited\n"}}}
	metadata := box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "container-id"}
	if err := New(Options{Runner: runner}).Enter(context.Background(), metadata); err != nil {
		t.Fatal(err)
	}
	want := []commandCall{{arguments: []string{"start", "--attach", "--interactive", "container-id"}}}
	if !reflect.DeepEqual(runner.runCalls, want) {
		t.Errorf("Run calls = %#v, want %#v", runner.runCalls, want)
	}
}

func TestDeletePurgeRemovesOnlyEnabledManagedVolumes(t *testing.T) {
	runner := &fakeRunner{}
	definition := box.NewDefinition("demo")
	definition.Configuration.Home.Enabled = true
	definition.Configuration.Caches.Enabled = true
	metadata := box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "container-id"}
	if err := New(Options{Runner: runner}).Delete(context.Background(), definition, metadata, backend.DeleteOptions{RemovePersistentData: true}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	want := []commandCall{
		{arguments: []string{"rm", "--force", "container-id"}},
		{arguments: []string{"volume", "rm", "box-demo-home"}},
		{arguments: []string{"volume", "rm", "box-demo-cache"}},
	}
	if !reflect.DeepEqual(runner.runCalls, want) {
		t.Errorf("Run calls = %#v, want %#v", runner.runCalls, want)
	}
}

func TestDeleteForRecreationPreservesManagedVolumes(t *testing.T) {
	runner := &fakeRunner{}
	definition := box.NewDefinition("demo")
	definition.Configuration.Home.Enabled = true
	definition.Configuration.Caches.Enabled = true
	metadata := box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "container-id"}
	if err := New(Options{Runner: runner}).Delete(context.Background(), definition, metadata, backend.DeleteOptions{}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	want := []commandCall{{arguments: []string{"rm", "--force", "container-id"}}}
	if !reflect.DeepEqual(runner.runCalls, want) {
		t.Errorf("Run calls = %#v, want %#v", runner.runCalls, want)
	}
}

func TestExecReturnsRunnerFailure(t *testing.T) {
	runner := &fakeRunner{runErr: errors.New("podman failed")}
	err := New(Options{Runner: runner}).Exec(context.Background(), box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "container-id"}, []string{"git", "status"})
	if !errors.Is(err, runner.runErr) {
		t.Errorf("Exec() error = %v", err)
	}
}
