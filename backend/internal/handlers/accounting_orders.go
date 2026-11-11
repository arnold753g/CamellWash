package handlers

import (
	"net/http"
	"time"

	"starter-backend/internal/respond"
)

func (h *ReportHandler) AccountingOrders(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	period, ok := reportDateRange(w, r, time.Now())
	if !ok {
		return
	}
	var invoice *bool
	if values, present := r.URL.Query()["invoice_requested"]; present {
		if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invoice filter must be true or false", nil)
			return
		}
		requested := values[0] == "true"
		invoice = &requested
	}
	page, limit := customerPage(r)
	report, err := h.reports.AccountingOrders(r.Context(), actor, period, invoice, page, limit)
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respond.JSON(w, http.StatusOK, report, "Accounting orders loaded")
}
