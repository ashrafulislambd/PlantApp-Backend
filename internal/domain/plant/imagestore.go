package plant

import (
	"context"
	"io"
)

// ImageStore persists plant photos.
type ImageStore interface {
	Save(ctx context.Context, key string, data []byte) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
