package httpapi_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/ai"
	"plantpal-backend/internal/infrastructure/repository/memory"
	"plantpal-backend/internal/infrastructure/repository/memory/seed"
	productseed "plantpal-backend/internal/infrastructure/repository/memory/seed/products"
	httpapi "plantpal-backend/internal/interface/http"
	"plantpal-backend/internal/interface/http/authmw"
	v1 "plantpal-backend/internal/interface/http/v1"
	chatuc "plantpal-backend/internal/usecase/chat"
	diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
	fertilizeruc "plantpal-backend/internal/usecase/fertilizer"
	orderuc "plantpal-backend/internal/usecase/order"
	plantuc "plantpal-backend/internal/usecase/plant"
	productuc "plantpal-backend/internal/usecase/product"
)

// newTestRouter wires the test dependency graph with mock auth
func newTestRouter() http.Handler {
	ids := idgen.New()

	plantRepo := memory.NewPlantRepository()
	fertilizerRepo := memory.NewFertilizerRepository(seed.Fertilizers()...)
	diagnosisRepo := memory.NewDiagnosisRepository()
	chatRepo := memory.NewChatRepository()
	productRepo := memory.NewProductRepository(productseed.Products(), productseed.Categories())
	orderRepo := memory.NewOrderRepository()

	noAuth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := authmw.WithUserID(r.Context(), "test_user_id")
			next(w, r.WithContext(ctx))
		}
	}

	deps := v1.Dependencies{
		PlantService:      plantuc.NewService(plantRepo, ids),
		FertilizerService: fertilizeruc.NewService(fertilizerRepo, ids),
		DiagnosisService:  diagnosisuc.NewService(diagnosisRepo, ai.NewMockDiagnosisProvider(), ids),
		ChatService:       chatuc.NewService(chatRepo, ai.NewMockChatReplyProvider(), ids),
		ProductService:    productuc.NewService(productRepo, nil),
		OrderService:      orderuc.NewService(orderRepo, productRepo, ids),
		RequireAuth:       noAuth,
	}

	return httpapi.NewRouter(deps, httpapi.GoogleOAuthConfig{})
}

type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func doRequest(t *testing.T, handler http.Handler, method, path, lang string, body any) (*httptest.ResponseRecorder, apiEnvelope) {
	t.Helper()
	var bodyReader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var env apiEnvelope
	if rec.Body.Len() > 0 && strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		_ = json.Unmarshal(rec.Body.Bytes(), &env)
	}
	return rec, env
}

func TestHealth(t *testing.T) {
	router := newTestRouter()
	rec, env := doRequest(t, router, http.MethodGet, "/health", "", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]string
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("status = %q, want %q", data["status"], "ok")
	}
}

func TestDocsEndpoints(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/docs status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "swagger-ui") {
		t.Errorf("/docs body does not look like the Swagger UI page")
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/openapi.yaml status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "openapi:") {
		t.Errorf("/openapi.yaml body does not look like an OpenAPI spec")
	}
}

func TestFertilizers_ListIncludesSeedData(t *testing.T) {
	router := newTestRouter()
	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/fertilizers", "", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var items []map[string]any
	_ = json.Unmarshal(env.Data, &items)
	if len(items) != 3 {
		t.Errorf("got %d fertilizers, want 3 seeded entries", len(items))
	}
}

func TestFertilizers_Search(t *testing.T) {
	router := newTestRouter()
	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/fertilizers?q=potassium", "", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var items []map[string]any
	_ = json.Unmarshal(env.Data, &items)
	if len(items) != 1 {
		t.Errorf("got %d results for potassium search, want 1", len(items))
	}
}

func TestFertilizers_LocalizedByAcceptLanguage(t *testing.T) {
	router := newTestRouter()

	_, enEnv := doRequest(t, router, http.MethodGet, "/api/v1/fertilizers?q=nitrogen", "en", nil)
	var enItems []map[string]any
	_ = json.Unmarshal(enEnv.Data, &enItems)

	_, bnEnv := doRequest(t, router, http.MethodGet, "/api/v1/fertilizers?q=nitrogen", "bn", nil)
	var bnItems []map[string]any
	_ = json.Unmarshal(bnEnv.Data, &bnItems)

	if len(enItems) != 1 || len(bnItems) != 1 {
		t.Fatalf("expected exactly one nitrogen fertilizer per language, got en=%d bn=%d", len(enItems), len(bnItems))
	}
	if enItems[0]["name"] == bnItems[0]["name"] {
		t.Errorf("expected different names for en/bn, both were %v", enItems[0]["name"])
	}
	if enItems[0]["id"] != bnItems[0]["id"] {
		t.Errorf("expected the same fertilizer id regardless of language")
	}
}

func TestFertilizers_CreateAndGet(t *testing.T) {
	router := newTestRouter()

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/fertilizers", "", map[string]string{
		"name":         "Custom Mix",
		"category":     "Custom",
		"instructions": "Mix it up.",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(env.Data, &created)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("created fertilizer has no id")
	}

	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/fertilizers/"+id, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", rec.Code, http.StatusOK)
	}
	var fetched map[string]any
	_ = json.Unmarshal(env.Data, &fetched)
	if fetched["name"] != "Custom Mix" {
		t.Errorf("fetched name = %v, want %q", fetched["name"], "Custom Mix")
	}
}

