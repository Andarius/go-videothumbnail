//go:build sentry

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

var sentryEnabled bool

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		log.Printf("Warning: invalid %s value %q, using default %g", key, v, fallback)
		return fallback
	}
	return f
}

// initMonitoring initializes error monitoring (Sentry) if SENTRY_DSN is set.
// Returns a cleanup function to flush buffered events.
func initMonitoring() func() {
	dsn := os.Getenv("SENTRY_DSN")
	sentryEnabled = dsn != ""

	if !sentryEnabled {
		return func() {}
	}

	sentryInitErr := sentry.Init(sentry.ClientOptions{
		Dsn:                dsn,
		Release:            os.Getenv("RELEASE_STAGE"),
		EnableTracing:      true,
		TracesSampleRate:   getEnvFloat("SENTRY_TRACES_SAMPLE_RATE", 1.0),
		ProfilesSampleRate: getEnvFloat("SENTRY_PROFILES_SAMPLE_RATE", 1.0),
	})
	if sentryInitErr != nil {
		log.Fatalf("sentry.Init: %s", sentryInitErr)
	}
	log.Println("Sentry enabled")

	return func() { sentry.Flush(2 * time.Second) }
}

func monitoringMiddleware(handler http.Handler) http.Handler {
	if !sentryEnabled {
		return handler
	}
	return sentryhttp.New(sentryhttp.Options{}).Handle(handler)
}

func reportError(r *http.Request, msg string) {
	if !sentryEnabled {
		return
	}
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureMessage(msg)
	}
}