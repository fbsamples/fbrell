package jobseasyapply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fixedNow is a Monday at noon UTC, so the first inventory day is the Tuesday
// after it and the window ends on Thursday 2026-07-30.
var fixedNow = time.Date(2026, time.June, 15, 12, 0, 0, 0, time.UTC)

var firstSlotStart = time.Date(2026, time.June, 16, firstSlotHourUTC, 0, 0, 0, time.UTC).Unix()

var lastSlotStart = time.Date(2026, time.July, 30, lastSlotEndHourUTC-1, 0, 0, 0, time.UTC).Unix()

const (
	// wantWeekdays is how many business days fall in that window.
	wantWeekdays  = 33
	wantSlotCount = wantWeekdays * slotsPerDay
)

func testHandler() *Handler {
	return &Handler{Now: func() time.Time { return fixedNow }}
}

func decodeLookup(t *testing.T, body []byte) AvailabilityLookupResponse {
	t.Helper()
	var resp AvailabilityLookupResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestInterviewAvailabilityLookupSuccess(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != wantSlotCount {
		t.Fatalf("got %d slots, want %d", len(resp.SlotTimeAvailability), wantSlotCount)
	}
	first := resp.SlotTimeAvailability[0]
	if first.SlotTime.StartTimestamp != firstSlotStart {
		t.Fatalf("got first start_timestamp %d, want %d", first.SlotTime.StartTimestamp, firstSlotStart)
	}
	if first.SlotTime.ExternalJobID != "job-1234" {
		t.Fatalf("got external_job_id %q, want %q", first.SlotTime.ExternalJobID, "job-1234")
	}
	if first.SlotTime.DurationSec != slotDurationSec {
		t.Fatalf("got duration_sec %d, want %d", first.SlotTime.DurationSec, slotDurationSec)
	}
	last := resp.SlotTimeAvailability[len(resp.SlotTimeAvailability)-1]
	if last.SlotTime.StartTimestamp != lastSlotStart {
		t.Fatalf("got last start_timestamp %d, want %d", last.SlotTime.StartTimestamp, lastSlotStart)
	}
}

func TestInterviewAvailabilityLookupSpansBusinessDays(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`, validToken)
	resp := decodeLookup(t, w.Body.Bytes())

	perDay := make(map[string]int)
	var prev int64
	for _, s := range resp.SlotTimeAvailability {
		start := time.Unix(s.SlotTime.StartTimestamp, 0).UTC()
		if isWeekend(start) {
			t.Fatalf("slot at %s falls on a %s, want business days only", start, start.Weekday())
		}
		if start.Hour() < firstSlotHourUTC || start.Hour() >= lastSlotEndHourUTC {
			t.Fatalf("slot at %s starts outside %02d:00-%02d:00 UTC",
				start, firstSlotHourUTC, lastSlotEndHourUTC)
		}
		if start.Minute() != 0 || start.Second() != 0 {
			t.Fatalf("slot at %s does not start on the hour", start)
		}
		if s.SlotTime.StartTimestamp <= prev {
			t.Fatalf("slot at %s is not after the previous slot", start)
		}
		prev = s.SlotTime.StartTimestamp
		perDay[start.Format(time.DateOnly)]++
	}

	if len(perDay) != wantWeekdays {
		t.Fatalf("got slots on %d days, want %d", len(perDay), wantWeekdays)
	}
	for day, n := range perDay {
		if n != slotsPerDay {
			t.Errorf("day %s has %d slots, want %d", day, n, slotsPerDay)
		}
	}
}

// TestInterviewAvailabilityLookupBusinessDayHours spells the hours out rather
// than deriving them from the constants the generator uses.
func TestInterviewAvailabilityLookupBusinessDayHours(t *testing.T) {
	endOfFirstDay := time.Date(2026, time.June, 16, 23, 59, 59, 0, time.UTC).Unix()
	body := fmt.Sprintf(`{"external_job_id":"job-1234","start_timestamp_until":%d}`, endOfFirstDay)
	w := post(t, Path+"interview/availability-lookup", body, validToken)

	var want []int64
	for hour := 9; hour < 17; hour++ {
		want = append(want, time.Date(2026, time.June, 16, hour, 0, 0, 0, time.UTC).Unix())
	}

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != len(want) {
		t.Fatalf("got %d slots on the first day, want %d", len(resp.SlotTimeAvailability), len(want))
	}
	for i, s := range resp.SlotTimeAvailability {
		if s.SlotTime.StartTimestamp != want[i] {
			t.Errorf("slot %d starts at %s, want %s", i,
				time.Unix(s.SlotTime.StartTimestamp, 0).UTC(), time.Unix(want[i], 0).UTC())
		}
	}
}

// TestInterviewAvailabilityLookupOmitsSomeDurations decodes into a raw map to
// tell an omitted duration_sec apart from one sent as zero.
func TestInterviewAvailabilityLookupOmitsSomeDurations(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`, validToken)

	var body struct {
		SlotTimeAvailability []struct {
			SlotTime map[string]json.RawMessage `json:"slot_time"`
		} `json:"slot_time_availability"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.SlotTimeAvailability) != wantSlotCount {
		t.Fatalf("got %d slots, want %d", len(body.SlotTimeAvailability), wantSlotCount)
	}
	for i, s := range body.SlotTimeAvailability {
		_, got := s.SlotTime["duration_sec"]
		want := i%omitDurationEvery != omitDurationEvery-1
		if got != want {
			t.Errorf("slot %d: duration_sec present = %v, want %v", i, got, want)
		}
	}
}

// TestInterviewAvailabilityLookupResponseShape decodes into a map because a
// decode into AvailabilityLookupResponse would silently ignore an extra field.
func TestInterviewAvailabilityLookupResponseShape(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`, validToken)

	var body map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for field := range body {
		if field != "slot_time_availability" {
			t.Errorf("unexpected response field %q; the spec defines only slot_time_availability", field)
		}
	}
	if _, ok := body["slot_time_availability"]; !ok {
		t.Error("response is missing slot_time_availability")
	}
}

