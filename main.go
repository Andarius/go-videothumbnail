package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var supportedFormats = map[string]bool{
	"mov,mp4,m4a,3gp,3g2,mj2": true,
	"matroska,webm":            true,
	"avi":                      true,
	"flv":                      true,
	"mpegts":                   true,
	"asf":                      true,
}

func getVideoFormat(videoPath string) (string, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=format_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("ffprobe failed: %s", output)
	}
	return output, nil
}

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

func writeError(w http.ResponseWriter, r *http.Request, msg string, status int) {
	log.Println("Error:", msg)
	reportError(r, msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func genThumbHandler(w http.ResponseWriter, r *http.Request) {
	videoPath := r.FormValue("path")
	outputPath := r.FormValue("output")

	format, fmtErr := getVideoFormat(videoPath)
	if fmtErr != nil {
		// If the file exists but ffprobe can't identify it, it's not a video
		if _, statErr := os.Stat(videoPath); statErr == nil {
			writeError(w, r, fmt.Sprintf("unsupported file for %s: %s", videoPath, fmtErr), http.StatusBadRequest)
			return
		}
		writeError(w, r, fmt.Sprintf("detect format for %s: %s", videoPath, fmtErr), http.StatusInternalServerError)
		return
	}
	if !supportedFormats[format] {
		writeError(w, r, fmt.Sprintf("unsupported format %q for %s", format, videoPath), http.StatusBadRequest)
		return
	}

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
	_ = json.NewEncoder(w).Encode(dimensions)
}

func main() {
	cleanup := initMonitoring()
	defer cleanup()

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", openapiHandler)
	mux.HandleFunc("/gen-thumb", genThumbHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", monitoringMiddleware(mux))
	if err != nil {
		log.Fatalf("Error happened while starting server. Err: %s", err)
	}
}
