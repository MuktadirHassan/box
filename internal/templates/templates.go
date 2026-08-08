package templates

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/MuktadirHassan/box/templates"
	"github.com/pelletier/go-toml/v2"
)

var namePattern = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)

type Descriptor struct {
	ID, Name, Description     string
	Version                   int
	ImageFamily, ImageVersion string
	Shells, Prompts           []string
}
type Request struct{ Image, Shell, Prompt string }
type Resolved interface {
	Descriptor() Descriptor
	Validate(Request) error
	BuildContext(string) error
}
type Catalog interface {
	List() ([]Descriptor, error)
	Resolve(string) (Resolved, error)
}

type Template struct {
	Name          string `toml:"name"`
	Version       int    `toml:"version"`
	Description   string `toml:"description"`
	Family        string `toml:"family"`
	Compatibility struct {
		ImageFamily  string `toml:"image_family"`
		ImageVersion string `toml:"image_version"`
	} `toml:"compatibility"`
	Capabilities struct {
		Shells  []string `toml:"shells"`
		Prompts []string `toml:"prompts"`
	} `toml:"capabilities"`
	root  string
	files fs.FS
}

func (t Template) Descriptor() Descriptor {
	return Descriptor{t.Name, t.Name, t.Description, t.Version, t.Compatibility.ImageFamily, t.Compatibility.ImageVersion, append([]string(nil), t.Capabilities.Shells...), append([]string(nil), t.Capabilities.Prompts...)}
}
func (t Template) Validate(r Request) error {
	if r.Image == "" {
		return fmt.Errorf("image cannot be empty")
	}
	base, digest, ok := splitImage(r.Image)
	if !ok || strings.HasSuffix(base, ":latest") {
		return fmt.Errorf("template %q requires %s:%s image", t.Name, t.Compatibility.ImageFamily, t.Compatibility.ImageVersion)
	}
	colon := strings.LastIndexByte(base, ':')
	repository, tag := base[:colon], base[colon+1:]
	if path.Base(repository) != t.Compatibility.ImageFamily || tag != t.Compatibility.ImageVersion || (digest != "" && !validDigest(digest)) {
		return fmt.Errorf("template %q is incompatible with image %q", t.Name, r.Image)
	}
	if !contains(t.Capabilities.Shells, r.Shell) {
		return fmt.Errorf("template %q does not support shell %q", t.Name, r.Shell)
	}
	if !contains(t.Capabilities.Prompts, r.Prompt) {
		return fmt.Errorf("template %q does not support prompt %q", t.Name, r.Prompt)
	}
	return nil
}
func splitImage(image string) (string, string, bool) {
	if strings.TrimSpace(image) != image || strings.ContainsAny(image, " \t\r\n") {
		return "", "", false
	}
	base, digest, hasDigest := strings.Cut(image, "@")
	if hasDigest && (digest == "" || strings.Contains(digest, "@")) {
		return "", "", false
	}
	colon := strings.LastIndexByte(base, ':')
	slash := strings.LastIndexByte(base, '/')
	if colon <= slash || colon == len(base)-1 {
		return "", "", false
	}
	return base, digest, true
}
func validDigest(digest string) bool {
	algorithm, value, ok := strings.Cut(digest, ":")
	return ok && algorithm == "sha256" && len(value) == sha256.Size*2 && strings.Trim(value, "0123456789abcdefABCDEF") == ""
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
func (t Template) BuildContext(destination string) error {
	return fs.WalkDir(t.files, t.root, func(source string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(source, t.root), "/")
		target := path.Join(destination, rel)
		if e.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := fs.ReadFile(t.files, source)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

type EmbeddedCatalog struct{ files fs.FS }

// NewRegistry combines catalog providers and rejects invalid registrations.
func NewRegistry(providers ...Catalog) (Catalog, error) {
	seen := map[string]struct{}{}
	for i, provider := range providers {
		if provider == nil || isNilCatalog(provider) {
			return nil, fmt.Errorf("template catalog provider %d is nil", i)
		}
		descriptors, err := provider.List()
		if err != nil {
			return nil, fmt.Errorf("list template catalog provider %d: %w", i, err)
		}
		for _, descriptor := range descriptors {
			if _, exists := seen[descriptor.ID]; exists {
				return nil, fmt.Errorf("duplicate template %q", descriptor.ID)
			}
			seen[descriptor.ID] = struct{}{}
		}
	}
	return registry{providers: append([]Catalog(nil), providers...)}, nil
}

type registry struct{ providers []Catalog }

func isNilCatalog(provider Catalog) bool {
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	}
	return false
}

func (r registry) List() ([]Descriptor, error) {
	seen := map[string]bool{}
	var out []Descriptor
	for _, p := range r.providers {
		ds, err := p.List()
		if err != nil {
			return nil, err
		}
		for _, d := range ds {
			if seen[d.ID] {
				return nil, fmt.Errorf("duplicate template %q", d.ID)
			}
			seen[d.ID] = true
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (r registry) Resolve(id string) (Resolved, error) {
	for _, p := range r.providers {
		if v, err := p.Resolve(id); err == nil {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unsupported template %q", id)
}

func NewEmbeddedCatalog(assets fs.FS) Catalog { return EmbeddedCatalog{files: assets} }
func (c EmbeddedCatalog) Resolve(id string) (Resolved, error) {
	if !namePattern.MatchString(id) {
		return nil, fmt.Errorf("invalid template name %q", id)
	}
	return c.load(id)
}
func (c EmbeddedCatalog) List() ([]Descriptor, error) {
	entries, err := fs.ReadDir(c.files, ".")
	if err != nil {
		return nil, err
	}
	out := []Descriptor{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t, err := c.load(e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, t.Descriptor())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (c EmbeddedCatalog) load(id string) (Template, error) {
	root := id
	data, err := fs.ReadFile(c.files, path.Join(root, "template.toml"))
	if err != nil {
		return Template{}, fmt.Errorf("unsupported template %q", id)
	}
	var t Template
	if err = toml.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("decode manifest: %w", err)
	}
	if t.Name != id || t.Version < 1 || t.Description == "" || t.Compatibility.ImageFamily == "" || t.Compatibility.ImageVersion == "" || len(t.Capabilities.Shells) == 0 || len(t.Capabilities.Prompts) == 0 {
		return t, fmt.Errorf("manifest is invalid")
	}
	t.root = root
	t.files = c.files
	for _, asset := range []string{"Containerfile", "initialize-home", "dotfiles"} {
		info, statErr := fs.Stat(c.files, path.Join(root, asset))
		if statErr != nil {
			return t, fmt.Errorf("%s: %w", asset, statErr)
		}
		if asset == "dotfiles" && !info.IsDir() {
			return t, fmt.Errorf("%s is not a directory", asset)
		}
	}
	return t, nil
}
func Resolve(name, image string) (Template, error) {
	if name == "" {
		return Template{}, nil
	}
	r, err := NewEmbeddedCatalog(templates.FS()).Resolve(name)
	if err != nil {
		return Template{}, err
	}
	t := r.(Template)
	if err = t.Validate(Request{Image: image, Shell: "sh", Prompt: "none"}); err != nil {
		return Template{}, err
	}
	return t, nil
}
func ValidateCompatibility(name, image string) error {
	if name == "" {
		return nil
	}
	_, err := Resolve(name, image)
	return err
}
func imageFamily(image string) string {
	image = strings.Split(image, "@")[0]
	base := path.Base(image)
	family, _, _ := strings.Cut(base, ":")
	return family
}
func All() ([]Template, error) {
	ds, err := NewEmbeddedCatalog(templates.FS()).List()
	if err != nil {
		return nil, err
	}
	out := make([]Template, 0, len(ds))
	catalog := NewEmbeddedCatalog(templates.FS()).(EmbeddedCatalog)
	for _, d := range ds {
		t, _ := catalog.load(d.ID)
		out = append(out, t)
	}
	return out, nil
}
