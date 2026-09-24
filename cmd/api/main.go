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
	chatdomain "myplantpal-backend/internal/domain/chat"
	diagnosisdomain "myplantpal-backend/internal/domain/diagnosis"
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

	var diagnosisProvider diagnosisdomain.Provider
	var chatProvider chatdomain.ReplyProvider
	if cfg.GroqAPIKey != "" {
		diagnosisProvider = ai.NewGroqDiagnosisProvider(cfg.GroqAPIKey, cfg.GroqVisionModel)
		chatProvider = ai.NewGroqChatReplyProvider(cfg.GroqAPIKey, cfg.GroqChatModel)
		log.Println("AI Doctor / AI Chat: using Groq")
	} else {
		diagnosisProvider = ai.NewMockDiagnosisProvider()
		chatProvider = ai.NewMockChatReplyProvider()
		log.Println("AI Doctor / AI Chat: GROQ_API_KEY not set, using mock provider")
	}

	deps := v1.Dependencies{
		PlantService:      plantuc.NewService(plantRepo, ids),
		FertilizerService: fertilizeruc.NewService(fertilizerRepo, ids),
		DiagnosisService:  diagnosisuc.NewService(diagnosisRepo, diagnosisProvider, ids),
		ChatService:       chatuc.NewService(chatRepo, chatProvider, ids),
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
