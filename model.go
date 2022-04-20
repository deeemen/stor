package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type (
	FileSystem interface {
		http.FileSystem
		OpenHash(*MD5Hash) (http.File, error)
		Store(context.Context, io.Reader, *FileMetadata) (*FileMetadata, error)
		Delete(*MD5Hash) error
	}
	MD5Hash      [md5.Size]byte
	FileMetadata struct {
		MD5    *MD5Hash
		Length int64
	}
)

var (
	ErrMD5Mismatch = errors.New("md5 mismatch")
)

func (h *MD5Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}
func (h *MD5Hash) UnmarshalJSON(b []byte) error {
	s := ""
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	return h.FromString(s)
}
func (h *MD5Hash) String() string {
	return hex.EncodeToString(h[:])
}
func (h *MD5Hash) FromString(s string) error {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(decoded) != md5.Size {
		return errors.New("invalid md5")
	}
	copy(h[:], decoded)
	return nil
}
func (h *MD5Hash) Equals(a *MD5Hash) bool {
	return bytes.Equal(h[:], a[:])
}
