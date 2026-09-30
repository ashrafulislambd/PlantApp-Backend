package diagnosis

import (
"context"
"fmt"
"io"
"log"
"time"

"plantpal-backend/internal/domain/apperr"
"plantpal-backend/internal/domain/diagnosis"
"plantpal-backend/internal/idgen"
)

type Service struct {
repo     diagnosis.Repository
provider diagnosis.Provider
ids      idgen.Generator
images   diagnosis.ImageStore // optional; nil means photos are not kept
}

func NewService(repo diagnosis.Repository, provider diagnosis.Provider, ids idgen.Generator) *Service {
return &Service{repo: repo, provider: provider, ids: ids}
}

// SetImageStore turns on saving the submitted photo alongside the Diagnosis.
func (s *Service) SetImageStore(store diagnosis.ImageStore) { s.images = store }

type AnalyzeInput struct {
UserID      string
PlantID     *string
ImageData   []byte
ContentType string // image/jpeg, image/png or image/webp
}

var imageExt = map[string]string{
"image/jpeg": ".jpg",
"image/png":  ".png",
"image/webp": ".webp",
}

func (s *Service) Analyze(ctx context.Context, in AnalyzeInput) (*diagnosis.Diagnosis, error) {
if len(in.ImageData) == 0 {
return nil, fmt.Errorf("%w: imageBase64 is required", apperr.ErrInvalidInput)
}
if in.UserID == "" {
return nil, fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
}

result, err := s.provider.Analyze(ctx, in.ImageData)
if err != nil {
return nil, err
}

d := &diagnosis.Diagnosis{
ID:           s.ids.New("diag"),
UserID:       in.UserID,
PlantID:      in.PlantID,
Issue:        result.Issue,
Cure:         result.Cure,
Disclaimer:   diagnosis.Disclaimer,
IssueBn:      result.IssueBn,
CureBn:       result.CureBn,
DisclaimerBn: diagnosis.DisclaimerBn,
CreatedAt:    time.Now().UTC(),
Provider:     string(result.Provider),
}

if s.images != nil {
ct := in.ContentType
if _, ok := imageExt[ct]; !ok {
ct = "image/jpeg"
}
key := d.ID + imageExt[ct]
if err := s.images.Save(ctx, key, in.ImageData); err != nil {
return nil, fmt.Errorf("save diagnosis image: %w", err)
}
d.ImageKey, d.ImageContentType = key, ct
}

if err := s.repo.Create(ctx, d); err != nil {
if d.ImageKey != "" {
if delErr := s.images.Delete(ctx, d.ImageKey); delErr != nil {
log.Printf("diagnosis: cleanup of orphaned image %s failed: %v", d.ImageKey, delErr)
}
}
return nil, err
}
return d, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*diagnosis.Diagnosis, error) {
return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) List(ctx context.Context, userID string, plantID *string) ([]*diagnosis.Diagnosis, error) {
return s.repo.List(ctx, userID, plantID)
}

// Image opens the stored photo for one of userID's diagnoses and returns it
// with its content type. The caller must Close the reader.
func (s *Service) Image(ctx context.Context, id, userID string) (io.ReadCloser, string, error) {
d, err := s.repo.GetByID(ctx, id, userID)
if err != nil {
return nil, "", err
}
if s.images == nil || d.ImageKey == "" {
return nil, "", apperr.ErrNotFound
}
rc, err := s.images.Open(ctx, d.ImageKey)
if err != nil {
return nil, "", err
}
ct := d.ImageContentType
if ct == "" {
ct = "image/jpeg"
}
return rc, ct, nil
}