package v1

import (
"encoding/base64"
"encoding/json"
"errors"
"fmt"
"io"
"net/http"
"strings"

"plantpal-backend/internal/domain/apperr"
"plantpal-backend/internal/domain/diagnosis"
"plantpal-backend/internal/interface/http/authmw"
"plantpal-backend/internal/interface/http/reqlocale"
"plantpal-backend/internal/interface/http/respond"
diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
)

const defaultMaxImageBytes int64 = 8 << 20 // 8 MB

var allowedImageTypes = map[string]bool{
"image/jpeg": true,
"image/png":  true,
"image/webp": true,
}

type DiagnosisHandler struct {
svc           *diagnosisuc.Service
maxImageBytes int64
}

func NewDiagnosisHandler(svc *diagnosisuc.Service, maxImageBytes int64) *DiagnosisHandler {
if maxImageBytes <= 0 {
maxImageBytes = defaultMaxImageBytes
}
return &DiagnosisHandler{svc: svc, maxImageBytes: maxImageBytes}
}

type createDiagnosisRequest struct {
PlantID     *string `json:"plantId,omitempty"`
ImageBase64 string  `json:"imageBase64"`
// SessionID, when set, also records the scan in that chat session.
SessionID string `json:"sessionId,omitempty"`
// Note is the user's optional caption for the photo.
Note string `json:"note,omitempty"`
}

type upload struct {
plantID     *string
data        []byte
contentType string
sessionID   string
note        string
}

// Create accepts either multipart/form-data (fields: "image" file, optional
// "plantId") or the legacy JSON body with imageBase64. Both are capped at
// the configured max image size.
func (h *DiagnosisHandler) Create(w http.ResponseWriter, r *http.Request) {
var (
up  upload
err error
)
if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
up, err = h.readMultipart(w, r)
} else {
up, err = h.readJSON(w, r)
}
if err != nil {
respond.Error(w, err)
return
}

userID, _ := authmw.UserID(r.Context())
d, err := h.svc.Analyze(r.Context(), diagnosisuc.AnalyzeInput{
UserID:      userID,
PlantID:     up.plantID,
ImageData:   up.data,
ContentType: up.contentType,
SessionID:   up.sessionID,
Note:        up.note,
Lang:        reqlocale.Resolve(r),
})
if err != nil {
respond.Error(w, err)
return
}
respond.JSON(w, http.StatusCreated, presentDiagnosis(*d, reqlocale.Resolve(r)))
}

func (h *DiagnosisHandler) readMultipart(w http.ResponseWriter, r *http.Request) (upload, error) {
// Small allowance on top of the image limit for boundaries and text fields.
r.Body = http.MaxBytesReader(w, r.Body, h.maxImageBytes+(64<<10))
if err := r.ParseMultipartForm(1 << 20); err != nil {
return upload{}, h.bodyErr(err, "invalid multipart form")
}
defer r.MultipartForm.RemoveAll()

file, _, err := r.FormFile("image")
if err != nil {
return upload{}, fmt.Errorf("%w: multipart field \"image\" is required", apperr.ErrInvalidInput)
}
defer file.Close()

data, err := io.ReadAll(io.LimitReader(file, h.maxImageBytes+1))
if err != nil {
return upload{}, h.bodyErr(err, "could not read image")
}
if int64(len(data)) > h.maxImageBytes {
return upload{}, h.tooLarge()
}
ct := http.DetectContentType(data)
if !allowedImageTypes[ct] {
return upload{}, fmt.Errorf("%w: unsupported image type (use JPEG, PNG or WebP)", apperr.ErrInvalidInput)
}

up := upload{
data:        data,
contentType: ct,
sessionID:   strings.TrimSpace(r.FormValue("sessionId")),
note:        strings.TrimSpace(r.FormValue("note")),
}
if v := strings.TrimSpace(r.FormValue("plantId")); v != "" {
up.plantID = &v
}
return up, nil
}

