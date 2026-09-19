package podman

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MuktadirHassan/box/internal/backend"
	"github.com/MuktadirHassan/box/internal/box"
	"github.com/MuktadirHassan/box/internal/templates"
)

func testIdentity() (int, int) { return 1000, 1001 }

type fakeCatalog struct{ resolved templates.Resolved }

func (c fakeCatalog) List() ([]templates.Descriptor, error) {
	return []templates.Descriptor{c.resolved.Descriptor()}, nil
}
func (c fakeCatalog) Resolve(string) (templates.Resolved, error) { return c.resolved, nil }

type fakeResolved struct {
	descriptor  templates.Descriptor
	validateErr error
	built       bool
}

func (r *fakeResolved) Descriptor() templates.Descriptor { return r.descriptor }
func (r *fakeResolved) Validate(templates.Request) error { return r.validateErr }
func (r *fakeResolved) BuildContext(dir string) error {
	r.built = true
	return os.WriteFile(filepath.Join(dir, "Containerfile"), []byte("FROM scratch\n"), 0644)
}

func TestCreateUsesCatalogBuildContextAndRejectsBeforeRunner(t *testing.T) {
	resolved := &fakeResolved{descriptor: templates.Descriptor{ID: "custom"}, validateErr: errors.New("incompatible")}
	runner := &fakeRunner{}
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: "custom:1", User: "dev", Network: "outbound", Template: "custom"}
	if _, err := New(Options{Runner: runner, Catalog: fakeCatalog{resolved}}).Create(context.Background(), definition); err == nil {
		t.Fatal("Create() error = nil")
	}
	if len(runner.outputCalls) != 0 || len(runner.runCalls) != 0 || resolved.built {
		t.Errorf("runner/build context used before compatibility rejection: output=%#v, run=%#v, built=%v", runner.outputCalls, runner.runCalls, resolved.built)
	}

	resolved.validateErr = nil
	runner.outputs = []outputResult{{output: "container\n"}}
	if _, err := New(Options{Runner: runner, Catalog: fakeCatalog{resolved}}).Create(context.Background(), definition); err != nil {
		t.Fatal(err)
	}
	if !resolved.built {
		t.Error("BuildContext was not called")
	}
	build := strings.Join(runner.runCalls[0].arguments, "\x00")
	if !strings.Contains(build, "BASE_IMAGE=custom:1") || !strings.Contains(build, "--file") || !strings.Contains(build, "Containerfile") {
		t.Errorf("build arguments = %v", build)
	}
}

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
	backend := New(Options{Runner: runner, Identity: testIdentity})
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
	for _, want := range []string{"--user\x001000:1001", "dev:x:1000:1001::/home/dev:/bin/sh", "HOME=/home/dev", "XDG_RUNTIME_DIR=/home/dev", "type=volume,src=box-demo-cache,dst=/home/dev/.cache,rw,U=true", "type=bind,src=" + mount + ",dst=/workspace,rw,nosuid,nodev"} {
		if !strings.Contains(joined, want) {
			t.Errorf("arguments missing %q: %#v", want, arguments)
		}
	}
	if !strings.Contains(joined, "--userns\x00keep-id") {
		t.Errorf("arguments do not preserve the configured identity mapping: %#v", arguments)
	}
	for _, forbidden := range []string{"--read-only", "--cap-drop", "no-new-privileges", "--privileged", "--pid=host", "--network=host", "podman.sock", "DOCKER_HOST", "CONTAINER_HOST"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("arguments include unsupported restriction or host access %q: %#v", forbidden, arguments)
		}
	}
}

