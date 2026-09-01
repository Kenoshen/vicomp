package web

import (
	"net/http"

	"github.com/mwingfield/vicomp/internal/models"
)

type clinicianShowData struct {
	Clinician models.Clinician
	Clients   []models.Client
}

func (s *Server) cliniciansList(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListClinicians(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "clinicians_list", pageData{Title: "Clinicians", Data: list})
}

func (s *Server) clinicianNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "clinician_form", pageData{Title: "New Clinician", Data: models.Clinician{}})
}

func (s *Server) clinicianShow(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c, err := s.store.GetClinician(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	clients, err := s.store.ListClientsByClinician(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "clinician_show", pageData{
		Title: c.FullName(),
		Data:  clinicianShowData{Clinician: c, Clients: clients},
	})
}

func (s *Server) clinicianEdit(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c, err := s.store.GetClinician(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.render(w, "clinician_form", pageData{Title: "Edit Clinician", Data: c})
}

func (s *Server) clinicianCreate(w http.ResponseWriter, r *http.Request) {
	c := clinicianFromForm(r)
	if _, err := s.store.CreateClinician(r.Context(), c); err != nil {
		s.render(w, "clinician_form", pageData{Title: "New Clinician", Error: err.Error(), Data: c})
		return
	}
	http.Redirect(w, r, "/clinicians", http.StatusSeeOther)
}

func (s *Server) clinicianUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c := clinicianFromForm(r)
	c.ID = id
	if err := s.store.UpdateClinician(r.Context(), c); err != nil {
		s.render(w, "clinician_form", pageData{Title: "Edit Clinician", Error: err.Error(), Data: c})
		return
	}
	http.Redirect(w, r, "/clinicians", http.StatusSeeOther)
}

func (s *Server) clinicianDelete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteClinician(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func clinicianFromForm(r *http.Request) models.Clinician {
	return models.Clinician{
		FirstName:   formStr(r, "first_name"),
		LastName:    formStr(r, "last_name"),
		Credentials: formStr(r, "credentials"),
		NPI:         formStr(r, "npi"),
		TaxID:       formStr(r, "tax_id"),
		Address:     formStr(r, "address"),
		City:        formStr(r, "city"),
		State:       formStr(r, "state"),
		Zip:         formStr(r, "zip"),
		Phone:       formStr(r, "phone"),
		Email:       formStr(r, "email"),
	}
}
