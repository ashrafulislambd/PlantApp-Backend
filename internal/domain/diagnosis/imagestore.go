package diagnosis

import (
"context"
"io"
)

// ImageStore keeps the photo behind a Diagnosis. The local-disk
// implementation lives in internal/infrastructure/storage/localfs; swap in
// S3/GCS later without touching the usecase or delivery layers.
type ImageStore interface {
Save(ctx context.Context, key string, data []byte) error
// Open returns apperr.ErrNotFound when the key doesn't exist.
Open(ctx context.Context, key string) (io.ReadCloser, error)
Delete(ctx context.Context, key string) error
}