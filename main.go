package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type Response struct {
    dimensions string
}

func genThumb(videoPath string, outputPath string) string {

    out, err := exec.Command( "ffmpeg",
	"-y",  // override output file if exists
	"-hide_banner", "-loglevel", "error",  // Less verbose
	"-i", videoPath,
	"-ss", "00:00:01.000",
	"-vframes", "1",
	outputPath).Output()


    if err != nil {
        log.Fatal(err)
    }
	return strings.TrimSpace(string(out))
}


func getVideoDimensions(videoPath string) map[string]int16 {

    out, err := exec.Command(
		"ffprobe",
        "-v", "error",
        "-select_streams", "v:0",
        "-show_entries", "stream=width,height",
        "-of", "csv=s=x:p=0",
        videoPath).Output()

    if err != nil {
        log.Fatal(err)
    }

	dimensionsStr := strings.Split(strings.TrimSpace(string(out)), "x")

	dimensions := make(map[string]int16)
	width, widthErr := strconv.ParseInt(dimensionsStr[0], 10, 16)
	if widthErr != nil {
        log.Fatal(err)
    }
	height, heightErr := strconv.ParseInt(dimensionsStr[1], 10, 16)
	if heightErr != nil {
		log.Fatal(err)
	}

	dimensions["width"] = int16(width)
	dimensions["height"] = int16(height)
	return dimensions
}

func genThumbHandler(w http.ResponseWriter, r *http.Request) {
	videoPath := r.FormValue("path")
	outputPath := r.FormValue("output")

	dimensions := getVideoDimensions(videoPath)

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
    http.HandleFunc("/gen-thumb", func(res http.ResponseWriter, req *http.Request) {
        genThumbHandler(res, req)
    })
    http.ListenAndServe(":8080", nil)
}
