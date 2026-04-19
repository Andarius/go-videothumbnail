package main

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var sentryEnabled bool

func genThumb(videoPath string, outputPath string) (string, error) {
	out, err := exec.Command("ffmpeg",
		"-y",                                 // override output file if exists
		"-hide_banner", "-loglevel", "error", // Less verbose
		"-i", videoPath,
		"-ss", "00:00:01.000",
		"-vframes", "1",
		outputPath).CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		return output, fmt.Errorf("ffmpeg failed: %s", output)
	}
	if _, statErr := os.Stat(outputPath); statErr != nil {
		return output, fmt.Errorf("thumbnail not created at %s", outputPath)
	}
	return output, nil
}

func getVideoDimensions(videoPath string) (map[string]int16, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		videoPath).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %s", output)
	}

	dimensionsStr := strings.Split(output, "x")
	if len(dimensionsStr) != 2 {
		return nil, fmt.Errorf("unexpected ffprobe output: %q", output)
	}

	width, widthErr := strconv.ParseInt(dimensionsStr[0], 10, 16)
	if widthErr != nil {
		return nil, fmt.Errorf("parsing width: %w", widthErr)
	}
	height, heightErr := strconv.ParseInt(dimensionsStr[1], 10, 16)
	if heightErr != nil {
		return nil, fmt.Errorf("parsing height: %w", heightErr)
	}

	dimensions := make(map[string]int16)
	dimensions["width"] = int16(width)
	dimensions["height"] = int16(height)
	return dimensions, nil
}

func captureMessage(r *http.Request, msg string) {
	if !sentryEnabled {
		return
	}
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureMessage(msg)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, msg string, status int) {
	log.Println("Error:", msg)
	captureMessage(r, msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func genThumbHandler(w http.ResponseWriter, r *http.Request) {
	videoPath := r.FormValue("path")
	outputPath := r.FormValue("output")

	dimensions, dimErr := getVideoDimensions(videoPath)
	if dimErr != nil {
		writeError(w, r, fmt.Sprintf("get dimensions for %s: %s", videoPath, dimErr), http.StatusInternalServerError)
		return
	}

	log.Println("Generating thumbnail for video:", videoPath, "to path:", outputPath)
	_, thumbErr := genThumb(videoPath, outputPath)
	if thumbErr != nil {
		writeError(w, r, fmt.Sprintf("generate thumbnail for %s: %s", videoPath, thumbErr), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dimensions)
}

func main() {
	dsn := os.Getenv("SENTRY_DSN")
	sentryEnabled = dsn != ""

	if sentryEnabled {
		sentryInitErr := sentry.Init(sentry.ClientOptions{
			Dsn:                dsn,
			Release:            os.Getenv("RELEASE_STAGE"),
			EnableTracing:      true,
			TracesSampleRate:   1.0,
			ProfilesSampleRate: 1.0,
		})
		if sentryInitErr != nil {
			log.Fatalf("sentry.Init: %s", sentryInitErr)
		}
		defer sentry.Flush(2 * time.Second)
		log.Println("Sentry enabled")
	}

	mux := http.NewServeMux()

	// Wrap handlers with Sentry middleware only if enabled
	if sentryEnabled {
		sentryHandler := sentryhttp.New(sentryhttp.Options{})
		mux.HandleFunc("/gen-thumb", sentryHandler.HandleFunc(genThumbHandler))
		mux.HandleFunc("/health", sentryHandler.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
	} else {
		mux.HandleFunc("/gen-thumb", genThumbHandler)
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		})
	}

	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Error happened while starting server. Err: %s", err)
	}
}