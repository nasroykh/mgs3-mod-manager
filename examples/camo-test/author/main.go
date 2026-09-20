package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	input := flag.String("in", "", "input uncompressed A8R8G8B8 DDS")
	output := flag.String("out", "", "output recolored DDS")
	check := flag.Bool("check", false, "validate input without writing output")
	flag.Parse()

	if *input == "" || flag.NArg() != 0 || (*check && *output != "") || (!*check && *output == "") {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/camo-test/author -check -in INPUT.dds")
		fmt.Fprintln(os.Stderr, "   or: go run ./examples/camo-test/author -in INPUT.dds -out OUTPUT.dds")
		os.Exit(2)
	}
	inPath, err := filepath.Abs(*input)
	if err != nil {
		fail(err)
	}
	inputBytes, err := os.ReadFile(inPath)
	if err != nil {
		fail(err)
	}
	if *check {
		layout, err := parseDDS(inputBytes)
		if err != nil {
			fail(err)
		}
		fmt.Printf("valid: %s: %dx%d, %d mip levels, %d pixels\n", inPath, layout.width, layout.height, len(layout.mips), layout.pixelCount())
		return
	}

	outPath, err := filepath.Abs(*output)
	if err != nil {
		fail(err)
	}
	if strings.EqualFold(filepath.Clean(inPath), filepath.Clean(outPath)) {
		fail(fmt.Errorf("input and output paths must differ"))
	}
	outputBytes, layout, err := recolorDDS(inputBytes)
	if err != nil {
		fail(err)
	}
	if err := writeExclusive(outPath, outputBytes); err != nil {
		fail(err)
	}
	fmt.Printf("wrote %s: %dx%d, %d mip levels, %d pixels\n", outPath, layout.width, layout.height, len(layout.mips), layout.pixelCount())
}

func writeExclusive(path string, data []byte) (err error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if err == nil {
			return
		}
		if !closed {
			err = errors.Join(err, file.Close())
		}
		err = errors.Join(err, os.Remove(path))
	}()
	var written int
	if written, err = file.Write(data); err != nil {
		return err
	}
	if written != len(data) {
		return io.ErrShortWrite
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		closed = true
		return err
	}
	closed = true
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
