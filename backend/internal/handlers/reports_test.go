package handlers

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestReportDateRangeUsesLaPazInclusiveDates(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest("GET", "/reports/operations?date_from=2026-08-01&date_to=2026-08-20", nil)
	response := httptest.NewRecorder()
	reportRange, ok := reportDateRange(response, request, time.Date(2026, 8, 20, 18, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if reportRange.DateFrom != "2026-08-01" || reportRange.DateTo != "2026-08-20" || reportRange.Timezone != "America/La_Paz" {
		t.Fatalf("range=%+v", reportRange)
	}
	if reportRange.FromUTC.Format(time.RFC3339) != "2026-08-01T04:00:00Z" || reportRange.ToUTC.Format(time.RFC3339) != "2026-08-21T04:00:00Z" {
		t.Fatalf("utc range=%s..%s", reportRange.FromUTC, reportRange.ToUTC)
	}
}

func TestReportDateRangeRejectsInvalidAndOversizedRanges(t *testing.T) {
	t.Parallel()
	for _, target := range []string{
		"/reports/operations?date_from=2026-08-21&date_to=2026-08-20",
		"/reports/operations?date_from=2024-01-01&date_to=2026-01-02",
		"/reports/operations?date_from=20-08-2026&date_to=2026-08-20",
	} {
		request := httptest.NewRequest("GET", target, nil)
		response := httptest.NewRecorder()
		if _, ok := reportDateRange(response, request, time.Now()); ok || response.Code != 422 {
			t.Fatalf("target=%s status=%d ok=%v", target, response.Code, ok)
		}
	}
}
