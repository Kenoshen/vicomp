package web

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mwingfield/vicomp/internal/models"
	"github.com/mwingfield/vicomp/internal/store"
)

// invoiceDetail is the fully-resolved view of an invoice used by both the HTML
// detail page and the PDF template.
type invoiceDetail struct {
	Invoice   models.Invoice
	Client    models.Client
	County    models.County
	Clinician models.Clinician
	Sessions  []models.Session
	Total     float64
	Settings  models.Settings
	From      fromParty
	Now       time.Time
}

// fromParty is the invoice "From" block: practice defaults from Settings with any
// non-empty clinician field taking precedence.
type fromParty struct {
	ClinicianName string
	PracticeName  string
	Address       string
	City          string
	State         string
	Zip           string
	Phone         string
	Email         string
	TaxID         string
	NPI           string
}

func mergeFrom(c models.Clinician, st models.Settings) fromParty {
	pick := func(override, fallback string) string {
		if strings.TrimSpace(override) != "" {
			return override
		}
		return fallback
	}
	f := fromParty{
		PracticeName: st.PracticeName,
		Address:      pick(c.Address, st.Address),
		City:         pick(c.City, st.City),
		State:        pick(c.State, st.State),
		Zip:          pick(c.Zip, st.Zip),
		Phone:        pick(c.Phone, st.Phone),
		Email:        pick(c.Email, st.Email),
		TaxID:        pick(c.TaxID, st.TaxID),
		NPI:          pick(c.NPI, st.NPI),
	}
	if c.ID != 0 {
		f.ClinicianName = c.FullName()
	}
	return f
}

func (s *Server) invoicesList(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListInvoices(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "invoices_list", pageData{Title: "Invoices", Data: list})
}

func (s *Server) invoiceNew(w http.ResponseWriter, r *http.Request) {
	clients, err := s.store.ListClients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "invoice_new", pageData{Title: "New Invoice", Data: clients})
}

// invoicePreview is an HTMX fragment: given ?client_id=, list the uninvoiced
// sessions that would be rolled into a new invoice, with a running total.
func (s *Server) invoicePreview(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(r.URL.Query().Get("client_id"))
	if err != nil {
		s.renderFragment(w, "invoice_new", "invoice_preview", nil)
		return
	}
	sessions, err := s.store.ListUninvoicedSessions(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var total float64
	for _, sess := range sessions {
		total += sess.Amount
	}
	s.renderFragment(w, "invoice_new", "invoice_preview", struct {
		ClientID int64
		Sessions []models.Session
		Total    float64
	}{id, sessions, total})
}

func (s *Server) invoiceCreate(w http.ResponseWriter, r *http.Request) {
	clientID := int64(formInt(r, "client_id"))
	if clientID == 0 {
		http.Error(w, "select a client", http.StatusBadRequest)
		return
	}
	invoiceID, err := s.store.CreateInvoiceFromSessions(r.Context(), clientID)
	if errors.Is(err, store.ErrNoSessions) {
		http.Error(w, "that client has no uninvoiced sessions", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/invoices/%d", invoiceID), http.StatusSeeOther)
}

func (s *Server) loadInvoiceDetail(r *http.Request, id int64) (invoiceDetail, error) {
	var d invoiceDetail
	inv, err := s.store.GetInvoice(r.Context(), id)
	if err != nil {
		return d, err
	}
	d.Invoice = inv
	d.Now = time.Now()

	client, err := s.store.GetClient(r.Context(), inv.ClientID)
	if err != nil {
		return d, err
	}
	d.Client = client
	if client.CountyID != nil {
		if c, err := s.store.GetCounty(r.Context(), *client.CountyID); err == nil {
			d.County = c
		}
	}
	if client.ClinicianID != nil {
		if c, err := s.store.GetClinician(r.Context(), *client.ClinicianID); err == nil {
			d.Clinician = c
		}
	}

	sessions, err := s.store.ListSessionsByInvoice(r.Context(), id)
	if err != nil {
		return d, err
	}
	d.Sessions = sessions
	for _, sess := range sessions {
		d.Total += sess.Amount
	}

	if st, err := s.store.GetSettings(r.Context()); err == nil {
		d.Settings = st
	}
	d.From = mergeFrom(d.Clinician, d.Settings)
	return d, nil
}

func (s *Server) invoiceShow(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	d, err := s.loadInvoiceDetail(r, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.render(w, "invoice_show", pageData{Title: "Invoice " + d.Invoice.InvoiceNumber, Data: d})
}

func (s *Server) invoicePDF(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	d, err := s.loadInvoiceDetail(r, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	tmpl, ok := s.templates["invoice_pdf"]
	if !ok {
		http.Error(w, "missing invoice_pdf template", http.StatusInternalServerError)
		return
	}
	var html bytes.Buffer
	if err := tmpl.ExecuteTemplate(&html, "invoice_pdf", d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfBytes, err := s.pdf.RenderHTML(r.Context(), html.Bytes())
	if err != nil {
		http.Error(w, "pdf render failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.pdf"`, d.Invoice.InvoiceNumber))
	w.Write(pdfBytes)
}

func (s *Server) invoiceStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	status := formStr(r, "status")
	switch status {
	case "draft", "sent", "paid":
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateInvoiceStatus(r.Context(), id, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/invoices/%d", id), http.StatusSeeOther)
}

func (s *Server) invoiceDelete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteInvoice(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
