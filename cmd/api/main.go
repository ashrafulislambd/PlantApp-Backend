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
	"myplantpal-backend/internal/domain/chat"
	"myplantpal-backend/internal/domain/diagnosis"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/ai"
	"myplantpal-backend/internal/infrastructure/ai/gemini"
	"myplantpal-backend/internal/infrastructure/ai/groq"
	"myplantpal-backend/internal/infrastructure/repository/memory"
	"myplantpal-backend/internal/infrastructure/repository/memory/seed"
	productseed "myplantpal-backend/internal/infrastructure/repository/memory/seed/products"
	mongorepo "myplantpal-backend/internal/infrastructure/repository/mongo"
	"myplantpal-backend/internal/infrastructure/security"
	httpapi "myplantpal-backend/internal/interface/http"
	"myplantpal-backend/internal/interface/http/authmw"
	v1 "myplantpal-backend/internal/interface/http/v1"
	authuc "myplantpal-backend/internal/usecase/auth"
	chatuc "myplantpal-backend/internal/usecase/chat"
	diagnosisuc "myplantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "myplantpal-backend/internal/usecase/fertilizer"
	plantuc "myplantpal-backend/internal/usecase/plant"
	productuc "myplantpal-backend/internal/usecase/product"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	ids := idgen.New()

	mongoCtx := context.Background()
	mongoClient, err := mongorepo.Connect(mongoCtx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	db := mongoClient.Database(cfg.MongoDBName)

	userRepo := mongorepo.NewUserRepository(db)
	refreshRepo := mongorepo.NewRefreshTokenRepository(db)
	if err := userRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("user indexes: %v", err)
	}
	if err := refreshRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("refresh token indexes: %v", err)
	}

	jwtIssuer := security.NewJWTIssuer(cfg.JWTSecret, "myplantpal-backend", cfg.JWTAccessTTL)
	authService := authuc.NewService(userRepo, refreshRepo, ids, jwtIssuer, cfg.JWTRefreshTTL)

	// In-memory repositories for now; swap each for a MongoDB-backed
	// implementation later without touching usecases or handlers.
	plantRepo := memory.NewPlantRepository()
	fertilizerRepo := memory.NewFertilizerRepository(seed.Fertilizers()...)
	diagnosisRepo := memory.NewDiagnosisRepository()
	chatRepo := memory.NewChatRepository()
	productRepo := memory.NewProductRepository(productseed.Products(), productseed.Categories())

	// Groq Compound price refresher — only enabled when GROQ_API_KEY is set.
	var priceRefresher productuc.PriceRefresherPort
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		priceRefresher = groq.New(key)
		log.Println("Groq price refresh: enabled")
	} else {
		log.Println("Groq price refresh: disabled (set GROQ_API_KEY to enable)")
	}

	// AI Chat Box / AI Doctor providers: Gemini (primary), Groq (fallback),
	// mock (last resort). Only providers whose API key is configured join
	// the chain; mock is always last so the chain can never fail outright.
	var chatEntries []ai.ChatProviderEntry
	var diagnosisEntries []ai.DiagnosisProviderEntry
	if cfg.GeminiAPIKey != "" {
		geminiClient := gemini.New(cfg.GeminiAPIKey, cfg.GeminiModel)
		chatEntries = append(chatEntries, ai.ChatProviderEntry{Name: "gemini", Provider: geminiClient})
		diagnosisEntries = append(diagnosisEntries, ai.DiagnosisProviderEntry{Name: "gemini", Provider: geminiClient})
	}
	if cfg.GroqAPIKey != "" {
		chatEntries = append(chatEntries, ai.ChatProviderEntry{
			Name:     "groq",
			Provider: groq.NewChatProvider(cfg.GroqAPIKey, cfg.GroqChatModel),
		})
		diagnosisEntries = append(diagnosisEntries, ai.DiagnosisProviderEntry{
			Name:     "groq",
			Provider: groq.NewDiagnosisProvider(cfg.GroqAPIKey, cfg.GroqVisionModel),
		})
	}
	chatEntries = append(chatEntries, ai.ChatProviderEntry{Name: "mock", Provider: ai.NewMockChatReplyProvider()})
	diagnosisEntries = append(diagnosisEntries, ai.DiagnosisProviderEntry{Name: "mock", Provider: ai.NewMockDiagnosisProvider()})

	log.Printf("AI providers: gemini=%v groq=%v (mock always available as last resort)",
		cfg.GeminiAPIKey != "", cfg.GroqAPIKey != "")

	var chatReplyProvider chat.ReplyProvider = ai.NewFallbackChatProvider(chatEntries...)
	var diagnosisProvider diagnosis.Provider = ai.NewFallbackDiagnosisProvider(diagnosisEntries...)

	deps := v1.Dependencies{
		PlantService:      plantuc.NewService(plantRepo, ids),
		FertilizerService: fertilizeruc.NewService(fertilizerRepo, ids),
		DiagnosisService:  diagnosisuc.NewService(diagnosisRepo, diagnosisProvider, ids),
		ChatService:       chatuc.NewService(chatRepo, chatReplyProvider, ids),
		ProductService:    productuc.NewService(productRepo, priceRefresher),
		AuthService:       authService,
		RequireAuth:       authmw.RequireAuth(jwtIssuer),
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