func TestCreateBuildsSelectedTemplate(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "container-id\n"}}}
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: "ubuntu:24.04", User: "dev", Network: "outbound", Template: "ubuntu-24.04-terminal-tools", Shell: "fish", Prompt: "starship"}

	if _, err := New(Options{Runner: runner, Identity: testIdentity}).Create(context.Background(), definition); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(runner.runCalls) != 1 {
		t.Fatalf("Run calls = %d, want 1", len(runner.runCalls))
	}
	if len(runner.outputCalls) != 1 {
		t.Fatalf("Output calls = %d, want 1", len(runner.outputCalls))
	}
	build := runner.runCalls[0].arguments
	joinedBuild := strings.Join(build, "\x00")
	if !strings.Contains(joinedBuild, "build\x00--build-arg\x00BASE_IMAGE=ubuntu:24.04") || strings.Contains(joinedBuild, "--quiet") || !strings.Contains(joinedBuild, "BOX_USER=dev") || !strings.Contains(joinedBuild, "BOX_UID=1000") || !strings.Contains(joinedBuild, "BOX_GID=1001") || !strings.Contains(joinedBuild, "BOX_SHELL=fish") || !strings.Contains(joinedBuild, "BOX_PROMPT=starship") || !strings.Contains(joinedBuild, "BOX_INSECURE_MODE=false") || !strings.Contains(joinedBuild, "BOX_TEMPLATE_REVISION=0") || !strings.Contains(joinedBuild, "--tag\x00box-demo-template") {
		t.Errorf("build arguments = %#v", build)
	}
	create := strings.Join(runner.outputCalls[0].arguments, "\x00")
	if !strings.Contains(create, "box-demo-template\x00/usr/bin/fish") {
		t.Errorf("create arguments do not use template image: %#v", runner.outputCalls[0].arguments)
	}
	if !strings.Contains(create, "SHELL=/usr/bin/fish") || strings.Contains(create, "--passwd-entry") {
		t.Errorf("create arguments do not rely on the template's fish user consistently: %#v", runner.outputCalls[1].arguments)
	}
}

