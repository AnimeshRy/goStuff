package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculatorAPI(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		request        CalculationRequest
		expectedStatus int
		expectedResult *int
		expectedError  string
	}{
		{
			name:           "Add success",
			endpoint:       "/add",
			request:        CalculationRequest{A: 5, B: 5},
			expectedStatus: http.StatusOK,
			expectedResult: intPtr(10),
		},
		{
			name:           "Subtract success",
			endpoint:       "/subtract",
			request:        CalculationRequest{A: 5, B: 3},
			expectedStatus: http.StatusOK,
			expectedResult: intPtr(2),
		},
		{
			name:           "Multiply success",
			endpoint:       "/multiply",
			request:        CalculationRequest{A: 5, B: 3},
			expectedStatus: http.StatusOK,
			expectedResult: intPtr(15),
		},
		{
			name:           "Divide success",
			endpoint:       "/divide",
			request:        CalculationRequest{A: 6, B: 3},
			expectedStatus: http.StatusOK,
			expectedResult: intPtr(2),
		},
		{
			name:           "Divide by zero",
			endpoint:       "/divide",
			request:        CalculationRequest{A: 5, B: 0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Division by zero is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Request Body
			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatal(err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create handler based on endpoint
			var handler http.HandlerFunc
			switch tt.endpoint {
			case "/add":
				handler = addHandler
			case "/subtract":
				handler = subtractHandler
			case "/multiply":
				handler = multiplyHandler
			case "/divide":
				handler = divideHandler
			}

			// Call handler
			handler(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)
			}

			// Parse response
			if tt.expectedResult != nil {
				var response CalculationResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Result != *tt.expectedResult {
					t.Errorf("handler returned unexpected response body: got %v want %v", response.Result, *tt.expectedResult)
				}
			}

			if tt.expectedError != "" {
				var response ErrorResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected response body: got %v want %v", response.Error, tt.expectedError)
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
