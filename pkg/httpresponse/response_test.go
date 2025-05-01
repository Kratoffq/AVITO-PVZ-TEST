package httpresponse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		payload    interface{}
		wantBody   string
	}{
		{
			name:       "success with data",
			statusCode: http.StatusOK,
			payload: map[string]string{
				"message": "test",
			},
			wantBody: `{"message":"test"}`,
		},
		{
			name:       "success with nil data",
			statusCode: http.StatusOK,
			payload:    nil,
			wantBody:   "",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			payload: map[string]string{
				"error": "not found",
			},
			wantBody: `{"error":"not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			JSON(recorder, tt.statusCode, tt.payload)

			require.Equal(t, tt.statusCode, recorder.Code)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			if tt.wantBody != "" {
				require.JSONEq(t, tt.wantBody, recorder.Body.String())
			} else {
				require.Empty(t, recorder.Body.String())
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantBody   string
	}{
		{
			name:       "error with message",
			statusCode: http.StatusBadRequest,
			message:    "invalid input",
			wantBody:   `{"error":"invalid input"}`,
		},
		{
			name:       "error with empty message",
			statusCode: http.StatusInternalServerError,
			message:    "",
			wantBody:   `{"error":""}`,
		},
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			message:    "resource not found",
			wantBody:   `{"error":"resource not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			Error(recorder, tt.statusCode, tt.message)

			require.Equal(t, tt.statusCode, recorder.Code)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			require.JSONEq(t, tt.wantBody, recorder.Body.String())
		})
	}
}
