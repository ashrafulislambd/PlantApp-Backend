package diagnosis

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/idgen"
)

// ChatRecorder saves a scan as two turns of a chat session. It is
// implemented by the chat usecase; declared here so this package does not
// depend on it.
type ChatRecorder interface {
	RecordScan(ctx context.Context, userID, sessionID, note string, d *diagnosis.Diagnosis, lang string) error
}

type Service struct {
	repo     diagnosis.Repository
	provider diagnosis.Provider
	ids      idgen.Generator
	images   diagnosis.ImageStore // optional; nil means photos are not kept
	chat     ChatRecorder         // optional; nil means scans never reach chat history
}

func NewService(repo diagnosis.Repository, provider diagnosis.Provider, ids idgen.Generator) *Service {
	return &Service{repo: repo, provider: provider, ids: ids}
}

// SetImageStore turns on saving the submitted photo alongside the Diagnosis.
func (s *Service) SetImageStore(store diagnosis.ImageStore) { s.images = store }

// SetChatRecorder lets a scan sent from the chat (AnalyzeInput.SessionID) be
// recorded in that chat session.
func (s *Service) SetChatRecorder(r ChatRecorder) { s.chat = r }

type AnalyzeInput struct {
	UserID      string
	PlantID     *string
	ImageData   []byte
	ContentType string // image/jpeg, image/png or image/webp

	// Note is the user's optional caption. It is passed to the vision model
	// as a hint and, with SessionID, saved as the user's chat turn.
	Note string
	// SessionID, when set, also saves this scan as two turns (the user's
	// photo + note, and a short assistant summary) in that chat session.
	SessionID string
	// Lang ("en"/"bn") picks the language of the saved assistant summary.
	Lang string
}

// maxSessionIDLen guards the chat session id taken from a request.
const maxSessionIDLen = 128

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

	note := strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(note) > diagnosis.MaxNoteRunes {
		note = string([]rune(note)[:diagnosis.MaxNoteRunes])
	}
	sessionID := strings.TrimSpace(in.SessionID)
	if len(sessionID) > maxSessionIDLen {
		return nil, fmt.Errorf("%w: sessionId is too long", apperr.ErrInvalidInput)
	}

	result, err := s.provider.Analyze(ctx, in.ImageData, note)
	if err != nil {
		return nil, err
	}

	d := &diagnosis.Diagnosis{
		ID:           s.ids.New("diag"),
		UserID:       in.UserID,
		PlantID:      in.PlantID,
		Issue:        result.Issue,
		Cure:         result.Cure,
		Confidence:   result.Confidence,
		Severity:     result.Severity,
		Fertilizer:   result.Fertilizer,
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

	// The photo and the Diagnosis are saved exactly once, above. Recording
	// the chat turns only adds references to them (DiagnosisID). A failure
	// here must not discard a result the user already waited for, so it is
	// logged and the diagnosis is still returned.
	if sessionID != "" && s.chat != nil {
		if err := s.chat.RecordScan(ctx, in.UserID, sessionID, note, d, in.Lang); err != nil {
			log.Printf("diagnosis: recording scan %s in chat session failed: %v", d.ID, err)
		}
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
