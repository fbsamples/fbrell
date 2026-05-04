package capisetup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validToken = "Bearer mock_token|test_app|write_capi_setup"

func TestBusinessContextsReturnsContexts(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"business_contexts", nil)
	req.Header.Set("Authorization", validToken)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var resp BusinessContextsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Contexts) != 2 {
		t.Fatalf("got %d contexts, want 2", len(resp.Contexts))
	}
	if resp.Contexts[0].ContextID != "123" {
		t.Fatalf("got context_id %q, want %q", resp.Contexts[0].ContextID, "123")
	}
	if resp.Contexts[1].ContextID != "456" {
		t.Fatalf("got context_id %q, want %q", resp.Contexts[1].ContextID, "456")
	}
}

func TestBusinessContextsRejectsPost(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("POST", Path+"business_contexts", nil)
	req.Header.Set("Authorization", validToken)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestBusinessContextsRejectsNoAuth(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"business_contexts", nil)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestConnectAndSharePixelSuccess(t *testing.T) {
	h := &Handler{}
	body := `{"context_id":"123","pixel_id":"12345","business_id":"67890","debug_id":"dbg_1"}`
	req := httptest.NewRequest("POST", Path+"connect_and_share_pixel", strings.NewReader(body))
	req.Header.Set("Authorization", validToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	var resp ConnectAndSharePixelResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatal("expected success=true")
	}
}

func TestConnectAndSharePixelUnknownContext(t *testing.T) {
	h := &Handler{}
	body := `{"context_id":"999","pixel_id":"12345","business_id":"67890"}`
	req := httptest.NewRequest("POST", Path+"connect_and_share_pixel", strings.NewReader(body))
	req.Header.Set("Authorization", validToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestConnectAndSharePixelMissingPixelID(t *testing.T) {
	h := &Handler{}
	body := `{"business_id":"67890"}`
	req := httptest.NewRequest("POST", Path+"connect_and_share_pixel", strings.NewReader(body))
	req.Header.Set("Authorization", validToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestConnectAndSharePixelMissingBusinessID(t *testing.T) {
	h := &Handler{}
	body := `{"pixel_id":"12345"}`
	req := httptest.NewRequest("POST", Path+"connect_and_share_pixel", strings.NewReader(body))
	req.Header.Set("Authorization", validToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestConnectAndSharePixelRejectsGet(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"connect_and_share_pixel", nil)
	req.Header.Set("Authorization", validToken)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestConnectAndSharePixelRejectsNoAuth(t *testing.T) {
	h := &Handler{}
	body := `{"pixel_id":"12345","business_id":"67890"}`
	req := httptest.NewRequest("POST", Path+"connect_and_share_pixel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestConnectAndSharePixelInvalidJSON(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("POST", Path+"connect_and_share_pixel", strings.NewReader("not json"))
	req.Header.Set("Authorization", validToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRejectsMissingScope(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"business_contexts", nil)
	req.Header.Set("Authorization", "Bearer mock_token|test_app|read,write")
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestUnknownEndpoint(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("GET", Path+"unknown", nil)
	req.Header.Set("Authorization", validToken)
	w := httptest.NewRecorder()

	if err := h.Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}
