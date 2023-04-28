package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
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


	fmt.Printf("Command finished with error: %s", err)
	fmt.Printf("Output: %s", string(out))
    if err != nil {
        log.Fatal(err)
    }
	return strings.TrimSpace(string(out))
}


func getVideoDimensions(videoPath string) string {

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
	return strings.TrimSpace(string(out))
}

func genThumbHandler(w http.ResponseWriter, r *http.Request) {
	videoPath := r.FormValue("path")
	outputPath := r.FormValue("output")

	videoDimensions := getVideoDimensions(videoPath)

	genThumb(videoPath, outputPath)

	// Writing response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)

	resp["dimensions"] = videoDimensions
	jsonResp, err := json.Marshal(resp)
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
