package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	testcases := []struct {
		rr             *httptest.ResponseRecorder
		data           map[string]string
		headers        http.Header
		expectedCode   int
		expectedStatus string
		expectedData   string
	}{
		{
			rr:             httptest.NewRecorder(),
			data:           map[string]string{"key": "value"},
			headers:        http.Header{"Test-Header": []string{"test-value"}},
			expectedCode:   http.StatusOK,
			expectedStatus: "OK",
			expectedData:   `{"key":"value"}` + "\n",
		},
	}

	// Suppress output from logger
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	slog.SetDefault(logger)
	app := &application{}

	for _, tc := range testcases {
		err := app.writeJSON(tc.rr, tc.expectedCode, tc.data, tc.headers)
		if err != nil {
			t.Errorf("Expected nil but got '%v'", err)
		}
		if tc.rr.Code != tc.expectedCode {
			t.Errorf("Expected code %d but got %d", tc.expectedCode, tc.rr.Code)
		}
		if tc.rr.Body.String() != tc.expectedData {
			t.Errorf("Expected body '%s' but got '%s'", tc.expectedData, tc.rr.Body.String())
		}
		if tc.rr.Header().Get("Test-Header") != "test-value" {
			t.Errorf(
				"Expected header '%s' but got '%s'",
				tc.rr.Header().Get("Test-Header"),
				"test-value",
			)
		}
		if tc.rr.Header().Get("Content-Type") != "application/json" {
			t.Errorf(
				"Expected header '%s' but got '%s'",
				tc.rr.Header().Get("Content-Type"),
				"application/json",
			)
		}
	}
}
