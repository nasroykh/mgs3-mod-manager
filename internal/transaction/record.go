// Package transaction implements immutable, checksummed journal records.
package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const MaxRecord = 8 << 20

type envelope struct {
	Schema int             `json:"schema"`
	SHA256 string          `json:"sha256"`
	Data   json.RawMessage `json:"data"`
}

func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// Decode accepts only the exact encoding emitted by this program. Besides
// rejecting unknown fields, this rejects duplicate keys and ambiguous encodings.
func Decode[T any](b []byte, dst *T) error {
	if len(b) > MaxRecord {
		return fmt.Errorf("record exceeds size limit")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	normalized, err := json.Marshal(dst)
	if err != nil {
		return err
	}
	if !bytes.Equal(normalized, b) {
		return fmt.Errorf("noncanonical or duplicate JSON fields")
	}
	return nil
}

func Read[T any](root *os.Root, path string, dst *T) error {
	f, err := root.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, MaxRecord+1))
	if err != nil {
		return err
	}
	var e envelope
	if err = Decode(b, &e); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if e.Schema != 1 || e.SHA256 != Hash(e.Data) {
		return fmt.Errorf("%s: invalid record checksum/schema", path)
	}
	return Decode(e.Data, dst)
}

func Write(root *os.Root, path string, value any) error {
	return WriteUsing(root, path, value, nil)
}

// SyncWriter is the small durable-write boundary. Production uses *os.File;
// synthetic tests wrap it to inject partial writes and failed synchronization.
type SyncWriter interface {
	io.Writer
	Sync() error
	Close() error
}
type FileFactory func(root *os.Root, path string) (SyncWriter, error)

func Create(root *os.Root, path string, factory FileFactory) (SyncWriter, error) {
	if factory != nil {
		return factory(root, path)
	}
	return root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
}

func WriteUsing(root *os.Root, path string, value any, factory FileFactory) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	e, err := json.Marshal(envelope{1, Hash(b), b})
	if err != nil {
		return err
	}
	if len(e) > MaxRecord {
		return fmt.Errorf("record exceeds size limit")
	}
	return WriteBytesUsing(root, path, e, factory)
}

// Exclusive creation prevents replacing surviving evidence after an error.
func WriteBytes(root *os.Root, path string, b []byte) error {
	return WriteBytesUsing(root, path, b, nil)
}

func WriteBytesUsing(root *os.Root, path string, b []byte, factory FileFactory) error {
	f, err := Create(root, path, factory)
	if err != nil {
		return err
	}
	n, writeErr := f.Write(b)
	if writeErr == nil && n != len(b) {
		writeErr = io.ErrShortWrite
	}
	if writeErr == nil {
		writeErr = f.Sync()
	}
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
