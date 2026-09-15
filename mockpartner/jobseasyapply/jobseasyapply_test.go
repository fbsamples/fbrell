package jobseasyapply

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const validToken = "Bearer mock_token|test_app|write_jobs_easy_apply"

// validBody is a minimal well-formed application payload: the required fields
// and nothing else.
const validBody = `{
  "applied_at": 1750000000000,
  "external_job_id": "1983664855648430",
  "job_application_id": "1290452048837161",
  "question_responses": {
    "contact_information_question_responses": {
      "first_name_answer": {"value": "Jane"},
      "last_name_answer": {"value": "Doe"},
      "email_answer": {"value": "jane@example.com"}
    }
  }
}`

// bodyWithSlot is a well-formed application that also picked an interview time.
const bodyWithSlot = `{
  "applied_at": 1750000000000,
  "external_job_id": "job-1234",
  "job_application_id": "1290452048837161",
  "question_responses": {
    "contact_information_question_responses": {
      "first_name_answer": {"value": "Jane"},
      "last_name_answer": {"value": "Doe"},
      "email_answer": {"value": "jane@example.com"}
    }
  },
  "interview_slot": {"start_timestamp": 1750000000, "duration_sec": 1800},
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
      "question_responses": {
        "contact_information_question_responses": {
          "first_name_answer": {"value": "Jane"},
          "last_name_answer": {"value": "Doe"}
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

// TestWirePayloadContract pins the submit wire shape against the golden fixture,
// which exercises every field the payload has — required and optional.
//
// The two checks are complementary: DisallowUnknownFields proves no golden key
// is missing from SubmitApplicationRequest, and comparing against wantGolden
// proves no struct field is missing from the golden or bound to the wrong key.
// A field silently decoding to its zero value therefore fails here rather than
// at the far end of a live call.
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
	if !reflect.DeepEqual(req, wantGolden) {
		t.Fatalf("golden fixture decoded to\n%s\nwant\n%s", toJSON(t, req), toJSON(t, wantGolden))
	}
	if err := validate(&req); err != nil {
		t.Fatalf("golden fixture failed validation: %v", err)
	}
}

// wantGolden is testdata/export_job_application_golden.json decoded field by
// field. Kept next to the test that uses it so the two stay in step.
var wantGolden = SubmitApplicationRequest{
	AppliedAt:        1750000000000,
	ExternalJobID:    "1983664855648430",
	JobApplicationID: "1290452048837161",
	QuestionResponses: QuestionResponses{
		ContactInformationQuestionResponses: ContactInformationQuestionResponses{
			FirstNameAnswer:               AnswerValue{Value: "Jane"},
			LastNameAnswer:                AnswerValue{Value: "Doe"},
			EmailAnswer:                   AnswerValue{Value: "jane@example.com"},
			CellphoneNumberQuestionAnswer: &CellphoneAnswer{NationalNumber: "5551234567"},
		},
		ResumeQuestionResponses: &ResumeQuestionResponses{
			ResumeQuestionAnswer: &ResumeAnswer{MediaURL: "https://example.com/resume.pdf"},
		},
		AdditionalQuestionResponses: &AdditionalQuestionResponses{
			CustomQuestionSetResponses: []CustomQuestionSetResponse{{
				CustomQuestionResponses: []CustomQuestionResponse{{
					QuestionIdentifier: "q1",
					Answer:             CustomAnswer{TextAnswerValue: AnswerValue{Value: "answer1"}},
				}},
			}},
		},
	},
	InterviewSlot:    &SlotTime{StartTimestamp: 1750000000, DurationSec: 1800},
	IdempotencyToken: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
}

// toJSON renders a payload for a failure message, where %+v would print
// pointer addresses for the nested optional sections.
func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
