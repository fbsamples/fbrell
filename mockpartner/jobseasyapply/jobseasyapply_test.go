package jobseasyapply

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validToken = "Bearer mock_token|test_app|write_jobs_easy_apply"

// validBody is a minimal well-formed EXPORT_JOB_APPLICATION payload.
const validBody = `{
  "type": "EXPORT_JOB_APPLICATION",
  "appliedAt": 1750000000000,
  "externalJobId": "1983664855648430",
  "jobApplicant": "",
  "jobApplicationId": "",
  "questionResponses": {
    "contactInformationQuestionResponses": {
      "firstNameAnswer": {"value": "Jane"},
      "lastNameAnswer": {"value": "Doe"},
      "emailAnswer": {"value": "jane@example.com"}
    }
  }
}`

// bodyWithSlot is a well-formed application that also picked an interview time.
const bodyWithSlot = `{
  "type": "EXPORT_JOB_APPLICATION",
  "appliedAt": 1750000000000,
  "externalJobId": "job-1234",
  "jobApplicant": "",
  "jobApplicationId": "",
  "questionResponses": {
    "contactInformationQuestionResponses": {
      "firstNameAnswer": {"value": "Jane"},
      "lastNameAnswer": {"value": "Doe"},
      "emailAnswer": {"value": "jane@example.com"}
    }
  },
  "interview_slot": {"external_job_id": "job-1234", "start_timestamp": 1750000000, "duration_sec": 1800},
  "idempotency_token": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}`

// post sends a request through a clock-pinned handler and returns the recorded
// response.
func post(t *testing.T, path, body, auth string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	if err := testHandler().Handle(w, req); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestSubmitApplicationSuccess(t *testing.T) {
	w := post(t, Path+"submit_application", validBody, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	var resp SubmitApplicationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ApplicationID == "" {
		t.Fatal("expected a non-empty applicationId")
	}
	if resp.Booking != nil {
		t.Fatal("expected no booking for an application that picked no interview time")
	}
}

func TestSubmitApplicationWithInterviewSlot(t *testing.T) {
	w := post(t, Path+"submit_application", bodyWithSlot, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	var resp SubmitApplicationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Booking == nil {
		t.Fatal("expected a booking")
	}
	if resp.Booking.BookingID == "" {
		t.Error("expected a non-empty booking_id")
	}
	if resp.Booking.Status != mockBookingStatus {
		t.Errorf("got status %q, want %q", resp.Booking.Status, mockBookingStatus)
	}
}

func TestSubmitApplicationBookingError(t *testing.T) {
	w := post(t, Path+"submit_application/booking_error", bodyWithSlot, validToken)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	var resp SubmitApplicationErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("got %d errors, want 1", len(resp.Errors))
	}
	if resp.Errors[0].ErrorCode != mockBookingFailure {
		t.Errorf("got error code %q, want %q", resp.Errors[0].ErrorCode, mockBookingFailure)
	}
}

func TestSubmitApplicationClientError(t *testing.T) {
	w := post(t, Path+"submit_application/client_error", validBody, validToken)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestSubmitApplicationDeliveryError(t *testing.T) {
	w := post(t, Path+"submit_application/delivery_error", validBody, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	var resp DeliveryErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ApplicationDeliveryError == "" {
		t.Fatal("expected a non-empty applicationDeliveryError")
	}
}

func TestSubmitApplicationServerError(t *testing.T) {
	// server_error short-circuits before validation, so an empty body is fine.
	w := post(t, Path+"submit_application/server_error", "", validToken)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestSubmitApplicationRejectsGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, Path+"submit_application", nil)
	req.Header.Set("Authorization", validToken)
	w := httptest.NewRecorder()
	if err := testHandler().Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestSubmitApplicationRejectsNoAuth(t *testing.T) {
	w := post(t, Path+"submit_application", validBody, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestSubmitApplicationRejectsMissingScope(t *testing.T) {
	w := post(t, Path+"submit_application", validBody, "Bearer mock_token|test_app|read,write")
	if w.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestSubmitApplicationMissingRequiredField(t *testing.T) {
	// Omit the email answer.
	body := `{
      "type": "EXPORT_JOB_APPLICATION",
      "questionResponses": {
        "contactInformationQuestionResponses": {
          "firstNameAnswer": {"value": "Jane"},
          "lastNameAnswer": {"value": "Doe"}
        }
      }
    }`
	w := post(t, Path+"submit_application", body, validToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSubmitApplicationInvalidJSON(t *testing.T) {
	w := post(t, Path+"submit_application", "not json", validToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUnknownEndpoint(t *testing.T) {
	w := post(t, Path+"unknown", validBody, validToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}

// TestWirePayloadContract checks that the golden fixture matches the expected
// EXPORT_JOB_APPLICATION wire shape. DisallowUnknownFields proves the Go request
// struct accepts exactly that shape — any renamed or added wire field fails here.
func TestWirePayloadContract(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "export_job_application_golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var req SubmitApplicationRequest
	if err := dec.Decode(&req); err != nil {
		t.Fatalf("golden fixture does not match SubmitApplicationRequest: %v", err)
	}
	if req.Type != expectedType {
		t.Fatalf("got type %q, want %q", req.Type, expectedType)
	}
	if err := validate(&req); err != nil {
		t.Fatalf("golden fixture failed validation: %v", err)
	}
}
