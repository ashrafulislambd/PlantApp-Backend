package diagnosis

import (
	"context"
	"errors"
	"testing"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/repository/memory"
)

type stubProvider struct {
	result diagnosis.AnalysisResult
	err    error
}

func (p stubProvider) Analyze(_ context.Context, _ []byte, _ string) (diagnosis.AnalysisResult, error) {
	return p.result, p.err
}

func newTestService(provider diagnosis.Provider) *Service {
	return NewService(memory.NewDiagnosisRepository(), provider, idgen.New())
}

func TestAnalyze_RequiresImageData(t *testing.T) {
	svc := newTestService(stubProvider{})
	_, err := svc.Analyze(context.Background(), AnalyzeInput{UserID: "user_1", ImageData: nil})
	if !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("Analyze() error = %v, want %v", err, apperr.ErrInvalidInput)
	}
}

func TestAnalyze_PersistsResultWithDisclaimerAndTranslations(t *testing.T) {
	provider := stubProvider{result: diagnosis.AnalysisResult{
		Issue:   "Phosphorus deficiency.",
		Cure:    "Use TSP.",
		IssueBn: "ফসফরাসের অভাব।",
		CureBn:  "TSP ব্যবহার করুন।",
	}}
	svc := newTestService(provider)
	plantID := "pl_1"
	userID := "user_1"

	d, err := svc.Analyze(context.Background(), AnalyzeInput{UserID: userID, PlantID: &plantID, ImageData: []byte{1, 2, 3}})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if d.Issue != provider.result.Issue || d.Cure != provider.result.Cure {
		t.Errorf("Analyze() = %+v, want issue/cure from provider", d)
	}
	if d.Disclaimer != diagnosis.Disclaimer {
		t.Errorf("Disclaimer = %q, want the standard disclaimer", d.Disclaimer)
	}
	if d.IssueBn != provider.result.IssueBn || d.CureBn != provider.result.CureBn || d.DisclaimerBn != diagnosis.DisclaimerBn {
		t.Errorf("Bengali translations not propagated onto Diagnosis: %+v", d)
	}
	if d.PlantID == nil || *d.PlantID != plantID {
		t.Errorf("PlantID = %v, want %q", d.PlantID, plantID)
	}

	fetched, err := svc.Get(context.Background(), d.ID, userID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if fetched.ID != d.ID {
		t.Error("Get() returned a different diagnosis than was created")
	}
}

func TestAnalyze_PropagatesProviderError(t *testing.T) {
	wantErr := errors.New("provider exploded")
	svc := newTestService(stubProvider{err: wantErr})
	_, err := svc.Analyze(context.Background(), AnalyzeInput{UserID: "user_1", ImageData: []byte{1}})
	if !errors.Is(err, wantErr) {
		t.Errorf("Analyze() error = %v, want %v", err, wantErr)
	}
}

func TestList_FiltersByPlantID(t *testing.T) {
	svc := newTestService(stubProvider{result: diagnosis.AnalysisResult{Issue: "x", Cure: "y"}})
	ctx := context.Background()
	plantA := "pl_a"
	plantB := "pl_b"
	userID := "user_1"

	if _, err := svc.Analyze(ctx, AnalyzeInput{UserID: userID, PlantID: &plantA, ImageData: []byte{1}}); err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if _, err := svc.Analyze(ctx, AnalyzeInput{UserID: userID, PlantID: &plantB, ImageData: []byte{1}}); err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	list, err := svc.List(ctx, userID, &plantA)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List(plantA) returned %d diagnoses, want 1", len(list))
	}
}

// --- chat session + note -------------------------------------------------

type noteStub struct{ gotNote string }

func (p *noteStub) Analyze(_ context.Context, _ []byte, note string) (diagnosis.AnalysisResult, error) {
	p.gotNote = note
	return diagnosis.AnalysisResult{Issue: "Leaf spot", Cure: "Prune"}, nil
}

type recorderStub struct {
	calls     int
	sessionID string
	note      string
	diagID    string
	err       error
}

func (r *recorderStub) RecordScan(_ context.Context, _, sessionID, note string, d *diagnosis.Diagnosis, _ string) error {
	r.calls++
	r.sessionID, r.note, r.diagID = sessionID, note, d.ID
	return r.err
}

func TestAnalyze_PassesNoteToProviderAndRecordsInChatOnce(t *testing.T) {
	p := &noteStub{}
	rec := &recorderStub{}
	repo := memory.NewDiagnosisRepository()
	svc := NewService(repo, p, idgen.New())
	svc.SetChatRecorder(rec)

	d, err := svc.Analyze(context.Background(), AnalyzeInput{
		UserID: "u1", ImageData: []byte{1}, SessionID: " s1 ", Note: "  white spots  ", Lang: "en",
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if p.gotNote != "white spots" {
		t.Errorf("provider note = %q, want trimmed note", p.gotNote)
	}
	if rec.calls != 1 || rec.sessionID != "s1" || rec.note != "white spots" || rec.diagID != d.ID {
		t.Errorf("recorder = %+v, want one call for this diagnosis", rec)
	}
	if all, _ := repo.List(context.Background(), "u1", nil); len(all) != 1 {
		t.Errorf("%d diagnoses stored, want exactly 1", len(all))
	}
}

func TestAnalyze_NoSessionMeansNoChatRecording(t *testing.T) {
	rec := &recorderStub{}
	svc := NewService(memory.NewDiagnosisRepository(), &noteStub{}, idgen.New())
	svc.SetChatRecorder(rec)
	if _, err := svc.Analyze(context.Background(), AnalyzeInput{UserID: "u1", ImageData: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	if rec.calls != 0 {
		t.Errorf("recorder called %d times without a sessionId", rec.calls)
	}
}

func TestAnalyze_ChatRecordingFailureDoesNotLoseTheDiagnosis(t *testing.T) {
	rec := &recorderStub{err: errors.New("db down")}
	svc := NewService(memory.NewDiagnosisRepository(), &noteStub{}, idgen.New())
	svc.SetChatRecorder(rec)
	d, err := svc.Analyze(context.Background(), AnalyzeInput{UserID: "u1", ImageData: []byte{1}, SessionID: "s1"})
	if err != nil || d == nil {
		t.Fatalf("Analyze() = %v, %v; want the diagnosis despite the recorder failing", d, err)
	}
}
