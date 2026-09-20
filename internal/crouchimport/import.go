// Package crouchimport converts the pinned offline crouch-walk archive.
package crouchimport

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/winfs"
	"os"
	"path/filepath"
	"strings"
)

const (
	UpstreamSHA256 = "81c364f1253f33cda9244c89b7d8f9bc499fe780e4c00149d2ec505eca27a757"
	UpstreamBytes  = int64(22061918)
	maxReadBytes   = UpstreamBytes
)

// Convert validates upstream archive and creates normalized package. dryRun performs all validation without output/staging writes.
func Convert(input, output string, dryRun bool) (*packagefmt.Package, error) {
	data, err := readUpstream(input)
	if err != nil {
		return nil, err
	}
	if err := packagefmt.CheckOutput(output); err != nil {
		return nil, err
	}
	pkg, err := normalize(data)
	if err != nil {
		return nil, err
	}
	manifestBytes, err := json.Marshal(pkg.Manifest)
	if err != nil {
		return nil, fmt.Errorf("encode normalized manifest: %w", err)
	}
	if _, err := packagefmt.Parse(manifestBytes, false); err != nil {
		return nil, fmt.Errorf("validate normalized manifest: %w", err)
	}
	if dryRun {
		return pkg, nil
	}
	parent := filepath.Dir(output)
	stage, err := os.MkdirTemp(parent, ".crouchimport-")
	if err != nil {
		return nil, fmt.Errorf("create staging directory: %w", err)
	}
	defer os.RemoveAll(stage)
	if err := writeFolder(stage, pkg); err != nil {
		return nil, err
	}
	return packagefmt.Pack(stage, output)
}

func readUpstream(path string) ([]byte, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("upstream path: %w", err)
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, fmt.Errorf("upstream archive: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("upstream archive must be a regular file")
	}
	if err := winfs.CheckPath(filepath.Dir(abs), filepath.Base(abs), false); err != nil {
		return nil, fmt.Errorf("upstream archive path: %w", err)
	}
	if info.Size() != UpstreamBytes {
		return nil, fmt.Errorf("upstream archive size %d does not match pinned size %d", info.Size(), UpstreamBytes)
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open upstream archive: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxReadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read upstream archive: %w", err)
	}
	if int64(len(data)) != UpstreamBytes {
		return nil, errors.New("upstream archive truncated")
	}
	h := sha256.Sum256(data)
	if hex.EncodeToString(h[:]) != UpstreamSHA256 {
		return nil, errors.New("upstream archive SHA-256 does not match pinned release")
	}
	return data, nil
}

func normalize(data []byte) (*packagefmt.Package, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read upstream ZIP: %w", err)
	}
	files := profile.CrouchFiles()
	if len(files) != 119 {
		return nil, fmt.Errorf("pinned crouch metadata has %d files, want 119", len(files))
	}
	byTarget := make(map[string]profile.CrouchFile, len(files))
	for _, f := range files {
		byTarget[f.Target] = f
	}
	payloads := make(map[string][]byte, len(files))
	seen := make(map[string]bool, len(files))
	for _, e := range zr.File {
		name := strings.TrimSuffix(e.Name, "/")
		if e.FileInfo().IsDir() {
			continue
		}
		if strings.EqualFold(name, "d3d11.dll") {
			continue
		}
		target := name
		if strings.EqualFold(filepath.Ext(name), ".asi") || strings.EqualFold(filepath.Ext(name), ".ini") {
			target = strings.ToLower(filepath.Base(name))
		}
		f, ok := byTarget[target]
		if !ok {
			return nil, fmt.Errorf("upstream entry %q is not pinned", e.Name)
		}
		if seen[target] {
			return nil, fmt.Errorf("duplicate upstream target %q", target)
		}
		if e.UncompressedSize64 != uint64(f.PayloadBytes) {
			return nil, fmt.Errorf("payload %q size mismatch", target)
		}
		r, err := e.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(io.LimitReader(r, int64(f.PayloadBytes)+1))
		r.Close()
		if err != nil {
			return nil, err
		}
		if int64(len(b)) != f.PayloadBytes {
			return nil, fmt.Errorf("payload %q truncated", target)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != f.PayloadSHA256 {
			return nil, fmt.Errorf("payload %q SHA-256 mismatch", target)
		}
		source := "payload/" + target
		payloads[source] = b
		seen[target] = true
	}
	if len(seen) != len(files) {
		return nil, fmt.Errorf("upstream payload set has %d pinned files, want %d", len(seen), len(files))
	}
	manifestFiles := make([]packagefmt.File, 0, len(files))
	for _, f := range files {
		manifestFiles = append(manifestFiles, packagefmt.File{Source: "payload/" + f.Target, Target: f.Target, OriginalSHA256: f.OriginalSHA256, OriginalAbsent: f.OriginalAbsent, PayloadSHA256: f.PayloadSHA256, PayloadBytes: f.PayloadBytes})
	}
	m := packagefmt.Manifest{SchemaVersion: 4, ID: profile.CrouchID, Version: profile.CrouchVersion, Name: "MGS3 Crouch Walk", Profile: profile.ID, Files: manifestFiles}
	return &packagefmt.Package{Manifest: m, Payloads: payloads}, nil
}

func writeFolder(root string, p *packagefmt.Package) error {
	b, err := json.Marshal(p.Manifest)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), b, 0600); err != nil {
		return err
	}
	for name, data := range p.Payloads {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			return err
		}
	}
	return nil
}
