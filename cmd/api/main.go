package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"plantpal-backend/internal/config"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/ai"
	"plantpal-backend/internal/infrastructure/ai/gemini"
	"plantpal-backend/internal/infrastructure/ai/groq"
	"plantpal-backend/internal/infrastructure/repository/memory"
	"plantpal-backend/internal/infrastructure/repository/memory/seed"
	productseed "plantpal-backend/internal/infrastructure/repository/memory/seed/products"
	mongorepo "plantpal-backend/internal/infrastructure/repository/mongo"
	"plantpal-backend/internal/infrastructure/security"
	httpapi "plantpal-backend/internal/interface/http"
	"plantpal-backend/internal/interface/http/authmw"
	v1 "plantpal-backend/internal/interface/http/v1"
	authuc "plantpal-backend/internal/usecase/auth"
	chatuc "plantpal-backend/internal/usecase/chat"
	diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "plantpal-backend/internal/usecase/fertilizer"
	orderuc "plantpal-backend/internal/usecase/order"
	plantuc "plantpal-backend/internal/usecase/plant"
	productuc "plantpal-backend/internal/usecase/product"
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
	orderRepo := mongorepo.NewOrderRepository(db)
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
	if err := orderRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("order indexes: %v", err)
	}

	jwtIssuer := security.NewJWTIssuer(cfg.JWTSecret, "plantpal-backend", cfg.JWTAccessTTL)
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
		OrderService:      orderuc.NewService(orderRepo, productRepo, ids),
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
