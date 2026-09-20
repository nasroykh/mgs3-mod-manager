package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func syntheticDDS() []byte {
	// 8x2 full chain has independent dimensions 8x2, 4x1, 2x1, 1x1.
	// Expected byte offsets are 128, 192, 208, and 216; total size is 220.
	dds := make([]byte, 220)
	copy(dds[:4], "DDS ")
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(dds[offset:offset+4], value) }
	put(4, 0x7c)
	put(8, 0x2100f)
	put(12, 2)
	put(16, 8)
	put(20, 32)
	put(28, 4)
	put(76, 0x20)
	put(80, 0x41)
	put(88, 32)
	put(92, 0x00ff0000)
	put(96, 0x0000ff00)
	put(100, 0x000000ff)
	put(104, 0xff000000)
	put(108, 0x401008)
	for offset := 128; offset < len(dds); offset += 4 {
		copy(dds[offset:offset+4], []byte{0x00, 0x96, 0xff, 0x80})
	}
	return dds
}

func TestParseDDSValidatesNonSquareMipOffsets(t *testing.T) {
	layout, err := parseDDS(syntheticDDS())
	if err != nil {
		t.Fatal(err)
	}
	want := []mipLayout{
		{offset: 128, width: 8, height: 2, size: 64},
		{offset: 192, width: 4, height: 1, size: 16},
		{offset: 208, width: 2, height: 1, size: 8},
		{offset: 216, width: 1, height: 1, size: 4},
	}
	if len(layout.mips) != len(want) {
		t.Fatalf("mip count = %d, want %d", len(layout.mips), len(want))
	}
	for i := range want {
		if layout.mips[i] != want[i] {
			t.Errorf("mip %d = %+v, want %+v", i, layout.mips[i], want[i])
		}
	}
}

func TestParseDDSRejectsEachWrongMask(t *testing.T) {
	tests := []struct {
		name   string
		offset int
	}{
		{name: "red", offset: 92},
		{name: "green", offset: 96},
		{name: "blue", offset: 100},
		{name: "alpha", offset: 104},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dds := syntheticDDS()
			binary.LittleEndian.PutUint32(dds[test.offset:test.offset+4], 0)
			_, err := parseDDS(dds)
			if err == nil || !strings.Contains(err.Error(), "DDS masks are") {
				t.Fatalf("error = %v, want mask rejection", err)
			}
		})
	}
}

func TestRecolorPreservesHeaderMipBoundariesAndAlpha(t *testing.T) {
	input := syntheticDDS()
	output, layout, err := recolorDDS(input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output[:128], input[:128]) {
		t.Fatal("DDS header changed")
	}
	if len(output) != len(input) {
		t.Fatalf("output length = %d, want %d", len(output), len(input))
	}
	for _, mip := range layout.mips {
		if mip.offset+mip.size > len(output) {
			t.Fatalf("mip exceeds output: %+v", mip)
		}
		for offset := mip.offset; offset < mip.offset+mip.size; offset += 4 {
			// Input BGRA 00 96 FF 80 becomes 7F 32 FF 80.
			want := []byte{0x7f, 0x32, 0xff, 0x80}
			if !bytes.Equal(output[offset:offset+4], want) {
				t.Fatalf("pixel at offset %d = % x, want % x", offset, output[offset:offset+4], want)
			}
			if output[offset+3] != input[offset+3] {
				t.Fatalf("alpha at offset %d changed", offset)
			}
		}
	}
}

func TestParseDDSRejectsTrailingBytes(t *testing.T) {
	dds := append(syntheticDDS(), 0)
	_, err := parseDDS(dds)
	if err == nil || !strings.Contains(err.Error(), "unexplained trailing bytes") {
		t.Fatalf("error = %v, want trailing-byte rejection", err)
	}
}

func TestParseDDSRejectsWrongPitch(t *testing.T) {
	dds := syntheticDDS()
	binary.LittleEndian.PutUint32(dds[20:24], 0)
	_, err := parseDDS(dds)
	if err == nil || !strings.Contains(err.Error(), "DDS pitch is") {
		t.Fatalf("error = %v, want pitch rejection", err)
	}
}

func TestWriteExclusiveRefusesExistingOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.dds")
	if err := os.WriteFile(path, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(path, []byte("replace")); err == nil {
		t.Fatal("writeExclusive accepted an existing output")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep" {
		t.Fatalf("existing output = %q, want %q", got, "keep")
	}
}

func TestWriteExclusiveSyncsAndClosesNewOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.dds")
	if err := writeExclusive(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("output = %q, want %q", got, "new")
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("output remained open: %v", err)
	}
}

func TestWriteExclusiveRefusesExistingSameFileAlias(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.dds")
	alias := filepath.Join(dir, "alias.dds")
	if err := os.WriteFile(input, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(input, alias); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(alias, []byte("replace")); err == nil {
		t.Fatal("writeExclusive accepted an existing same-file alias")
	}
	got, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep" {
		t.Fatalf("input through alias = %q, want %q", got, "keep")
	}
}