func TestPlants_CreateGeneratesRoadmap(t *testing.T) {
	router := newTestRouter()
	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/plants", "", map[string]string{
		"nickname": "Rose",
		"species":  "Water based",
		"ageStage": "mature",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(env.Data, &created)
	note, _ := created["fertilizerNote"].(string)
	if !strings.Contains(note, "Potassium") {
		t.Errorf("fertilizerNote = %q, want it to mention Potassium for mature stage", note)
	}
}

func TestPlants_Create_MissingName(t *testing.T) {
	router := newTestRouter()
	rec, _ := doRequest(t, router, http.MethodPost, "/api/v1/plants", "", map[string]string{"species": "cactus"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPlants_Get_NotFound(t *testing.T) {
	router := newTestRouter()
	rec, _ := doRequest(t, router, http.MethodGet, "/api/v1/plants/does-not-exist", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDiagnoses_CreateSucceeds(t *testing.T) {
	router := newTestRouter()
	imageB64 := base64.StdEncoding.EncodeToString([]byte("fake-image-bytes"))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/diagnoses", "", map[string]string{
		"imageBase64": imageB64,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(env.Data, &created)
	if created["issue"] == "" || created["cure"] == "" || created["disclaimer"] == "" {
		t.Errorf("expected issue/cure/disclaimer to be populated, got %+v", created)
	}
}

func TestChat_SendAndList(t *testing.T) {
	router := newTestRouter()

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/chat/messages", "", map[string]string{
		"sessionId": "s1",
		"content":   "Why are my leaves yellow?",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var msgs []map[string]any
	_ = json.Unmarshal(env.Data, &msgs)
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2 (user + assistant)", len(msgs))
	}

	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/chat/messages?sessionId=s1", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}
	var history []map[string]any
	_ = json.Unmarshal(env.Data, &history)
	if len(history) != 2 {
		t.Errorf("history has %d messages, want 2", len(history))
	}
}

func TestCORSHeadersPresent(t *testing.T) {
	router := newTestRouter()
	rec, _ := doRequest(t, router, http.MethodGet, "/health", "", nil)
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", rec.Header().Get("Access-Control-Allow-Origin"), "*")
	}
}

func orderRequestBody() map[string]any {
	return map[string]any{
		"items":            []map[string]any{{"productId": "water-can-ongkoor", "quantity": 2}},
		"shippingName":     "Promitee",
		"shippingPhone":    "01700000000",
		"shippingAddress":  "House 1, Road 2, Dhaka",
		"deliveryOptionId": "standard",
		"paymentMethodId":  "cod",
	}
}

func TestOrders_CreateListGet(t *testing.T) {
	router := newTestRouter()

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/orders", "", orderRequestBody())
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(env.Data, &created)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("created order has no id")
	}
	// water-can-ongkoor is seeded at 550 BDT; quantity 2 -> 1100 subtotal.
	if subtotal, _ := created["subtotalBdt"].(float64); subtotal != 1100 {
		t.Errorf("subtotalBdt = %v, want 1100 (550 * 2)", created["subtotalBdt"])
	}
	if created["status"] != "processing" {
		t.Errorf("status = %v, want processing", created["status"])
	}

	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/orders", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}
	var list []map[string]any
	_ = json.Unmarshal(env.Data, &list)
	if len(list) != 1 {
		t.Fatalf("got %d orders, want 1", len(list))
	}

	rec, env = doRequest(t, router, http.MethodGet, "/api/v1/orders/"+id, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", rec.Code, http.StatusOK)
	}
	var fetched map[string]any
	_ = json.Unmarshal(env.Data, &fetched)
	if fetched["id"] != id {
		t.Errorf("fetched id = %v, want %q", fetched["id"], id)
	}
}

func TestOrders_RejectsUnknownProduct(t *testing.T) {
	router := newTestRouter()
	body := orderRequestBody()
	body["items"] = []map[string]any{{"productId": "does-not-exist", "quantity": 1}}

	rec, _ := doRequest(t, router, http.MethodPost, "/api/v1/orders", "", body)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestOrders_LocalizedByAcceptLanguage mirrors TestFertilizers_LocalizedByAcceptLanguage:
// the stored delivery option id never changes, only its displayed title/eta.
func TestOrders_LocalizedByAcceptLanguage(t *testing.T) {
	router := newTestRouter()
	body := orderRequestBody()

	_, enEnv := doRequest(t, router, http.MethodPost, "/api/v1/orders", "en", body)
	var enOrder map[string]any
	_ = json.Unmarshal(enEnv.Data, &enOrder)
	enDelivery, _ := enOrder["delivery"].(map[string]any)

	_, bnEnv := doRequest(t, router, http.MethodPost, "/api/v1/orders", "bn", body)
	var bnOrder map[string]any
	_ = json.Unmarshal(bnEnv.Data, &bnOrder)
	bnDelivery, _ := bnOrder["delivery"].(map[string]any)

	if enDelivery["id"] != bnDelivery["id"] {
		t.Errorf("expected the same delivery option id regardless of language")
	}
	if enDelivery["title"] == bnDelivery["title"] {
		t.Errorf("expected different delivery titles for en/bn, both were %v", enDelivery["title"])
	}
	if bnDelivery["title"] != "স্ট্যান্ডার্ড ডেলিভারি" {
		t.Errorf("bn delivery title = %v, want স্ট্যান্ডার্ড ডেলিভারি", bnDelivery["title"])
	}
}
