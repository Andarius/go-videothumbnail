package main

import (
	"encoding/json"
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

func genThumb(videoPath string, outputPath string) string {
	out, err := exec.Command("ffmpeg",
		"-y",                                 // override output file if exists
		"-hide_banner", "-loglevel", "error", // Less verbose
		"-i", videoPath,
		"-ss", "00:00:01.000",
		"-vframes", "1",
		outputPath).CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		log.Fatalf("Error happened while generating thumb. Error: \"%s\"", output)
	}
	return output
}

func getVideoDimensions(videoPath string) map[string]int16 {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		videoPath).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		log.Fatalf("Error happened while getting video dimensions. Error: \"%s\"", output)
	}

	dimensionsStr := strings.Split(output, "x")

	dimensions := make(map[string]int16)
	width, widthErr := strconv.ParseInt(dimensionsStr[0], 10, 16)
	if widthErr != nil {
		log.Fatalf("Error happened while parsing width. Err: %s", err)
	}
	height, heightErr := strconv.ParseInt(dimensionsStr[1], 10, 16)
	if heightErr != nil {
		log.Fatalf("Error happened while parsing height. Err: %s", err)
	}

	dimensions["width"] = int16(width)
	dimensions["height"] = int16(height)
	return dimensions
}

func genThumbHandler(w http.ResponseWriter, r *http.Request) {
	videoPath := r.FormValue("path")
	outputPath := r.FormValue("output")

	dimensions := getVideoDimensions(videoPath)
	log.Println("Generating thumbnail for video:", videoPath, "to path:", outputPath)
	genThumb(videoPath, outputPath)

	// Writing response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	jsonResp, err := json.Marshal(dimensions)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
	return
}

func main() {
	sentryInitErr := sentry.Init(sentry.ClientOptions{
		Dsn:           os.Getenv("SENTRY_DSN"),
		Release:       os.Getenv("RELEASE"),
		EnableTracing: true,
		// Specify a fixed sample rate:
		// We recommend adjusting this value in production
		TracesSampleRate:   1.0,
		ProfilesSampleRate: 1.0,
	})
	if sentryInitErr != nil {
		log.Fatalf("sentry.Init: %s", sentryInitErr)
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(2 * time.Second)

	// Create an instance of sentryhttp
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	log.Println("Starting server on port 8080...")
	http.HandleFunc("/gen-thumb", sentryHandler.HandleFunc(
		func(res http.ResponseWriter, req *http.Request) {
			genThumbHandler(res, req)
		}))
	http.HandleFunc("/health", sentryHandler.HandleFunc(func(res http.ResponseWriter, req *http.Request) {
		// Write json response
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(`{"status": "ok"}`))
	}))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error happened while starting server. Err: %s", err)
	}
}
