// Package localfs stores uploaded images as plain files under one directory.
package localfs

import (
"context"
"errors"
"fmt"
"io"
"io/fs"
"os"
"path/filepath"
"strings"

"plantpal-backend/internal/domain/apperr"
)

type Store struct {
root string
}

func New(root string) (*Store, error) {
if err := os.MkdirAll(root, 0o750); err != nil {
return nil, fmt.Errorf("create upload dir: %w", err)
}
return &Store{root: root}, nil
}

// path rejects anything that isn't a bare file name, so a key can never
// escape the upload directory.
func (s *Store) path(key string) (string, error) {
if key == "" || key != filepath.Base(key) || strings.ContainsAny(key, `/\`) || strings.HasPrefix(key, ".") {
return "", fmt.Errorf("%w: invalid image key", apperr.ErrInvalidInput)
}
return filepath.Join(s.root, key), nil
}

func (s *Store) Save(_ context.Context, key string, data []byte) error {
p, err := s.path(key)
if err != nil {
return err
}
return os.WriteFile(p, data, 0o640)
}

func (s *Store) Open(_ context.Context, key string) (io.ReadCloser, error) {
p, err := s.path(key)
if err != nil {
return nil, err
}
f, err := os.Open(p)
if errors.Is(err, fs.ErrNotExist) {
return nil, apperr.ErrNotFound
}
if err != nil {
return nil, err
}
return f, nil
}

func (s *Store) Delete(_ context.Context, key string) error {
p, err := s.path(key)
if err != nil {
return err
}
if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
return err
}
return nil
}