func TestInterviewAvailabilityLookupReportsUnavailableSlots(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`, validToken)
	resp := decodeLookup(t, w.Body.Bytes())

	var available, unavailable int
	for _, s := range resp.SlotTimeAvailability {
		if s.Available {
			available++
		} else {
			unavailable++
		}
	}
	if available == 0 || unavailable == 0 {
		t.Fatalf("got %d available and %d unavailable, want a mix of both", available, unavailable)
	}
}

func TestInterviewAvailabilityLookupAllUnavailable(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup/all_unavailable",
		`{"external_job_id":"job-1234"}`, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != wantSlotCount {
		t.Fatalf("got %d slots, want %d", len(resp.SlotTimeAvailability), wantSlotCount)
	}
	for _, s := range resp.SlotTimeAvailability {
		if s.Available {
			t.Fatalf("slot at %d is available, want every slot taken", s.SlotTime.StartTimestamp)
		}
	}
}

func TestInterviewAvailabilityLookupEmpty(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup/empty", `{"external_job_id":"job-1234"}`, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != 0 {
		t.Fatalf("got %d slots, want none", len(resp.SlotTimeAvailability))
	}
	// An empty list must serialize as [] rather than null so clients can iterate
	// it without a nil check.
	if !bytes.Contains(w.Body.Bytes(), []byte(`"slot_time_availability":[]`)) {
		t.Fatalf("expected an empty JSON array, got %s", w.Body.String())
	}
}

func TestInterviewAvailabilityLookupFiltersByStartTimestampFrom(t *testing.T) {
	// Skip the first two slots.
	from := firstSlotStart + 2*slotStrideSec
	body := fmt.Sprintf(`{"external_job_id":"job-1234","start_timestamp_from":%d}`, from)
	w := post(t, Path+"interview/availability-lookup", body, validToken)

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != wantSlotCount-2 {
		t.Fatalf("got %d slots, want %d", len(resp.SlotTimeAvailability), wantSlotCount-2)
	}
	for _, s := range resp.SlotTimeAvailability {
		if s.SlotTime.StartTimestamp < from {
			t.Fatalf("slot at %d starts before the requested lower bound %d", s.SlotTime.StartTimestamp, from)
		}
	}
}

func TestInterviewAvailabilityLookupFiltersToAWindow(t *testing.T) {
	// Friday 2026-06-19 through Monday 2026-06-22, inclusive of both, so the
	// weekend between them must contribute nothing.
	from := time.Date(2026, time.June, 19, firstSlotHourUTC, 0, 0, 0, time.UTC)
	until := time.Date(2026, time.June, 22, lastSlotEndHourUTC, 0, 0, 0, time.UTC)
	body := fmt.Sprintf(`{"external_job_id":"job-1234","start_timestamp_from":%d,"start_timestamp_until":%d}`,
		from.Unix(), until.Unix())
	w := post(t, Path+"interview/availability-lookup", body, validToken)

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != 2*slotsPerDay {
		t.Fatalf("got %d slots, want %d", len(resp.SlotTimeAvailability), 2*slotsPerDay)
	}
	for _, s := range resp.SlotTimeAvailability {
		start := time.Unix(s.SlotTime.StartTimestamp, 0).UTC()
		if start.Before(from) || start.After(until) {
			t.Errorf("slot at %s falls outside the requested window", start)
		}
	}
}

func TestInterviewAvailabilityLookupFiltersByStartTimestampUntil(t *testing.T) {
	// Keep only the first three slots.
	until := firstSlotStart + 2*slotStrideSec
	body := fmt.Sprintf(`{"external_job_id":"job-1234","start_timestamp_until":%d}`, until)
	w := post(t, Path+"interview/availability-lookup", body, validToken)

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != 3 {
		t.Fatalf("got %d slots, want 3", len(resp.SlotTimeAvailability))
	}
	for _, s := range resp.SlotTimeAvailability {
		if s.SlotTime.StartTimestamp > until {
			t.Fatalf("slot at %d starts after the requested upper bound %d", s.SlotTime.StartTimestamp, until)
		}
	}
}

func TestInterviewAvailabilityLookupFiltersBySlotTime(t *testing.T) {
	body := fmt.Sprintf(
		`{"external_job_id":"job-1234","slot_time":[{"external_job_id":"job-1234","start_timestamp":%d}]}`,
		firstSlotStart,
	)
	w := post(t, Path+"interview/availability-lookup", body, validToken)

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != 1 {
		t.Fatalf("got %d slots, want 1", len(resp.SlotTimeAvailability))
	}
	if resp.SlotTimeAvailability[0].SlotTime.StartTimestamp != firstSlotStart {
		t.Fatalf("got start_timestamp %d, want %d", resp.SlotTimeAvailability[0].SlotTime.StartTimestamp, firstSlotStart)
	}
}

// TestInterviewAvailabilityLookupOmitsUnknownSlotTime pins the spec's rule that
// an unrecognized requested slot is dropped rather than rejected.
func TestInterviewAvailabilityLookupOmitsUnknownSlotTime(t *testing.T) {
	body := `{"external_job_id":"job-1234","slot_time":[{"external_job_id":"job-1234","start_timestamp":1}]}`
	w := post(t, Path+"interview/availability-lookup", body, validToken)
	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != 0 {
		t.Fatalf("got %d slots, want none", len(resp.SlotTimeAvailability))
	}
}

func TestInterviewAvailabilityLookupFiltersByDuration(t *testing.T) {
	body := fmt.Sprintf(
		`{"external_job_id":"job-1234","slot_time":[{"external_job_id":"job-1234","start_timestamp":%d,`+
			`"duration_sec":%d}]}`,
		firstSlotStart, slotDurationSec*2,
	)
	w := post(t, Path+"interview/availability-lookup", body, validToken)

	resp := decodeLookup(t, w.Body.Bytes())
	if len(resp.SlotTimeAvailability) != 0 {
		t.Fatalf("got %d slots, want none", len(resp.SlotTimeAvailability))
	}
}

func TestInterviewAvailabilityLookupRequiresExternalJobID(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{}`, validToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestInterviewAvailabilityLookupRejectsGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, Path+"interview/availability-lookup", nil)
	req.Header.Set("Authorization", validToken)
	w := httptest.NewRecorder()
	if err := testHandler().Handle(w, req); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestInterviewAvailabilityLookupInvalidJSON(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", "not json", validToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestInterviewAvailabilityLookupServerError(t *testing.T) {
	// server_error short-circuits before validation, so an empty body is fine.
	w := post(t, Path+"interview/availability-lookup/server_error", "", validToken)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestInterviewAvailabilityLookupRejectsNoAuth(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestInterviewAvailabilityLookupRejectsMissingScope(t *testing.T) {
	w := post(t, Path+"interview/availability-lookup", `{"external_job_id":"job-1234"}`,
		"Bearer mock_token|test_app|read,write")
	if w.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandlerDefaultsToTimeNow(t *testing.T) {
	// A zero-value Handler is what main wires up, so it must generate inventory
	// without a clock injected.
	slots := (&Handler{}).buildInventory("job-1234", inventoryNormal)
	if len(slots) == 0 {
		t.Fatal("expected a zero-value Handler to generate inventory")
	}
	if slots[0].slotTime.StartTimestamp <= time.Now().Unix() {
		t.Fatal("expected generated slots to be in the future")
	}
}

func TestInterviewAvailabilityLookupRequestContract(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "interview_availability_lookup_request_golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var req AvailabilityLookupRequest
	if err := dec.Decode(&req); err != nil {
		t.Fatalf("golden fixture does not match AvailabilityLookupRequest: %v", err)
	}
	if req.ExternalJobID == "" {
		t.Fatal("golden fixture is missing external_job_id")
	}
}