func (h *DiagnosisHandler) readJSON(w http.ResponseWriter, r *http.Request) (upload, error) {
// base64 inflates data by about 4/3.
r.Body = http.MaxBytesReader(w, r.Body, h.maxImageBytes*4/3+4096)
var req createDiagnosisRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
return upload{}, h.bodyErr(err, "invalid JSON body")
}
data, err := base64.StdEncoding.DecodeString(req.ImageBase64)
if err != nil {
return upload{}, fmt.Errorf("%w: imageBase64 is not valid base64", apperr.ErrInvalidInput)
}
if int64(len(data)) > h.maxImageBytes {
return upload{}, h.tooLarge()
}
// Legacy clients were never type-checked, so an unrecognised payload is
// stored as JPEG instead of being rejected.
ct := http.DetectContentType(data)
if !allowedImageTypes[ct] {
ct = "image/jpeg"
}
return upload{
plantID:     req.PlantID,
data:        data,
contentType: ct,
sessionID:   strings.TrimSpace(req.SessionID),
note:        strings.TrimSpace(req.Note),
}, nil
}

func (h *DiagnosisHandler) tooLarge() error {
return fmt.Errorf("%w: image must be at most %d MB", apperr.ErrPayloadTooLarge, h.maxImageBytes>>20)
}

func (h *DiagnosisHandler) bodyErr(err error, msg string) error {
var tooBig *http.MaxBytesError
if errors.As(err, &tooBig) {
return h.tooLarge()
}
return fmt.Errorf("%w: %s", apperr.ErrInvalidInput, msg)
}

func (h *DiagnosisHandler) List(w http.ResponseWriter, r *http.Request) {
var plantID *string
if q := r.URL.Query().Get("plantId"); q != "" {
plantID = &q
}
userID, _ := authmw.UserID(r.Context())
items, err := h.svc.List(r.Context(), userID, plantID)
if err != nil {
respond.Error(w, err)
return
}
lang := reqlocale.Resolve(r)
out := make([]diagnosis.Diagnosis, len(items))
for i, item := range items {
out[i] = presentDiagnosis(*item, lang)
}
respond.JSON(w, http.StatusOK, out)
}

func (h *DiagnosisHandler) Get(w http.ResponseWriter, r *http.Request) {
userID, _ := authmw.UserID(r.Context())
d, err := h.svc.Get(r.Context(), r.PathValue("id"), userID)
if err != nil {
respond.Error(w, err)
return
}
respond.JSON(w, http.StatusOK, presentDiagnosis(*d, reqlocale.Resolve(r)))
}

// MarkTreated flags a scan as treated (POST /diagnoses/{id}/treated).
func (h *DiagnosisHandler) MarkTreated(w http.ResponseWriter, r *http.Request) {
userID, _ := authmw.UserID(r.Context())
d, err := h.svc.MarkTreated(r.Context(), r.PathValue("id"), userID)
if err != nil {
respond.Error(w, err)
return
}
respond.JSON(w, http.StatusOK, presentDiagnosis(*d, reqlocale.Resolve(r)))
}

// Image streams the photo saved with a diagnosis (Bearer auth required).
func (h *DiagnosisHandler) Image(w http.ResponseWriter, r *http.Request) {
userID, _ := authmw.UserID(r.Context())
rc, contentType, err := h.svc.Image(r.Context(), r.PathValue("id"), userID)
if err != nil {
respond.Error(w, err)
return
}
defer rc.Close()
w.Header().Set("Content-Type", contentType)
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("Cache-Control", "private, max-age=86400")
_, _ = io.Copy(w, rc)
}

// presentDiagnosis localizes d and fills in the image URL when a photo exists.
func presentDiagnosis(d diagnosis.Diagnosis, lang string) diagnosis.Diagnosis {
out := d.Localized(lang)
if out.ImageKey != "" {
out.ImageURL = "/api/v1/diagnoses/" + out.ID + "/image"
}
return out
}