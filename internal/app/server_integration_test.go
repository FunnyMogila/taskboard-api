//go:build integration

package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"taskboard-api/internal/app"
	"taskboard-api/internal/config"
	"taskboard-api/internal/domain"
)

type apiResponse[T any] struct {
	Data  T      `json:"data"`
	Error string `json:"error"`
}

func TestCreateAndGetUserIntegration(t *testing.T) {
	cfg := config.Load()

	server := app.NewServer(cfg)
	defer server.Close()

	testServer := httptest.NewServer(server.Router())
	defer testServer.Close()

	email := fmt.Sprintf("integration_%d@test.com", time.Now().UnixNano())

	createBody := []byte(fmt.Sprintf(`{
		"name": "Integration User",
		"email": "%s"
	}`, email))

	createResp, err := http.Post(
		testServer.URL+"/api/v1/users",
		"application/json",
		bytes.NewReader(createBody),
	)
	if err != nil {
		t.Fatalf("create user request failed: %v", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", createResp.StatusCode)
	}

	var created apiResponse[domain.User]
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if created.Data.ID == 0 {
		t.Fatalf("expected created user ID")
	}

	getResp, err := http.Get(
		fmt.Sprintf("%s/api/v1/users/%d", testServer.URL, created.Data.ID),
	)
	if err != nil {
		t.Fatalf("get user request failed: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResp.StatusCode)
	}

	var got apiResponse[domain.User]
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("decode get response: %v", err)
	}

	if got.Data.Email != email {
		t.Fatalf("expected email %s, got %s", email, got.Data.Email)
	}
}
