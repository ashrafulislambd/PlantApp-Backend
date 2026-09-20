// Command api runs the MyPlantPal HTTP API.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myplantpal-backend/internal/config"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/ai"
	"myplantpal-backend/internal/infrastructure/repository/memory"
	"myplantpal-backend/internal/infrastructure/repository/memory/seed"
	httpapi "myplantpal-backend/internal/interface/http"
	v1 "myplantpal-backend/internal/interface/http/v1"
	chatuc "myplantpal-backend/internal/usecase/chat"
	diagnosisuc "myplantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "myplantpal-backend/internal/usecase/fertilizer"
	plantuc "myplantpal-backend/internal/usecase/plant"
)

func main() {
	cfg := config.Load()
	ids := idgen.New()

	// In-memory repositories for now; swap each for a MongoDB-backed
	// implementation later without touching usecases or handlers.
	plantRepo := memory.NewPlantRepository()
	fertilizerRepo := memory.NewFertilizerRepository(seed.Fertilizers()...)
	diagnosisRepo := memory.NewDiagnosisRepository()
	chatRepo := memory.NewChatRepository()

	deps := v1.Dependencies{
		PlantService:      plantuc.NewService(plantRepo, ids),
		FertilizerService: fertilizeruc.NewService(fertilizerRepo, ids),
		DiagnosisService:  diagnosisuc.NewService(diagnosisRepo, ai.NewMockDiagnosisProvider(), ids),
		ChatService:       chatuc.NewService(chatRepo, ai.NewMockChatReplyProvider(), ids),
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("myplantpal backend listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
}
