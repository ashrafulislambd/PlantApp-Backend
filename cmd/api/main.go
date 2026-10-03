// Command api runs the PlantPal HTTP API.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata" // IANA zone names work even in a minimal container image

	"plantpal-backend/internal/config"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/domain/notification"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/ai"
	"plantpal-backend/internal/infrastructure/ai/gemini"
	"plantpal-backend/internal/infrastructure/ai/groq"
	"plantpal-backend/internal/infrastructure/push"
	"plantpal-backend/internal/infrastructure/push/fcm"
	"plantpal-backend/internal/infrastructure/repository/memory"
	"plantpal-backend/internal/infrastructure/repository/memory/seed"
	productseed "plantpal-backend/internal/infrastructure/repository/memory/seed/products"
	mongorepo "plantpal-backend/internal/infrastructure/repository/mongo"
	"plantpal-backend/internal/infrastructure/security"
	"plantpal-backend/internal/infrastructure/storage/localfs"
	httpapi "plantpal-backend/internal/interface/http"
	"plantpal-backend/internal/interface/http/authmw"
	v1 "plantpal-backend/internal/interface/http/v1"
	authuc "plantpal-backend/internal/usecase/auth"
	chatuc "plantpal-backend/internal/usecase/chat"
	diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "plantpal-backend/internal/usecase/fertilizer"
	notificationuc "plantpal-backend/internal/usecase/notification"
	orderuc "plantpal-backend/internal/usecase/order"
	plantuc "plantpal-backend/internal/usecase/plant"
	productuc "plantpal-backend/internal/usecase/product"
)

