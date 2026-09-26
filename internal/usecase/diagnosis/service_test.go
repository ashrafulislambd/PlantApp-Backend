package diagnosis

import (
	"context"
	"errors"
	"testing"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/diagnosis"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/repository/memory"
)

type stubProvider struct {
	result diagnosis.AnalysisResult
	err    error
}

func (p stubProvider) Analyze(_ context.Context, _ []byte) (diagnosis.AnalysisResult, error) {
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
