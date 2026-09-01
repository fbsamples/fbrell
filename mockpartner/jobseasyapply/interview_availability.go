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

package jobseasyapply

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/fbsamples/fbrell/mockpartner"
)

// The interview endpoints speak the Booking Server wire vocabulary, where a
// slot belongs to the job posting it books an interview for. Nothing tells two
// slots at the same start time apart, so a posting cannot offer overlapping
// slots and external_job_id plus start_timestamp keys one.

type SlotTime struct {
	ExternalJobID  string `json:"external_job_id"`
	StartTimestamp int64  `json:"start_timestamp"`
	DurationSec    int64  `json:"duration_sec,omitempty"`
}

// Mock inventory shape. Generated slots are deterministic for a given clock.
const (
	secondsPerHour     = 3600
	firstSlotHourUTC   = 9
	lastSlotEndHourUTC = 17
	slotDurationSec    = secondsPerHour
	slotStrideSec      = secondsPerHour
	slotsPerDay        = (lastSlotEndHourUTC - firstSlotHourUTC) * secondsPerHour / slotStrideSec
	inventoryDays      = 45
	unavailableEvery   = 4
	omitDurationEvery  = 2
)

type inventoryMode int

const (
	inventoryNormal inventoryMode = iota
	inventoryAllUnavailable
	inventoryEmpty
)

type slot struct {
	slotTime  SlotTime
	available bool
}

// buildInventory generates a job posting's slots in ascending start order,
// filling each business day of the inventoryDays window that opens the day
// after the handler's clock.
func (h *Handler) buildInventory(externalJobID string, mode inventoryMode) []slot {
	if mode == inventoryEmpty {
		return nil
	}

	now := h.now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), firstSlotHourUTC, 0, 0, 0, time.UTC).AddDate(0, 0, 1)

	slots := make([]slot, 0, inventoryDays*slotsPerDay)
	for range inventoryDays {
		if !isWeekend(day) {
			for i := range slotsPerDay {
				// The taken and no-duration patterns run across the whole
				// inventory rather than restarting each day.
				n := len(slots)
				slots = append(slots, slot{
					slotTime: SlotTime{
						ExternalJobID:  externalJobID,
						StartTimestamp: day.Add(time.Duration(i) * slotStrideSec * time.Second).Unix(),
						DurationSec:    durationSecFor(n),
					},
					available: mode == inventoryNormal && n%unavailableEvery != unavailableEvery-1,
				})
			}
		}
		day = day.AddDate(0, 0, 1)
	}
	return slots
}

// durationSecFor returns the nth generated slot's duration, zero meaning the
// slot goes out without one.
func durationSecFor(n int) int64 {
	if n%omitDurationEvery == omitDurationEvery-1 {
		return 0
	}
	return slotDurationSec
}

func isWeekend(t time.Time) bool {
	d := t.Weekday()
	return d == time.Saturday || d == time.Sunday
}

var errMissingExternalJobID = errors.New("jobseasyapply: external_job_id is required")

type AvailabilityLookupRequest struct {
	ExternalJobID       string     `json:"external_job_id"`
	SlotTime            []SlotTime `json:"slot_time,omitzero"`
	StartTimestampFrom  int64      `json:"start_timestamp_from,omitempty"`
	StartTimestampUntil int64      `json:"start_timestamp_until,omitempty"`
}

type AvailabilityLookupResponse struct {
	AvailableSlotTime []SlotTimeAvailability `json:"available_slot_time"`
}

type SlotTimeAvailability struct {
	SlotTime  SlotTime `json:"slot_time"`
	Available bool     `json:"available"`
}

// availabilityLookup answers from the scenario's mock inventory.
func (h *Handler) availabilityLookup(w http.ResponseWriter, r *http.Request, mode inventoryMode) error {
	if r.Method != http.MethodPost {
		return mockpartner.WriteError(w, http.StatusMethodNotAllowed, "invalid_request",
			"jobseasyapply: interview/availability-lookup requires POST")
	}

	var req AvailabilityLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request",
			"jobseasyapply: invalid JSON body")
	}
	if req.ExternalJobID == "" {
		return mockpartner.WriteError(w, http.StatusBadRequest, "invalid_request",
			errMissingExternalJobID.Error())
	}

	available := make([]SlotTimeAvailability, 0)
	for _, s := range h.buildInventory(req.ExternalJobID, mode) {
		if !matchesFilters(s.slotTime, &req) {
			continue
		}
		available = append(available, SlotTimeAvailability{
			SlotTime:  s.slotTime,
			Available: s.available,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(AvailabilityLookupResponse{AvailableSlotTime: available})
}

// matchesFilters reports whether a slot satisfies every filter set on the request.
func matchesFilters(s SlotTime, req *AvailabilityLookupRequest) bool {
	if req.StartTimestampFrom != 0 && s.StartTimestamp < req.StartTimestampFrom {
		return false
	}
	if req.StartTimestampUntil != 0 && s.StartTimestamp > req.StartTimestampUntil {
		return false
	}
	if len(req.SlotTime) == 0 {
		return true
	}
	for _, want := range req.SlotTime {
		if requestedSlotMatches(want, s) {
			return true
		}
	}
	return false
}

// requestedSlotMatches reports whether an inventory slot is the one a caller
// named in slot_time.
func requestedSlotMatches(want, have SlotTime) bool {
	if want.ExternalJobID != have.ExternalJobID || want.StartTimestamp != have.StartTimestamp {
		return false
	}
	return want.DurationSec == 0 || want.DurationSec == have.DurationSec
}