// buildPushSender picks FCM when credentials are configured, otherwise a
// log-only sender so the reminder flow still runs in local dev.
func buildPushSender(cfg config.Config) notification.Sender {
	var (
		sender *fcm.Sender
		err    error
	)
	switch {
	case cfg.FCMCredentialsJSON != "":
		sender, err = fcm.NewFromJSON([]byte(cfg.FCMCredentialsJSON))
	case cfg.FCMCredentialsFile != "":
		sender, err = fcm.NewFromFile(cfg.FCMCredentialsFile)
	default:
		log.Println("Push notifications: log-only (set FCM_CREDENTIALS_FILE or FCM_CREDENTIALS_JSON to enable FCM)")
		return push.LogSender{}
	}
	if err != nil {
		log.Fatalf("fcm: %v", err)
	}
	log.Println("Push notifications: FCM enabled")
	return sender
}

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
	plantEventRepo := mongorepo.NewPlantEventRepository(db)
	diagnosisRepo := mongorepo.NewDiagnosisRepository(db)
	chatRepo := mongorepo.NewChatRepository(db)
	orderRepo := mongorepo.NewOrderRepository(db)
	notificationRepo := mongorepo.NewNotificationRepository(db)
	if err := userRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("user indexes: %v", err)
	}
	if err := refreshRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("refresh token indexes: %v", err)
	}
	if err := plantRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("plant indexes: %v", err)
	}
	if err := plantRepo.EnsureDueIndexes(mongoCtx); err != nil {
		log.Fatalf("plant due indexes: %v", err)
	}
	if err := plantEventRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("plant event indexes: %v", err)
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
	if err := notificationRepo.EnsureIndexes(mongoCtx); err != nil {
		log.Fatalf("notification indexes: %v", err)
	}

	jwtIssuer := security.NewJWTIssuer(cfg.JWTSecret, "plantpal-backend", cfg.JWTAccessTTL)
	authService := authuc.NewService(userRepo, refreshRepo, ids, jwtIssuer, cfg.JWTRefreshTTL)
	if len(cfg.GoogleClientIDs) > 0 {
		authService.SetGoogleVerifier(security.NewGoogleVerifier(cfg.GoogleClientIDs))
		log.Println("Google sign-in: enabled")
	} else {
		log.Println("Google sign-in: disabled (set GOOGLE_CLIENT_ID to enable)")
	}

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
	chatEntries = append(chatEntries, ai.ChatProviderEntry{Name: "mock", Provider: ai.NewMockChatReplyProvider(), Placeholder: true})
	diagnosisEntries = append(diagnosisEntries, ai.DiagnosisProviderEntry{Name: "mock", Provider: ai.NewMockDiagnosisProvider(), Placeholder: true})

	log.Printf("AI providers: gemini=%v groq=%v (mock always available as last resort)",
		cfg.GeminiAPIKey != "", cfg.GroqAPIKey != "")

	var chatReplyProvider chat.ReplyProvider = ai.NewFallbackChatProvider(chatEntries...)
	var diagnosisProvider diagnosis.Provider = ai.NewFallbackDiagnosisProvider(diagnosisEntries...)

	diagnosisService := diagnosisuc.NewService(diagnosisRepo, diagnosisProvider, ids)
	imageStore, err := localfs.New(cfg.UploadDir)
	if err != nil {
		log.Fatalf("upload dir: %v", err)
	}
	diagnosisService.SetImageStore(imageStore)

	// Chat and scans know about each other through small ports: a scan sent
	// from the chat is recorded in the chat session, and a message may carry a
	// scan (user-scoped lookup) that later turns remember.
	chatService := chatuc.NewService(chatRepo, chatReplyProvider, ids)
	chatService.SetScanLookup(diagnosisRepo)
	diagnosisService.SetChatRecorder(chatService)
	log.Printf("Diagnosis photos: saved to %s (max %d bytes)", cfg.UploadDir, cfg.MaxUploadBytes)

	notificationService := notificationuc.NewService(notificationRepo, buildPushSender(cfg), plantRepo)
	if cfg.ReminderInterval > 0 {
		go notificationService.Run(context.Background(), cfg.ReminderInterval)
		log.Printf("Care reminders: checking every %s", cfg.ReminderInterval)
	} else {
		log.Println("Care reminders: disabled (REMINDER_INTERVAL=0)")
	}

	plantService := plantuc.NewService(plantRepo, ids)
	plantService.SetEventRepository(plantEventRepo)
	plantService.SetImageStore(imageStore)
	plantService.SetScanSource(diagnosisuc.NewPlantScanSource(diagnosisRepo))
	diagnosisService.SetPlantEventRecorder(plantService)

	// Build identifier chain (Gemini first, mock fallback)
	var identifierEntries []ai.IdentifierProviderEntry
	if cfg.GeminiAPIKey != "" {
		geminiClient := gemini.New(cfg.GeminiAPIKey, cfg.GeminiModel)
		identifierEntries = append(identifierEntries, ai.IdentifierProviderEntry{Name: "gemini", Provider: geminiClient})
	}
	identifierEntries = append(identifierEntries, ai.IdentifierProviderEntry{Name: "mock", Provider: ai.NewMockPlantIdentifier(), Placeholder: true})
	plantService.SetIdentifier(ai.NewFallbackPlantIdentifier(identifierEntries...))

	deps := v1.Dependencies{
		PlantService:        plantService,
		FertilizerService:   fertilizeruc.NewService(fertilizerRepo, ids),
		DiagnosisService:    diagnosisService,
		ChatService:         chatService,
		ProductService:      productuc.NewService(productRepo, priceRefresher),
		OrderService:        orderuc.NewService(orderRepo, productRepo, ids),
		AuthService:         authService,
		NotificationService: notificationService,
		MaxImageBytes:       cfg.MaxUploadBytes,
		RequireAuth:         authmw.RequireAuth(jwtIssuer),
	}

	var googleOAuthClientID string
	if len(cfg.GoogleClientIDs) > 0 {
		googleOAuthClientID = cfg.GoogleClientIDs[0]
	}
	googleOAuth := httpapi.GoogleOAuthConfig{
		ClientID:     googleOAuthClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURI:  cfg.GoogleRedirectURI,
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpapi.NewRouter(deps, googleOAuth),
		// Photo uploads from a phone on mobile data can be slow.
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("PlantPal API listening on :%s\n", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}
