package web

import (
	"net/http"

	"github.com/mwingfield/vicomp/internal/models"
)

// clientFormData carries the client plus the option lists for its foreign keys.
type clientFormData struct {
	Client     models.Client
	Counties   []models.County
	Clinicians []models.Clinician
}

func (s *Server) clientsList(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListClients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "clients_list", pageData{Title: "Clients", Data: list})
}

func (s *Server) clientFormPage(w http.ResponseWriter, r *http.Request, title string, c models.Client, errMsg string) {
	counties, err := s.store.ListCounties(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	clinicians, err := s.store.ListClinicians(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "client_form", pageData{
		Title: title,
		Error: errMsg,
		Data:  clientFormData{Client: c, Counties: counties, Clinicians: clinicians},
	})
}

func (s *Server) clientNew(w http.ResponseWriter, r *http.Request) {
	s.clientFormPage(w, r, "New Client", models.Client{}, "")
}

func (s *Server) clientEdit(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c, err := s.store.GetClient(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.clientFormPage(w, r, "Edit Client", c, "")
}

func (s *Server) clientCreate(w http.ResponseWriter, r *http.Request) {
	c := clientFromForm(r)
	if _, err := s.store.CreateClient(r.Context(), c); err != nil {
		s.clientFormPage(w, r, "New Client", c, err.Error())
		return
	}
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

func (s *Server) clientUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c := clientFromForm(r)
	c.ID = id
	if err := s.store.UpdateClient(r.Context(), c); err != nil {
		s.clientFormPage(w, r, "Edit Client", c, err.Error())
		return
	}
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

func (s *Server) clientDelete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteClient(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func clientFromForm(r *http.Request) models.Client {
	return models.Client{
		FirstName:   formStr(r, "first_name"),
		LastName:    formStr(r, "last_name"),
		DateOfBirth: formDatePtr(r, "date_of_birth"),
		CountyID:    formID64Ptr(r, "county_id"),
		ClinicianID: formID64Ptr(r, "clinician_id"),
		ClaimNumber: formStr(r, "claim_number"),
		Notes:       formStr(r, "notes"),
	}
}
