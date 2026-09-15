/**
 * Copyright (c) 2014-present, Facebook, Inc. All rights reserved.
 *
 * You are hereby granted a non-exclusive, worldwide, royalty-free license to use,
 * copy, modify, and distribute this software in source code or binary form for use
 * in connection with the web services and APIs provided by Facebook.
 *
 * As with any software that integrates with the Facebook platform, your use of
 * this software is subject to the Facebook Developer Principles and Policies
 * [http://developers.facebook.com/policy/]. This copyright notice shall be
 * included in all copies or substantial portions of the software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package jobseasyapply implements mock partner API endpoints for a Jobs Easy
// Apply submit-to-partner flow. It provides a partner-shaped HTTP target that
// accepts a job-application submission and returns partner-style responses, and
// booking endpoints for scheduling the interview that can follow it.
//
// Each logical endpoint exposes scenario-selectable behavior chosen by sub-path
// (rather than a query parameter), so each outcome is reachable over plain HTTP,
// e.g. with curl or from Go tests:
//
//   - POST /mock-partner/jobs-easy-apply/submit_application
//     happy path — returns {"applicationId": "..."} at HTTP 200, carrying a
//     "booking" as well when the application picked an interview time.
//   - POST /mock-partner/jobs-easy-apply/submit_application/client_error
//     partner rejects a well-formed request — HTTP 422.
//   - POST /mock-partner/jobs-easy-apply/submit_application/delivery_error
//     partner accepts (HTTP 200) but reports a semantic delivery failure in the
//     body — {"applicationDeliveryError": "..."}.
//   - POST /mock-partner/jobs-easy-apply/submit_application/booking_error
//     partner cannot book the interview time the application picked, which
//     fails the submission as a whole — HTTP 422.
//   - POST /mock-partner/jobs-easy-apply/submit_application/server_error
//     partner-side failure — HTTP 500.
//   - POST /mock-partner/jobs-easy-apply/interview/availability-lookup
//     happy path — returns the job posting's interview slots and their
//     availability.
//   - POST /mock-partner/jobs-easy-apply/interview/availability-lookup/all_unavailable
//     every slot the posting offers is taken.
//   - POST /mock-partner/jobs-easy-apply/interview/availability-lookup/empty
//     the posting offers no slots at all.
//   - POST /mock-partner/jobs-easy-apply/interview/availability-lookup/server_error
//     partner-side failure — HTTP 500.
//
// Each scenario returns a plain partner body at the appropriate HTTP status,
// mirroring the capisetup sibling.
//
// All endpoints require a Bearer token issued by the mock OAuth provider that
// carries the write_jobs_easy_apply scope.
package jobseasyapply

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/fbsamples/fbrell/mockpartner"
)

const Path = "/mock-partner/jobs-easy-apply/"

// RequiredScope is the OAuth scope a Bearer token must carry to access any
// jobseasyapply endpoint. The mock OAuth provider issues tokens with the scopes
// requested at the authorize step, so callers must request
// scope=write_jobs_easy_apply.
const RequiredScope = "write_jobs_easy_apply"

// mockApplicationID is the partner-assigned application id returned on success.
const mockApplicationID = "mock-application-id"

// mockDeliveryError is the semantic-failure code returned by the delivery_error
// scenario (HTTP 200 body, not an HTTP error).
const mockDeliveryError = "expired_token"

// mockBookingID is the partner-assigned booking id returned for an application
// that picked an interview time, and mockBookingStatus the state it comes back
// in.
const (
	mockBookingID     = "mock-booking-id"
	mockBookingStatus = "CONFIRMED"
)

// mockBookingFailure is the error code returned by the booking_error scenario.
const mockBookingFailure = "BOOKING_FAILURE_SLOT_ALREADY_BOOKED_BY_USER"

var (
	errMissingFirstName  = errors.New("jobseasyapply: missing first_name_answer.value")
	errMissingLastName   = errors.New("jobseasyapply: missing last_name_answer.value")
	errMissingEmail      = errors.New("jobseasyapply: missing email_answer.value")
	errInsufficientScope = errors.New("jobseasyapply: token missing required scope " + RequiredScope)
)

// Handler serves mock Jobs Easy Apply partner API endpoints.
type Handler struct {
	// Now lets tests pin the generated interview inventory to a fixed clock.
	// When nil, time.Now is used.
	Now func() time.Time
}

func (h *Handler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

// Handle routes requests to the appropriate Jobs Easy Apply scenario endpoint.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) error {
	token, err := mockpartner.ParseBearerToken(r)
	if err != nil {
		return mockpartner.WriteError(w, http.StatusUnauthorized, "invalid_token", err.Error())
	}
	if !slices.Contains(token.Scopes, RequiredScope) {
		return mockpartner.WriteError(w, http.StatusForbidden, "insufficient_scope", errInsufficientScope.Error())
	}

	switch r.URL.Path {
	case Path + "submit_application":
		return h.submit(w, r, http.StatusOK, scenarioSuccess)
	case Path + "submit_application/client_error":
		return h.submit(w, r, http.StatusUnprocessableEntity, scenarioClientError)
	case Path + "submit_application/delivery_error":
		return h.submit(w, r, http.StatusOK, scenarioDeliveryError)
	case Path + "submit_application/booking_error":
		return h.submit(w, r, http.StatusUnprocessableEntity, scenarioBookingError)
	case Path + "submit_application/server_error":
		// A 5xx models a partner-side failure that can occur regardless of the
		// request body, so it short-circuits before decode/validation.
		return mockpartner.WriteError(w, http.StatusInternalServerError, "internal_error",
			"jobseasyapply: simulated partner server error")
	case Path + "interview/availability-lookup":
		return h.availabilityLookup(w, r, inventoryNormal)
	case Path + "interview/availability-lookup/all_unavailable":
		return h.availabilityLookup(w, r, inventoryAllUnavailable)
	case Path + "interview/availability-lookup/empty":
		return h.availabilityLookup(w, r, inventoryEmpty)
	case Path + "interview/availability-lookup/server_error":
		return mockpartner.WriteError(w, http.StatusInternalServerError, "internal_error",
			"jobseasyapply: simulated partner server error")
	default:
		return mockpartner.WriteError(w, http.StatusNotFound, "unknown_endpoint",
			fmt.Sprintf("No jobs-easy-apply endpoint at %s", r.URL.Path))
	}
}

type scenario int

const (
	scenarioSuccess scenario = iota
	scenarioClientError
	scenarioDeliveryError
	scenarioBookingError
)

// submit handles the POST scenarios that operate on a well-formed application:
// it enforces POST, decodes + validates the job-application payload, then emits
// the scenario-appropriate response.
func (h *Handler) submit(w http.ResponseWriter, r *http.Request, status int, s scenario) error {
	if r.Method != http.MethodPost {
		return mockpartner.WriteError(w, http.StatusMethodNotAllowed, "invalid_request",
			"jobseasyapply: submit_application requires POST")
	}

	var req SubmitApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request",
			"jobseasyapply: invalid JSON body")
	}
	if err := validate(&req); err != nil {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	}

	switch s {
	case scenarioClientError:
		return mockpartner.WriteError(w, status, "invalid_application",
			"jobseasyapply: simulated partner rejection of the application")
	case scenarioDeliveryError:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		return json.NewEncoder(w).Encode(DeliveryErrorResponse{ApplicationDeliveryError: mockDeliveryError})
	case scenarioBookingError:
		// An interview time that cannot be booked fails the whole submission, so
		// no application is recorded either.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		return json.NewEncoder(w).Encode(SubmitApplicationErrorResponse{
			Errors: []ApplicationDeliveryError{{ErrorCode: mockBookingFailure}},
		})
	default: // scenarioSuccess
		resp := SubmitApplicationResponse{ApplicationID: mockApplicationID}
		if req.InterviewSlot != nil {
			// The mock reserves nothing, so every application that picked an
			// interview time gets the same booking back — which is also what the
			// idempotency token asks for, since a redelivered application must
			// name the booking its first attempt made rather than a second one.
			resp.Booking = &Booking{BookingID: mockBookingID, Status: mockBookingStatus}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		return json.NewEncoder(w).Encode(resp)
	}
}

// validate enforces the required contact fields of the wire payload.
func validate(req *SubmitApplicationRequest) error {
	contact := req.QuestionResponses.ContactInformationQuestionResponses
	if contact.FirstNameAnswer.Value == "" {
		return errMissingFirstName
	}
	if contact.LastNameAnswer.Value == "" {
		return errMissingLastName
	}
	if contact.EmailAnswer.Value == "" {
		return errMissingEmail
	}
	return nil
}

// SubmitApplicationRequest is the expected wire payload for a job-application
// submission. Every field name is snake_case, matching the rest of the mock
// partner API and the client that calls it. The strict-decode test against
// testdata/export_job_application_golden.json asserts this shape.
type SubmitApplicationRequest struct {
	// AppliedAt is the submission time in epoch milliseconds — unlike every
	// other timestamp on these endpoints, which are epoch seconds.
	AppliedAt int64 `json:"applied_at"`
	// ExternalJobID is the partner-side id of the job listing applied to, and
	// JobApplicationID the client-side id of the same listing.
	ExternalJobID     string            `json:"external_job_id"`
	JobApplicationID  string            `json:"job_application_id"`
	QuestionResponses QuestionResponses `json:"question_responses"`
	// InterviewSlot is the interview time the candidate picked while applying,
	// one of the slots the availability lookup offered. Absent when the
	// application books no interview.
	InterviewSlot *SlotTime `json:"interview_slot,omitzero"`
	// IdempotencyToken identifies the submission across the retries a failed
	// delivery earns it, so a redelivery books one interview rather than one per
	// attempt.
	IdempotencyToken string `json:"idempotency_token,omitempty"`
}

// QuestionResponses groups the applicant's answers by section.
type QuestionResponses struct {
	ContactInformationQuestionResponses ContactInformationQuestionResponses `json:"contact_information_question_responses"`
	ResumeQuestionResponses             *ResumeQuestionResponses            `json:"resume_question_responses,omitzero"`
	AdditionalQuestionResponses         *AdditionalQuestionResponses        `json:"additional_question_responses,omitzero"`
}

// AnswerValue is the canonical {"value": "..."} wrapper used for scalar answers.
type AnswerValue struct {
	Value string `json:"value"`
}

// ContactInformationQuestionResponses holds the required contact answers plus an
// optional phone number.
type ContactInformationQuestionResponses struct {
	FirstNameAnswer               AnswerValue      `json:"first_name_answer"`
	LastNameAnswer                AnswerValue      `json:"last_name_answer"`
	EmailAnswer                   AnswerValue      `json:"email_answer"`
	CellphoneNumberQuestionAnswer *CellphoneAnswer `json:"cellphone_number_question_answer,omitzero"`
}

// CellphoneAnswer is the optional phone-number answer. Both halves are
// optional, so a caller that knows only the subscriber number sends an empty
// country code rather than omitting the answer.
type CellphoneAnswer struct {
	CountryCode    string `json:"country_code,omitempty"`
	NationalNumber string `json:"national_number,omitempty"`
}

// ResumeQuestionResponses wraps the optional resume answer.
type ResumeQuestionResponses struct {
	ResumeQuestionAnswer *ResumeAnswer `json:"resume_question_answer,omitzero"`
}

// ResumeAnswer is the optional resume media reference.
type ResumeAnswer struct {
	MediaURL string `json:"media_url,omitempty"`
	MediaURN string `json:"media_urn,omitempty"`
}

// AdditionalQuestionResponses carries partner-defined custom question answers.
type AdditionalQuestionResponses struct {
	CustomQuestionSetResponses []CustomQuestionSetResponse `json:"custom_question_set_responses"`
}

// CustomQuestionSetResponse groups one set of custom question responses.
type CustomQuestionSetResponse struct {
	CustomQuestionResponses []CustomQuestionResponse `json:"custom_question_responses"`
}

// CustomQuestionResponse is a single custom question answer.
type CustomQuestionResponse struct {
	QuestionIdentifier string       `json:"question_identifier"`
	Answer             CustomAnswer `json:"answer"`
}

// CustomAnswer wraps the typed value of a custom answer.
type CustomAnswer struct {
	TextAnswerValue AnswerValue `json:"text_answer_value"`
}

// SubmitApplicationResponse is the happy-path response body.
type SubmitApplicationResponse struct {
	ApplicationID string `json:"applicationId"`
	// Booking is the interview reserved for an application that picked a time,
	// omitted when it picked none.
	Booking *Booking `json:"booking,omitempty"`
}

// Booking is an interview reservation on the partner's side.
type Booking struct {
	BookingID string `json:"booking_id"`
	Status    string `json:"status"`
}

// DeliveryErrorResponse is the HTTP-200 semantic-failure body returned by the
// delivery_error scenario.
type DeliveryErrorResponse struct {
	ApplicationDeliveryError string `json:"applicationDeliveryError"`
}

// SubmitApplicationErrorResponse is the body of a submission the partner
// refused, returned at a non-200 status.
type SubmitApplicationErrorResponse struct {
	Errors []ApplicationDeliveryError `json:"errors"`
}

// ApplicationDeliveryError is one reason a submission was refused.
type ApplicationDeliveryError struct {
	ErrorCode string `json:"errorCode"`
}
