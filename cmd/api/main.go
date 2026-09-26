package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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
	plantRepo := mongorepo.NewPlantRepository(db)
	diagnosisRepo := mongorepo.NewDiagnosisRepository(db)
	chatRepo := mongorepo.NewChatRepository(db)
	if err := userRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("user indexes: %v", err)
	}
	if err := refreshRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("refresh token indexes: %v", err)
	}
	if err := plantRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("plant indexes: %v", err)
	}
	if err := diagnosisRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("diagnosis indexes: %v", err)
	}
	if err := chatRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("chat indexes: %v", err)
	}

	jwtIssuer := security.NewJWTIssuer(cfg.JWTSecret, "myplantpal-backend", cfg.JWTAccessTTL)
	authService := authuc.NewService(userRepo, refreshRepo, ids, jwtIssuer, cfg.JWTRefreshTTL)

	fertilizerRepo := memory.NewFertilizerRepository(seed.Fertilizers()...)
	productRepo := memory.NewProductRepository(productseed.Products(), productseed.Categories())

	var priceRefresher productuc.PriceRefresherPort
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		priceRefresher = groq.New(key)
		log.Println("Groq price refresh: enabled")
	} else {
		log.Println("Groq price refresh: disabled (set GROQ_API_KEY to enable)")
	}

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
		Addr:         ":" + cfg.Port,
		Handler:      httpapi.NewRouter(deps),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("PlantPal API listening on :%s\n", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}