func TestBuildTemplatePassesInsecureMode(t *testing.T) {
	runner := &fakeRunner{}
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{
		Image: "ubuntu:24.04", User: "dev", Template: "ubuntu-24.04-terminal-tools", Shell: "sh", Prompt: "none",
		Integrations: box.Integrations{InsecureMode: true},
	}

	if _, err := New(Options{Runner: runner, Identity: testIdentity}).buildTemplate(context.Background(), definition); err != nil {
		t.Fatal(err)
	}
	if len(runner.runCalls) != 1 {
		t.Fatalf("Run calls = %d, want 1", len(runner.runCalls))
	}
	if !strings.Contains(strings.Join(runner.runCalls[0].arguments, "\x00"), "BOX_INSECURE_MODE=true") {
		t.Errorf("build arguments do not enable the Podman package: %#v", runner.runCalls[0].arguments)
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

func TestCreateRejectsTemplateOnlyShellsWithoutTemplate(t *testing.T) {
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{Image: "alpine:latest", User: "dev", Network: "outbound", Shell: "fish"}
	if _, err := New(Options{Runner: &fakeRunner{}}).Create(context.Background(), definition); err == nil {
		t.Fatal("Create() error = nil for fish without an environment template")
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
	clipboardArguments, err := backend.withClipboard(nil, "/home/dev")
	if err != nil {
		t.Fatalf("withClipboard() error = %v", err)
	}
	clipboardJoined := strings.Join(clipboardArguments, "\x00")
	if strings.Contains(clipboardJoined, "XDG_RUNTIME_DIR=") || !strings.Contains(clipboardJoined, "dst=/home/dev/agent.sock") {
		t.Errorf("clipboard arguments do not preserve the configured runtime directory: %#v", clipboardArguments)
	}
	if _, err := backend.withSSHAgent(nil); err != nil {
		t.Fatalf("withSSHAgent() error = %v", err)
	}
	invalid := New(Options{Env: func(string) string { return "/tmp/not-a-socket" }})
	if _, err := invalid.withSSHAgent(nil); err == nil {
		t.Error("withSSHAgent() error = nil for regular path")
	}
}

func TestClipboardAndInsecureModeUseIndependentSocketPaths(t *testing.T) {
	runtimeDirectory := t.TempDir()
	if err := os.Chmod(runtimeDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	waylandSocket := filepath.Join(runtimeDirectory, "wayland-1")
	waylandListener, err := net.Listen("unix", waylandSocket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { waylandListener.Close() })
	podmanDirectory := filepath.Join(runtimeDirectory, "podman")
	if err := os.Mkdir(podmanDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	podmanSocket := filepath.Join(podmanDirectory, "podman.sock")
	podmanListener, err := net.Listen("unix", podmanSocket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { podmanListener.Close() })

	runner := &fakeRunner{outputs: []outputResult{{output: "container-id\n"}}}
	backend := New(Options{
		Runner:   runner,
		Identity: func() (int, int) { return os.Getuid(), os.Getgid() },
		Env: func(name string) string {
			return map[string]string{
				"XDG_RUNTIME_DIR": runtimeDirectory,
				"WAYLAND_DISPLAY": "wayland-1",
			}[name]
		},
	})
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{
		Image: "ubuntu:24.04", User: "dev", Network: "outbound",
		Integrations: box.Integrations{Clipboard: true, InsecureMode: true},
	}
	if _, err := backend.Create(context.Background(), definition); err != nil {
		t.Fatal(err)
	}

	joined := strings.Join(runner.outputCalls[0].arguments, "\x00")
	for _, want := range []string{
		"XDG_RUNTIME_DIR=/home/dev",
		"WAYLAND_DISPLAY=wayland-1",
		"type=bind,src=" + waylandSocket + ",dst=/home/dev/wayland-1,rw,nosuid,nodev",
		"CONTAINER_HOST=unix:///tmp/podman.sock",
		"type=bind,src=" + podmanSocket + ",dst=/tmp/podman.sock,rw,nosuid,nodev",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("arguments missing %q: %#v", want, runner.outputCalls[0].arguments)
		}
	}
}

func TestInsecureModeExposesOnlyHostRootlessPodmanSocket(t *testing.T) {
	runtimeDirectory := t.TempDir()
	if err := os.Chmod(runtimeDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	socketDirectory := filepath.Join(runtimeDirectory, "podman")
	if err := os.MkdirAll(socketDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(socketDirectory, "podman.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })

	runner := &fakeRunner{outputs: []outputResult{{output: "container-id\n"}}}
	backend := New(Options{
		Runner: runner,
		Identity: func() (int, int) {
			return os.Getuid(), os.Getgid()
		},
		Env: func(name string) string {
			if name == "XDG_RUNTIME_DIR" {
				return runtimeDirectory
			}
			return ""
		},
	})
	definition := box.NewDefinition("demo")
	definition.Configuration = box.Configuration{
		Image: "ubuntu:24.04", User: "dev", Network: "outbound",
		Integrations: box.Integrations{InsecureMode: true},
	}
	if _, err := backend.Create(context.Background(), definition); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	joined := strings.Join(runner.outputCalls[0].arguments, "\x00")
	for _, want := range []string{
		"DOCKER_HOST=unix:///tmp/podman.sock",
		"CONTAINER_HOST=unix:///tmp/podman.sock",
		"type=bind,src=" + socket + ",dst=/tmp/podman.sock,rw,nosuid,nodev",
		"--userns\x00keep-id",
		"--network\x00pasta",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("arguments missing %q: %#v", want, runner.outputCalls[0].arguments)
		}
	}
	for _, forbidden := range []string{"--privileged", "--pid=host", "--network=host"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("arguments include forbidden host access %q: %#v", forbidden, runner.outputCalls[0].arguments)
		}
	}
}

func TestInsecureModeRejectsUnavailableOrUnsafeSocket(t *testing.T) {
	t.Run("relative runtime directory", func(t *testing.T) {
		backend := New(Options{Env: func(string) string { return "relative" }})
		if _, err := backend.withHostPodmanSocket(nil); err == nil || !strings.Contains(err.Error(), "systemctl --user enable --now podman.socket") {
			t.Errorf("withHostPodmanSocket() error = %v", err)
		}
	})

	t.Run("foreign-owned runtime directory", func(t *testing.T) {
		runtimeDirectory := t.TempDir()
		backend := New(Options{
			Env:      func(string) string { return runtimeDirectory },
			Identity: func() (int, int) { return os.Getuid() + 1, os.Getgid() },
		})
		if _, err := backend.withHostPodmanSocket(nil); err == nil || !strings.Contains(err.Error(), "owned by UID") {
			t.Errorf("withHostPodmanSocket() error = %v", err)
		}
	})

	t.Run("world-writable runtime directory", func(t *testing.T) {
		runtimeDirectory := t.TempDir()
		if err := os.Chmod(runtimeDirectory, 0777); err != nil {
			t.Fatal(err)
		}
		backend := New(Options{Env: func(string) string { return runtimeDirectory }})
		if _, err := backend.withHostPodmanSocket(nil); err == nil || !strings.Contains(err.Error(), "XDG_RUNTIME_DIR must not be accessible") {
			t.Errorf("withHostPodmanSocket() error = %v", err)
		}
	})

	t.Run("parent directory symlink", func(t *testing.T) {
		target := t.TempDir()
		if err := os.Mkdir(filepath.Join(target, "runtime"), 0700); err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		if err := os.Symlink(target, filepath.Join(root, "linked")); err != nil {
			t.Fatal(err)
		}
		runtimeDirectory := filepath.Join(root, "linked", "runtime")
		backend := New(Options{Env: func(string) string { return runtimeDirectory }})
		if _, err := backend.withHostPodmanSocket(nil); err == nil || !strings.Contains(err.Error(), "cannot contain symlinks") {
			t.Errorf("withHostPodmanSocket() error = %v", err)
		}
	})

	for _, test := range []struct {
		name  string
		setup func(t *testing.T, socket string)
	}{
		{name: "absent", setup: func(*testing.T, string) {}},
		{name: "regular file", setup: func(t *testing.T, socket string) {
			if err := os.WriteFile(socket, []byte("not a socket"), 0600); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "world-writable Podman directory", setup: func(t *testing.T, socket string) {
			if err := os.Chmod(filepath.Dir(socket), 0777); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "world-writable socket", setup: func(t *testing.T, socket string) {
			listener, err := net.Listen("unix", socket)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { listener.Close() })
			if err := os.Chmod(socket, 0777); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "symlink", setup: func(t *testing.T, socket string) {
			target := filepath.Join(filepath.Dir(socket), "real.sock")
			listener, err := net.Listen("unix", target)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { listener.Close() })
			if err := os.Symlink(target, socket); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			runtimeDirectory := t.TempDir()
			socketDirectory := filepath.Join(runtimeDirectory, "podman")
			if err := os.MkdirAll(socketDirectory, 0755); err != nil {
				t.Fatal(err)
			}
			test.setup(t, filepath.Join(socketDirectory, "podman.sock"))
			backend := New(Options{Env: func(string) string { return runtimeDirectory }})
			if _, err := backend.withHostPodmanSocket(nil); err == nil || !strings.Contains(err.Error(), "enable insecure mode") || !strings.Contains(err.Error(), "systemctl --user enable --now podman.socket") {
				t.Errorf("withHostPodmanSocket() error = %v", err)
			}
		})
	}
}

func TestEnterOnlyStartsKnownStoppedStates(t *testing.T) {
	runner := &fakeRunner{outputs: []outputResult{{output: "exited\n"}}}
	metadata := box.RuntimeMetadata{Backend: box.PodmanBackend, ID: "container-id"}
	if err := New(Options{Runner: runner}).Enter(context.Background(), box.Definition{Configuration: box.DefaultConfiguration()}, metadata); err != nil {
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
