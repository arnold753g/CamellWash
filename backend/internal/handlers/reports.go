package handlers

import (
	"net/http"
	"strings"
	"time"

	"starter-backend/internal/respond"
	reportservice "starter-backend/internal/services"
)

type ReportHandler struct{ reports *reportservice.ReportService }

func NewReportHandler(reports *reportservice.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

func (h *ReportHandler) Operational(w http.ResponseWriter, r *http.Request) {
	reportRange, ok := reportDateRange(w, r, time.Now())
	if !ok {
		return
	}
	report, err := h.reports.Operational(r.Context(), reportRange)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Operational report could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, report, "Operational report loaded")
}

func (h *ReportHandler) Financial(w http.ResponseWriter, r *http.Request) {
	reportRange, ok := reportDateRange(w, r, time.Now())
	if !ok {
		return
	}
	report, err := h.reports.Financial(r.Context(), reportRange)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Financial report could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, report, "Financial report loaded")
}

func (h *ReportHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	reportRange, ok := reportDateRange(w, r, time.Now())
	if !ok {
		return
	}
	report, err := h.reports.Dashboard(r.Context(), reportRange)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Dashboard could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, report, "Dashboard loaded")
}

func reportDateRange(w http.ResponseWriter, r *http.Request, now time.Time) (reportservice.ReportRange, bool) {
	location, err := time.LoadLocation(reportservice.ReportTimezone)
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Operational timezone is unavailable", nil)
		return reportservice.ReportRange{}, false
	}
	today := now.In(location)
	defaultFrom := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, location)
	defaultTo := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	from, err := parseReportDate(r.URL.Query().Get("date_from"), defaultFrom, location)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Date must use YYYY-MM-DD", map[string]string{"date_from": "Invalid date"})
		return reportservice.ReportRange{}, false
	}
	to, err := parseReportDate(r.URL.Query().Get("date_to"), defaultTo, location)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Date must use YYYY-MM-DD", map[string]string{"date_to": "Invalid date"})
		return reportservice.ReportRange{}, false
	}
	if to.Before(from) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Start date cannot be after end date", map[string]string{"date_to": "Invalid range"})
		return reportservice.ReportRange{}, false
	}
	if to.After(from.AddDate(1, 0, 0)) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Report range cannot exceed one year", map[string]string{"date_to": "Maximum range is one year"})
		return reportservice.ReportRange{}, false
	}
	toExclusive := to.AddDate(0, 0, 1)
	return reportservice.ReportRange{
		FromUTC: from.UTC(), ToUTC: toExclusive.UTC(),
		DateFrom: from.Format("2006-01-02"), DateTo: to.Format("2006-01-02"), Timezone: reportservice.ReportTimezone,
	}, true
}

func parseReportDate(raw string, fallback time.Time, location *time.Location) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback, nil
	}
	return time.ParseInLocation("2006-01-02", value, location)
}
