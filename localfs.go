package main

import (
	"context"
	"crypto/md5"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type (
	LocalFS struct {
		path string
	}
)

func NewLocalFS(path string) (*LocalFS, error) {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to mkdir '%s'", path)
	}
	return &LocalFS{
		path: path,
	}, nil
}
func (fsys *LocalFS) createTempFile() (*os.File, error) {
	return os.CreateTemp(fsys.path, "store-*")
}
func (fsys *LocalFS) exists(hash *MD5Hash) bool {
	path := fsys.filePath(hash)
	_, err := os.Stat(path)

	return err == nil
}
func (fsys *LocalFS) OpenHash(h *MD5Hash) (http.File, error) {
	fileName := h.String()
	filePath := path.Join(fsys.path, fileName[:2], fileName)
	f, err := os.Open(filePath)
	return f, err
}

func (fsys *LocalFS) Open(name string) (http.File, error) {
	h := &MD5Hash{}
	err := h.FromString(name)
	if err != nil {
		return nil, err
	}
	return fsys.OpenHash(h)

}
func (fsys *LocalFS) filePath(hash *MD5Hash) string {
	fn := strings.ToLower(hash.String())
	return path.Join(fsys.path, fn[:2], fn)
}
func (fsys *LocalFS) Delete(h *MD5Hash) error {
	filePath := fsys.filePath(h)
	return os.Remove(filePath)
}

func (fsys *LocalFS) Store(ctx context.Context, r io.Reader, m *FileMetadata) (*FileMetadata, error) {
	tempf, err := fsys.createTempFile()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file")
	}
	defer tempf.Close()
	defer os.Remove(tempf.Name())

	hashW := md5.New()
	w := io.MultiWriter(tempf, hashW)
	written, err := io.Copy(w, r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write to disk")
	}
	if err := tempf.Sync(); err != nil {
		return nil, errors.Wrap(err, "failed to sync temp file")
	}

	md5Hash := (*MD5Hash)(hashW.Sum(nil))

	if m.MD5 != nil {
		if !m.MD5.Equals(md5Hash) {
			return nil, ErrMD5Mismatch
		}
	}

	filePath := fsys.filePath(md5Hash)
	if exists := fsys.exists(md5Hash); exists {
		return nil, fs.ErrExist
	}
	fileDir := path.Dir(filePath)

	if err := os.MkdirAll(fileDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "failed to create file dir '%s'", fileDir)
	}
	if err := os.Link(tempf.Name(), filePath); err != nil {
		return nil, errors.Wrap(err, "failed to hardlink tempfile to new location")
	}

	return &FileMetadata{
		MD5:    md5Hash,
		Length: written,
	}, nil

}
