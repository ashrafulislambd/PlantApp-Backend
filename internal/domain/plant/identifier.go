package plant

import "context"

// Identifier identifies a plant from an image.
type Identifier interface {
	Identify(ctx context.Context, imageData []byte) (IdentificationResult, error)
}
