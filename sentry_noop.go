//go:build !sentry

package main

import "net/http"

func initMonitoring() func() { return func() {} }

func monitoringMiddleware(handler http.Handler) http.Handler { return handler }

func reportError(_ *http.Request, _ string) {}