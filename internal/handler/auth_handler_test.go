package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {

	tests := []struct {
		name        string
		method      string
		path        string
		jsonData    string
		status      int
		correctJSON bool
	}{
		{name: "неправильный метод", method: "GET", path: "api/register", jsonData: "", status: http.StatusMethodNotAllowed, correctJSON: false},
		{name: "правильный медот и корерктные данные", method: "POST", path: "api/register", jsonData: `{"email" :"masha.shishkina@newmail.ru", "password" :"L@e543^5"}`, status: http.StatusCreated, correctJSON: true},
		{name: "правильный медот и некорерктный JSON", method: "POST", path: "api/register", jsonData: `{"example":2:]}}`, status: http.StatusBadRequest, correctJSON: false},
		{name: "правильный медот и некорерктный URL", method: "POST", path: "api/shorten", jsonData: `{"url":"/com/not/too/long/path"}`, status: http.StatusBadRequest, correctJSON: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/"+tt.path, strings.NewReader(tt.jsonData))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			// Register(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.status {
				t.Errorf("Ожидался статус %d, получено %d", tt.status, res.StatusCode)
			}
			contentType := res.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Ожидался Content-Type application/json, получено %s", contentType)
			}
			body, _ := io.ReadAll(res.Body)
			//проверяем корректность JSON
			if !json.Valid([]byte(body)) {
				t.Errorf("Ожидалось body - JSON, получено - некорректный JSON %q", string(body))
			} else {
				if tt.method == "POST" {
					var jsonContent CreatedMsg
					//проверяем что в теле пришел правильный JSON
					if err := json.Unmarshal(body, &jsonContent); err != nil && !tt.correctJSON {
						t.Errorf("Ожидалось correct_json = %t, получено - correct_JSON  = %t", tt.correctJSON, err != nil)
					}
				}
			}
		})

	}
}
