package testutils

import (
	"bytes"
	"centralized-wallet/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func ExecuteRequest(router *gin.Engine, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		reqBody, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	} else {
		// Pass nil body if no request body is needed
		req, _ = http.NewRequest(method, url, nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func AssertAPISuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedMessage string, data interface{}, statusCode ...int) {
	// Set the default status code to http.StatusOK if none is provided
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	// Marshal the `data` map into a JSON string
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	// Create the expected JSON response string
	expectedJSON := fmt.Sprintf(`{"status":"success","message":"%s","data":%s}`, expectedMessage, dataJSON)

	// Compare the actual and expected response
	assert.Equal(t, code, w.Code)
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

func AssertAPIErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedError *utils.AppError) {
	assert.Equal(t, expectedError.Code, w.Code)
	expectedJSON := fmt.Sprintf(`{"status":"error","message":"%s"}`, expectedError.Message)
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

type BaseHandlerTestCase struct {
	Name                  string
	Body                  map[string]interface{}
	URL                   string
	Method                string
	TestType              string // "success" or "error"
	MockSetup             func()
	MockAssert            func(t *testing.T)
	ExpectedEntity        interface{}
	ExpectedStatus        int
	ExpectedResponseError *utils.AppError
	ExpectedError         error
	ExpectedMessage       string
}

type TestHandlerRequest struct {
	Method string
	URL    string
}
