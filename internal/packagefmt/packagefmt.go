// Package packagefmt implements the on-disk package format used by mgs3mod.
// It deliberately keeps the format small: one manifest and one payload per
// manifest entry.
package packagefmt

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"mgs3mod/internal/profile"
	"mgs3mod/internal/winfs"
)

const (
	maxManifestBytes  = 1 << 20
	maxMappings       = 256
	maxPayloadBytes   = 64 << 20
	maxPayloadTotal   = 512 << 20
	maxPackageEntries = 512
)

var (
	idPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?(?:\+[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?$`)
	hexPattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Manifest is the strict JSON manifest in a package.
type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	ID            string `json:"id"`
	Version       string `json:"version"`
	Name          string `json:"name"`
	Profile       string `json:"profile"`
	Files         []File `json:"files"`
}

// File maps one package payload to one managed game target.
type File struct {
	Source         string `json:"source"`
	Target         string `json:"target"`
	OriginalSHA256 string `json:"originalSha256"`
	OriginalAbsent bool   `json:"originalAbsent,omitempty"`
	PayloadSHA256  string `json:"payloadSha256"`
	PayloadBytes   int64  `json:"payloadBytes"`
}

// Package is a validated package and its payload bytes, keyed by File.Source.
type Package struct {
	Manifest Manifest
	Payloads map[string][]byte
}

// Parse parses and validates a manifest. authoring permits an absent or zero
// payload hash/size; Pack and Load fill those values from payload bytes.
func Parse(data []byte, authoring bool) (Manifest, error) {
	if len(data) > maxManifestBytes {
		return Manifest{}, fmt.Errorf("manifest exceeds %d bytes", maxManifestBytes)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return Manifest{}, errors.New("manifest is empty")
	}
	if err := scanJSON(data); err != nil {
		return Manifest{}, fmt.Errorf("manifest JSON: %w", err)
	}
	if err := validateJSONFieldNames(data); err != nil {
		return Manifest{}, fmt.Errorf("manifest fields: %w", err)
	}
	var m Manifest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("manifest JSON: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return Manifest{}, errors.New("manifest has trailing JSON")
		}
		return Manifest{}, fmt.Errorf("manifest trailing data: %w", err)
	}
	if err := validateManifest(&m, authoring); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// Load reads and validates either a package folder or a normalized ZIP.
func Load(path string, authoring bool) (*Package, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("package path %q: %w", path, err)
	}
	path = absolute
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("package %q: %w", path, err)
	}
	if info.IsDir() {
		return loadFolder(path, authoring)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("package %q is not a regular file or directory", path)
	}
	if err := winfs.CheckPath(filepath.Dir(path), filepath.Base(path), false); err != nil {
		return nil, fmt.Errorf("package archive: %w", err)
	}
	return loadZIP(path, authoring)
}

// Pack validates authoring folder and writes a deterministic ZIP. Existing
// output is never replaced.
func Pack(folder, out string) (*Package, error) {
	absFolder, err := filepath.Abs(folder)
	if err != nil {
		return nil, fmt.Errorf("package folder %q: %w", folder, err)
	}
	absOut, err := filepath.Abs(out)
	if err != nil {
		return nil, fmt.Errorf("package output %q: %w", out, err)
	}
	folder, out = absFolder, absOut
	p, err := loadFolder(folder, true)
	if err != nil {
		return nil, err
	}
	if err := validateManifest(&p.Manifest, false); err != nil {
		return nil, err
	}
	if err := CheckOutput(out); err != nil {
		return nil, err
	}
	parent := filepath.Dir(out)
	if err := winfs.CheckDir(parent); err != nil {
		return nil, fmt.Errorf("package output parent: %w", err)
	}
	outputRoot, err := os.OpenRoot(parent)
	if err != nil {
		return nil, fmt.Errorf("open package output parent: %w", err)
	}
	defer outputRoot.Close()
	base := filepath.Base(out)
	f, err := outputRoot.OpenFile(base, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create package output: %w", err)
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = outputRoot.Remove(base)
		}
	}()
	w := zip.NewWriter(f)
	manifest, err := json.Marshal(p.Manifest)
	if err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	if err := writeZipFile(w, "manifest.json", manifest); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(p.Payloads))
	for key := range p.Payloads {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writeZipFile(w, key, p.Payloads[key]); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close package ZIP: %w", err)
	}
	if err := f.Sync(); err != nil {
		return nil, fmt.Errorf("sync package ZIP: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close package output: %w", err)
	}
	ok = true
	return p, nil
}

// CheckOutput validates a proposed output without creating it. Live Pack repeats
// this check and creates the destination exclusively to handle concurrent writes.
func CheckOutput(out string) (err error) {
	defer func() {
		if err != nil {
			err = &validationError{err}
		}
	}()
	absolute, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	parent, name := filepath.Dir(absolute), filepath.Base(absolute)
	if err = profile.Path(name); err != nil {
		return err
	}
	if err = winfs.CheckDir(parent); err != nil {
		return err
	}
	if err = winfs.CheckPath(parent, name, true); err != nil {
		return err
	}
	return refuseExisting(absolute)
}

type validationError struct{ err error }

func (e *validationError) Error() string { return e.err.Error() }
func (e *validationError) Unwrap() error { return e.err }

// ErrorCode distinguishes invalid packages/preflight destinations from actual
// operating-system I/O failures while retaining the underlying diagnostic.
func ErrorCode(err error) int {
	var validation *validationError
	if errors.As(err, &validation) && errors.Is(err, os.ErrNotExist) {
		return 2
	}
	var pathError *os.PathError
	var systemError syscall.Errno
	if errors.As(err, &pathError) || errors.As(err, &systemError) {
		return 1
	}
	return 2
}

// Digest returns a deterministic SHA-256 identity for complete package
// content. Payload map insertion order does not affect it.
func Digest(p *Package) string {
	if p == nil {
		return ""
	}
	h := sha256.New()
	manifest, _ := json.Marshal(p.Manifest)
	writeDigestPart(h, manifest)
	keys := make([]string, 0, len(p.Payloads))
	for key := range p.Payloads {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		writeDigestPart(h, []byte(key))
		writeDigestPart(h, p.Payloads[key])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func writeDigestPart(w io.Writer, data []byte) {
	var n [8]byte
	for i := range n {
		n[7-i] = byte(uint64(len(data)) >> (i * 8))
	}
	_, _ = w.Write(n[:])
	_, _ = w.Write(data)
}

func writeZipFile(w *zip.Writer, name string, data []byte) error {
	h := &zip.FileHeader{Name: name, Method: zip.Store}
	h.SetMode(0o644)
	h.SetModTime(time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC))
	entry, err := w.CreateHeader(h)
	if err != nil {
		return fmt.Errorf("create ZIP entry %q: %w", name, err)
	}
	if _, err := entry.Write(data); err != nil {
		return fmt.Errorf("write ZIP entry %q: %w", name, err)
	}
	return nil
}

func refuseExisting(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("output %q already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check output %q: %w", path, err)
	}
	return nil
}

func loadFolder(folder string, authoring bool) (*Package, error) {
	if err := winfs.CheckDir(folder); err != nil {
		return nil, fmt.Errorf("package folder: %w", err)
	}
	root, err := os.OpenRoot(folder)
	if err != nil {
		return nil, fmt.Errorf("open package folder: %w", err)
	}
	defer root.Close()
	var manifestBytes []byte
	files := make(map[string]struct{})
	dirs := make(map[string]struct{})
	entries := 0
	err = fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		entries++
		if entries > maxPackageEntries {
			return fmt.Errorf("package has more than %d entries", maxPackageEntries)
		}
		if err := validatePackagePath(path); err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("package path %q is a link", path)
		}
		if d.IsDir() {
			if err := winfs.CheckDir(filepath.Join(folder, filepath.FromSlash(path))); err != nil {
				return fmt.Errorf("package directory %q: %w", path, err)
			}
			dirs[path] = struct{}{}
			return nil
		}
		if err := winfs.CheckPath(folder, path, false); err != nil {
			return fmt.Errorf("package path %q: %w", path, err)
		}
		if path == "manifest.json" {
			manifestBytes, err = readRootLimited(root, path, maxManifestBytes)
			return err
		}
		if !strings.HasPrefix(path, "payload/") {
			return fmt.Errorf("unmapped package file %q", path)
		}
		if _, exists := files[path]; exists {
			return fmt.Errorf("duplicate package file %q", path)
		}
		files[path] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read package folder: %w", err)
	}
	if manifestBytes == nil {
		return nil, errors.New("package folder lacks manifest.json")
	}
	m, err := Parse(manifestBytes, authoring)
	if err != nil {
		return nil, err
	}
	p := &Package{Manifest: m, Payloads: make(map[string][]byte, len(m.Files))}
	if err := readMappedFolder(root, folder, files, dirs, p, authoring); err != nil {
		return nil, err
	}
	return p, nil
}

func readMappedFolder(root *os.Root, folder string, entries, dirs map[string]struct{}, p *Package, authoring bool) error {
	wanted := make(map[string]struct{}, len(p.Manifest.Files))
	knownDirs := map[string]struct{}{"payload": {}}
	var total int64
	for i := range p.Manifest.Files {
		f := &p.Manifest.Files[i]
		if _, exists := wanted[f.Source]; exists {
			return fmt.Errorf("duplicate payload mapping %q", f.Source)
		}
		wanted[f.Source] = struct{}{}
		if _, exists := entries[f.Source]; !exists {
			return fmt.Errorf("manifest payload %q is missing", f.Source)
		}
		data, err := readRootLimited(root, f.Source, maxPayloadBytes)
		if err != nil {
			return fmt.Errorf("read payload %q: %w", f.Source, err)
		}
		p.Payloads[f.Source] = data
		total += int64(len(data))
		if total > maxPayloadTotal {
			return fmt.Errorf("payloads exceed %d bytes total", maxPayloadTotal)
		}
		for parent := pathDir(f.Source); parent != "" && parent != "."; parent = pathDir(parent) {
			knownDirs[parent] = struct{}{}
		}
		if err := fillOrCheckPayload(f, data, authoring); err != nil {
			return fmt.Errorf("payload %q: %w", f.Source, err)
		}
	}
	for entry := range entries {
		if _, exists := wanted[entry]; !exists {
			return fmt.Errorf("unmapped package file %q", entry)
		}
	}
	for dir := range dirs {
		if _, exists := knownDirs[dir]; !exists {
			return fmt.Errorf("unmapped package directory %q", dir)
		}
	}
	if err := validateManifest(&p.Manifest, false); err != nil {
		return err
	}
	return nil
}

func readRootLimited(root *os.Root, path string, limit int64) ([]byte, error) {
	f, err := root.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("file exceeds %d bytes", limit)
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds %d bytes", limit)
	}
	return data, nil
}

func loadZIP(path string, authoring bool) (*Package, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open package ZIP: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, fmt.Errorf("read package ZIP: %w", err)
	}
	if len(zr.File) > maxPackageEntries {
		return nil, fmt.Errorf("package ZIP has more than %d entries", maxPackageEntries)
	}
	var manifestBytes []byte
	entries := make(map[string]*zip.File, len(zr.File))
	dirs := make(map[string]struct{})
	canonical := make(map[string]string, len(zr.File))
	for _, entry := range zr.File {
		name := entry.Name
		if err := validateZipName(name); err != nil {
			return nil, err
		}
		key := strings.ToLower(name)
		if prior, exists := canonical[key]; exists {
			return nil, fmt.Errorf("case-colliding or duplicate ZIP entries %q and %q", prior, name)
		}
		canonical[key] = name
		if entry.Flags&1 != 0 {
			return nil, fmt.Errorf("encrypted ZIP entry %q", name)
		}
		mode := entry.Mode()
		isDir := strings.HasSuffix(name, "/")
		if mode&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("ZIP entry %q is a link", name)
		}
		if isDir {
			if mode != 0 && !mode.IsDir() {
				return nil, fmt.Errorf("ZIP directory entry %q has regular-file mode", name)
			}
			dirs[strings.TrimSuffix(name, "/")] = struct{}{}
			continue
		}
		if mode.IsDir() {
			return nil, fmt.Errorf("ZIP file entry %q has directory mode", name)
		}
		if name == "manifest.json" {
			if entry.UncompressedSize64 > maxManifestBytes {
				return nil, fmt.Errorf("manifest exceeds %d bytes", maxManifestBytes)
			}
			manifestBytes, err = readZipEntry(entry, maxManifestBytes)
			if err != nil {
				return nil, fmt.Errorf("read manifest: %w", err)
			}
			continue
		}
		if !strings.HasPrefix(name, "payload/") {
			return nil, fmt.Errorf("unmapped ZIP file %q", name)
		}
		if entry.UncompressedSize64 > maxPayloadBytes {
			return nil, fmt.Errorf("payload %q exceeds %d bytes", name, maxPayloadBytes)
		}
		entries[name] = entry
	}
	if manifestBytes == nil {
		return nil, errors.New("package ZIP lacks manifest.json")
	}
	m, err := Parse(manifestBytes, authoring)
	if err != nil {
		return nil, err
	}
	p := &Package{Manifest: m, Payloads: make(map[string][]byte, len(m.Files))}
	var total int64
	for i := range p.Manifest.Files {
		file := &p.Manifest.Files[i]
		entry, exists := entries[file.Source]
		if !exists {
			return nil, fmt.Errorf("manifest payload %q is missing", file.Source)
		}
		data, err := readZipEntry(entry, maxPayloadBytes)
		if err != nil {
			return nil, fmt.Errorf("read payload %q: %w", file.Source, err)
		}
		total += int64(len(data))
		if total > maxPayloadTotal {
			return nil, fmt.Errorf("payloads exceed %d bytes total", maxPayloadTotal)
		}
		p.Payloads[file.Source] = data
		if err := fillOrCheckPayload(file, data, authoring); err != nil {
			return nil, fmt.Errorf("payload %q: %w", file.Source, err)
		}
	}
	for name := range entries {
		if _, exists := p.Payloads[name]; !exists {
			return nil, fmt.Errorf("unmapped ZIP file %q", name)
		}
	}
	knownDirs := map[string]struct{}{"payload": {}}
	for _, file := range p.Manifest.Files {
		for parent := pathDir(file.Source); parent != "" && parent != "."; parent = pathDir(parent) {
			knownDirs[parent] = struct{}{}
		}
	}
	for dir := range dirs {
		if _, exists := knownDirs[dir]; !exists {
			return nil, fmt.Errorf("unmapped ZIP directory %q", dir)
		}
	}
	if err := validateManifest(&p.Manifest, false); err != nil {
		return nil, err
	}
	return p, nil
}

func readZipEntry(entry *zip.File, limit int64) ([]byte, error) {
	r, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("entry exceeds %d bytes", limit)
	}
	return data, nil
}

func fillOrCheckPayload(file *File, data []byte, authoring bool) error {
	if profile.AddedTarget(file.Target) == nil {
		if len(data) == 0 {
			return errors.New("plugin payload is empty")
		}
		parsed, err := pe.NewFile(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("plugin payload is not a valid PE: %w", err)
		}
		defer parsed.Close()
		if parsed.FileHeader.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
			return fmt.Errorf("plugin payload machine is %v, want AMD64", parsed.FileHeader.Machine)
		}
		if parsed.FileHeader.Characteristics&pe.IMAGE_FILE_DLL == 0 {
			return errors.New("plugin payload is not a DLL")
		}
		header, ok := parsed.OptionalHeader.(*pe.OptionalHeader64)
		if !ok {
			return errors.New("plugin payload is not PE32+")
		}
		if header.AddressOfEntryPoint == 0 || header.SizeOfCode == 0 || len(parsed.Sections) == 0 {
			return errors.New("plugin payload has no executable entry section")
		}
		entryInSection := false
		for _, section := range parsed.Sections {
			end := uint64(section.Offset) + uint64(section.Size)
			// Zero-sized sections with zero raw offset are legitimate .bss-style
			// uninitialized data and have no bytes to validate.
			if section.Size == 0 {
				if section.Offset != 0 {
					return errors.New("plugin payload section exceeds file bounds")
				}
			} else if end > uint64(len(data)) {
				return errors.New("plugin payload section exceeds file bounds")
			}
			virtualEnd := uint64(section.VirtualAddress) + uint64(section.VirtualSize)
			entry := uint64(header.AddressOfEntryPoint)
			if section.Characteristics&pe.IMAGE_SCN_MEM_EXECUTE != 0 && entry >= uint64(section.VirtualAddress) && entry < virtualEnd && entry-uint64(section.VirtualAddress) < uint64(section.Size) {
				entryInSection = true
			}
		}
		if !entryInSection {
			return errors.New("plugin payload entry point is outside executable file-backed sections")
		}
	}
	hash := sha256.Sum256(data)
	want := hex.EncodeToString(hash[:])
	if file.PayloadBytes < 0 || file.PayloadBytes > maxPayloadBytes {
		return fmt.Errorf("invalid payloadBytes %d", file.PayloadBytes)
	}
	if authoring {
		if file.PayloadSHA256 != "" && !isZeroHash(file.PayloadSHA256) && file.PayloadSHA256 != want {
			return fmt.Errorf("payloadSha256 does not match payload")
		}
		if file.PayloadBytes != 0 && file.PayloadBytes != int64(len(data)) {
			return fmt.Errorf("payloadBytes does not match payload")
		}
		file.PayloadSHA256 = want
		file.PayloadBytes = int64(len(data))
		return nil
	}
	if file.PayloadSHA256 != want {
		return fmt.Errorf("payloadSha256 does not match payload")
	}
	if file.PayloadBytes != int64(len(data)) {
		return fmt.Errorf("payloadBytes does not match payload")
	}
	return nil
}

func isZeroHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

func validateManifest(m *Manifest, authoring bool) error {
	if m.SchemaVersion != 1 && m.SchemaVersion != 2 && m.SchemaVersion != 3 {
		return fmt.Errorf("unsupported schemaVersion %d", m.SchemaVersion)
	}
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("invalid id %q", m.ID)
	}
	if !versionPattern.MatchString(m.Version) {
		return fmt.Errorf("invalid version %q", m.Version)
	}
	if !validText(m.Name, 256) {
		return errors.New("invalid name")
	}
	if m.Profile != profile.ID {
		return fmt.Errorf("unsupported profile %q", m.Profile)
	}
	if len(m.Files) == 0 {
		return errors.New("package has no files")
	}
	if len(m.Files) > maxMappings {
		return fmt.Errorf("package has more than %d mappings", maxMappings)
	}
	if m.SchemaVersion == 3 && (m.ID != profile.LoaderID || m.Version != profile.LoaderVersion || len(m.Files) != len(profile.LoaderFiles)) {
		return errors.New("schemaVersion 3 requires the pinned asi-loader package identity and complete target set")
	}
	targets := make(map[string]string, len(m.Files))
	sources := make(map[string]string, len(m.Files))
	for i := range m.Files {
		f := &m.Files[i]
		if err := validatePayloadPath(f.Source); err != nil {
			return fmt.Errorf("files[%d].source: %w", i, err)
		}
		plugin := profile.PluginTarget(f.Target) == nil
		loader := m.SchemaVersion == 3 && profile.LoaderTarget(f.Target) == nil
		if m.SchemaVersion == 3 {
			if !loader || !f.OriginalAbsent || f.OriginalSHA256 != "" || f.PayloadSHA256 != profile.LoaderSHA256 || f.PayloadBytes != profile.LoaderBytes {
				return fmt.Errorf("files[%d]: loader target, absent origin, hash, or size differs from the pinned release", i)
			}
		}
		if !plugin && !loader {
			if err := profile.Target(f.Target); err != nil {
				return fmt.Errorf("files[%d].target: %w", i, err)
			}
		}
		if m.SchemaVersion == 1 && f.OriginalAbsent {
			return fmt.Errorf("files[%d].originalAbsent is not allowed in schemaVersion 1", i)
		}
		if m.SchemaVersion == 1 && plugin {
			return fmt.Errorf("files[%d].target: plugin targets require schemaVersion 2", i)
		}
		if m.SchemaVersion == 2 && plugin || loader {
			if !f.OriginalAbsent || f.OriginalSHA256 != "" {
				return fmt.Errorf("files[%d]: plugin target requires originalAbsent=true and empty originalSha256", i)
			}
		} else if f.OriginalAbsent || !hexPattern.MatchString(f.OriginalSHA256) {
			return fmt.Errorf("files[%d].originalSha256 must be lowercase SHA-256", i)
		}
		key := profile.Key(f.Target)
		if prior, exists := targets[key]; exists {
			return fmt.Errorf("duplicate target %q conflicts with %q", f.Target, prior)
		}
		targets[key] = f.Target
		sourceKey := strings.ToLower(f.Source)
		if prior, exists := sources[sourceKey]; exists {
			return fmt.Errorf("duplicate source %q conflicts with %q", f.Source, prior)
		}
		sources[sourceKey] = f.Source
		if f.PayloadBytes < 0 || f.PayloadBytes > maxPayloadBytes {
			return fmt.Errorf("files[%d].payloadBytes out of range", i)
		}
		if authoring {
			if f.PayloadSHA256 != "" && !isZeroHash(f.PayloadSHA256) && !hexPattern.MatchString(f.PayloadSHA256) {
				return fmt.Errorf("files[%d].payloadSha256 must be lowercase SHA-256", i)
			}
		} else {
			if !hexPattern.MatchString(f.PayloadSHA256) {
				return fmt.Errorf("files[%d].payloadSha256 must be lowercase SHA-256", i)
			}
		}
	}
	return nil
}

func pathDir(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[:i]
	}
	return "."
}

func validText(s string, max int) bool {
	if s == "" || len(s) > max {
		return false
	}
	for _, c := range []byte(s) {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

func validatePayloadPath(path string) error {
	if err := validatePackagePath(path); err != nil {
		return err
	}
	if !strings.HasPrefix(path, "payload/") || path == "payload/" {
		return errors.New("source must be beneath payload/")
	}
	return nil
}

func validatePackagePath(path string) error {
	return profile.Path(path)
}

func validateZipName(name string) error {
	isDir := strings.HasSuffix(name, "/")
	trimmed := strings.TrimSuffix(name, "/")
	if err := validatePackagePath(trimmed); err != nil {
		return err
	}
	if isDir && trimmed == "" {
		return errors.New("invalid ZIP root directory")
	}
	return nil
}

// scanJSON rejects duplicate object keys recursively before strict decoding.
func scanJSON(data []byte) error {
	dec := json.NewDecoder(bufio.NewReader(bytes.NewReader(data)))
	dec.UseNumber()
	if err := scanJSONValue(dec); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func validateJSONFieldNames(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return err
	}
	root, ok := value.(map[string]any)
	if !ok {
		return errors.New("manifest must be a JSON object")
	}
	manifestFields := map[string]struct{}{
		"schemaVersion": {}, "id": {}, "version": {}, "name": {}, "profile": {}, "files": {},
	}
	if err := checkJSONObjectFields(root, manifestFields, "manifest"); err != nil {
		return err
	}
	rawFiles, exists := root["files"]
	if !exists {
		return nil
	}
	items, ok := rawFiles.([]any)
	if !ok {
		return nil // Decoder below reports the type error with field context.
	}
	fileFields := map[string]struct{}{
		"source": {}, "target": {}, "originalSha256": {}, "originalAbsent": {}, "payloadSha256": {}, "payloadBytes": {},
	}
	for i, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue // Decoder below reports non-object file entries.
		}
		if err := checkJSONObjectFields(object, fileFields, fmt.Sprintf("manifest.files[%d]", i)); err != nil {
			return err
		}
		if schema, ok := root["schemaVersion"].(json.Number); ok && schema.String() == "1" {
			if _, present := object["originalAbsent"]; present {
				return fmt.Errorf("unknown field %q in manifest.files[%d] for schemaVersion 1", "originalAbsent", i)
			}
		}
	}
	return nil
}

func checkJSONObjectFields(object map[string]any, allowed map[string]struct{}, context string) error {
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("unknown field %q in %s", key, context)
		}
	}
	return nil
}

func scanJSONValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); ok {
		switch delim {
		case '{':
			keys := make(map[string]struct{})
			for dec.More() {
				key, err := dec.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return errors.New("object key is not string")
				}
				if _, exists := keys[name]; exists {
					return fmt.Errorf("duplicate JSON key %q", name)
				}
				keys[name] = struct{}{}
				if err := scanJSONValue(dec); err != nil {
					return err
				}
			}
			end, err := dec.Token()
			if err != nil {
				return err
			}
			if end != json.Delim('}') {
				return errors.New("unterminated JSON object")
			}
		case '[':
			for dec.More() {
				if err := scanJSONValue(dec); err != nil {
					return err
				}
			}
			end, err := dec.Token()
			if err != nil {
				return err
			}
			if end != json.Delim(']') {
				return errors.New("unterminated JSON array")
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delim)
		}
	}
	return nil
}
