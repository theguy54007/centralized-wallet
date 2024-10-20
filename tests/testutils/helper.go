package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
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
