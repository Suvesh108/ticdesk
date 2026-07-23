package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const MaxFileSize = 10 * 1024 * 1024 // 10 MB

var AllowedMimeTypes = map[string]bool{
	"image/jpeg":                                true,
	"image/png":                                 true,
	"image/gif":                                 true,
	"image/webp":                                true,
	"application/pdf":                           true,
	"text/plain":                                true,
	"text/csv":                                  true,
	"application/msword":                        true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.ms-excel":                  true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/zip":                           true,
}

type StorageService interface {
	SaveFile(ticketID string, file multipart.File, header *multipart.FileHeader) (string, int64, string, error)
	DeleteFile(filePath string) error
}

type LocalStorageService struct {
	baseDir string
}

func NewLocalStorageService(baseDir string) *LocalStorageService {
	_ = os.MkdirAll(baseDir, 0755)
	return &LocalStorageService{baseDir: baseDir}
}

func (s *LocalStorageService) SaveFile(ticketID string, file multipart.File, header *multipart.FileHeader) (string, int64, string, error) {
	if header.Size > MaxFileSize {
		return "", 0, "", fmt.Errorf("file size exceeds 10MB limit")
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Validate extension/MIME if required
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == ".exe" || ext == ".bat" || ext == ".sh" || ext == ".dll" {
		return "", 0, "", fmt.Errorf("executable files are not allowed")
	}

	ticketDir := filepath.Join(s.baseDir, ticketID)
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		return "", 0, "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	uniqueName := fmt.Sprintf("%s_%s", uuid.New().String(), filepath.Base(header.Filename))
	fullPath := filepath.Join(ticketDir, uniqueName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to write file content: %w", err)
	}

	relPath := filepath.Join("web", "static", "uploads", ticketID, uniqueName)
	return relPath, written, contentType, nil
}

func (s *LocalStorageService) DeleteFile(filePath string) error {
	return os.Remove(filePath)
}
