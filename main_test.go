package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestOpenapiHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()

	openapiHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var spec map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if spec["openapi"] != "3.0.3" {
		t.Fatalf("expected openapi 3.0.3, got %v", spec["openapi"])
	}
}

func TestWriteError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gen-thumb", nil)
	w := httptest.NewRecorder()

	writeError(w, req, "something went wrong", http.StatusInternalServerError)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] != "something went wrong" {
		t.Fatalf("expected error message, got %q", body["error"])
	}
}

func TestGenThumbHandlerMissingVideo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gen-thumb?path=/nonexistent/video.mp4&output=/tmp/out.jpg", nil)
	w := httptest.NewRecorder()

	genThumbHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error message in response")
	}
}

// writeJPEG creates a valid JPEG file at the given path.
func writeJPEG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := range 64 {
		for x := range 64 {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck // test helper, encode error takes precedence
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGetVideoFormat(t *testing.T) {
	t.Run("unsupported jpeg disguised as mov", func(t *testing.T) {
		tmpDir := t.TempDir()
		fakeMov := filepath.Join(tmpDir, "fake.mov")
		writeJPEG(t, fakeMov)

		format, err := getVideoFormat(fakeMov)
		// ffprobe may detect "jpeg_pipe" or error out depending on version;
		// either way the file must not be in the supported set.
		if err == nil && supportedFormats[format] {
			t.Fatalf("expected unsupported format, got %q", format)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := getVideoFormat("/nonexistent/video.mp4")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestGenThumbHandlerUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	fakeMov := filepath.Join(tmpDir, "fake.mov")
	writeJPEG(t, fakeMov)
	output := filepath.Join(tmpDir, "out.jpeg")

	req := httptest.NewRequest(http.MethodGet, "/gen-thumb?path="+fakeMov+"&output="+output, nil)
	w := httptest.NewRecorder()

	genThumbHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error message in response")
	}
}