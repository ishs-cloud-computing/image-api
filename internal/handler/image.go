package handler

import (
	"encoding/json"
	"fmt"
	"image-server/internal/model"
	"image-server/internal/service"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ImageHandler는 이미지 관련 HTTP 핸들러를 모읍니다.
type ImageHandler struct {
	s3  *service.S3Service
	ddb *service.DynamoService
}

// NewImageHandler는 ImageHandler를 생성합니다.
func NewImageHandler(s3 *service.S3Service, ddb *service.DynamoService) *ImageHandler {
	return &ImageHandler{s3: s3, ddb: ddb}
}

// Upload는 POST /images 요청을 처리합니다.
// multipart/form-data 형식으로 "image" 필드를 받습니다.
func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Upload] 요청 수신: %s %s", r.Method, r.URL.Path)

	// 최대 32MB 멀티파트 파싱
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("[Upload] 멀티파트 파싱 실패: %v", err)
		writeError(w, http.StatusBadRequest, "멀티파트 파싱 실패: "+err.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("[Upload] 'image' 필드 누락: %v", err)
		writeError(w, http.StatusBadRequest, "'image' 필드가 필요합니다")
		return
	}
	defer file.Close()

	// Content-Type 검증 (이미지만 허용)
	contentType := header.Header.Get("Content-Type")
	if !isAllowedContentType(contentType) {
		log.Printf("[Upload] 허용되지 않는 Content-Type: %s", contentType)
		writeError(w, http.StatusUnsupportedMediaType,
			"허용되지 않는 파일 형식입니다. (허용: image/jpeg, image/png, image/gif, image/webp)")
		return
	}

	// 고유 ID 및 S3 키 생성
	id := uuid.NewString()
	ext := extensionFromContentType(contentType)
	s3Key := fmt.Sprintf("images/%s%s", id, ext)
	log.Printf("[Upload] S3 업로드 시작: key=%s, size=%d bytes", s3Key, header.Size)

	// S3 업로드
	s3URL, err := h.s3.Upload(r.Context(), s3Key, file, contentType)
	if err != nil {
		log.Printf("[Upload] S3 업로드 실패: %v", err)
		writeError(w, http.StatusInternalServerError, "S3 업로드 실패: "+err.Error())
		return
	}
	log.Printf("[Upload] S3 업로드 완료: %s", s3URL)

	// DynamoDB에 메타데이터 저장
	img := &model.Image{
		ID:          id,
		FileName:    header.Filename,
		ContentType: contentType,
		SizeBytes:   header.Size,
		S3Key:       s3Key,
		S3URL:       s3URL,
		UploadedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	if err := h.ddb.PutImage(r.Context(), img); err != nil {
		log.Printf("[Upload] DynamoDB 저장 실패: %v", err)
		writeError(w, http.StatusInternalServerError, "메타데이터 저장 실패: "+err.Error())
		return
	}
	log.Printf("[Upload] 완료: id=%s, file=%s", id, header.Filename)

	writeJSON(w, http.StatusCreated, model.UploadResponse{
		Message: "이미지 업로드 성공",
		Image:   *img,
	})
}

// GetMetadata는 GET /images/{id} 요청을 처리합니다.
// DynamoDB에서 메타데이터를 조회하여 반환합니다.
func (h *ImageHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	log.Printf("[GetMetadata] 요청 수신: id=%s", id)

	if id == "" {
		log.Printf("[GetMetadata] ID 누락")
		writeError(w, http.StatusBadRequest, "이미지 ID가 필요합니다")
		return
	}

	img, err := h.ddb.GetImage(r.Context(), id)
	if err != nil {
		log.Printf("[GetMetadata] DynamoDB 조회 실패: %v", err)
		writeError(w, http.StatusInternalServerError, "메타데이터 조회 실패: "+err.Error())
		return
	}
	if img == nil {
		log.Printf("[GetMetadata] 이미지 없음: id=%s", id)
		writeError(w, http.StatusNotFound, "이미지를 찾을 수 없습니다")
		return
	}

	log.Printf("[GetMetadata] 완료: id=%s, file=%s", img.ID, img.FileName)
	writeJSON(w, http.StatusOK, img)
}

// Download는 GET /images/{id}/download 요청을 처리합니다.
// S3에서 이미지 바이트를 스트리밍합니다.
func (h *ImageHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	log.Printf("[Download] 요청 수신: id=%s", id)

	if id == "" {
		log.Printf("[Download] ID 누락")
		writeError(w, http.StatusBadRequest, "이미지 ID가 필요합니다")
		return
	}

	// 먼저 DynamoDB에서 S3 키 조회
	img, err := h.ddb.GetImage(r.Context(), id)
	if err != nil {
		log.Printf("[Download] DynamoDB 조회 실패: %v", err)
		writeError(w, http.StatusInternalServerError, "메타데이터 조회 실패: "+err.Error())
		return
	}
	if img == nil {
		log.Printf("[Download] 이미지 없음: id=%s", id)
		writeError(w, http.StatusNotFound, "이미지를 찾을 수 없습니다")
		return
	}

	log.Printf("[Download] S3 다운로드 시작: key=%s", img.S3Key)

	// S3에서 스트리밍 다운로드
	body, contentType, err := h.s3.Download(r.Context(), img.S3Key)
	if err != nil {
		log.Printf("[Download] S3 다운로드 실패: %v", err)
		writeError(w, http.StatusInternalServerError, "S3 다운로드 실패: "+err.Error())
		return
	}
	defer body.Close()

	log.Printf("[Download] 스트리밍 시작: id=%s, file=%s", id, img.FileName)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, img.FileName))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, body) //nolint:errcheck
	log.Printf("[Download] 스트리밍 완료: id=%s", id)
}

// Health는 GET /health 요청을 처리합니다.
func Health(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Health] 요청 수신")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- 헬퍼 함수 ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}

func isAllowedContentType(ct string) bool {
	allowed := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	for _, a := range allowed {
		if ct == a {
			return true
		}
	}
	return false
}

func extensionFromContentType(ct string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
