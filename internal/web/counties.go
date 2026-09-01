package web

import (
	"net/http"

	"github.com/mwingfield/vicomp/internal/models"
)

func (s *Server) countiesList(w http.ResponseWriter, r *http.Request) {
	counties, err := s.store.ListCounties(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "counties_list", pageData{Title: "Counties", Data: counties})
}

func (s *Server) countyNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "county_form", pageData{Title: "New County", Data: models.County{}})
}

func (s *Server) countyEdit(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c, err := s.store.GetCounty(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.render(w, "county_form", pageData{Title: "Edit County", Data: c})
}

func (s *Server) countyCreate(w http.ResponseWriter, r *http.Request) {
	c := countyFromForm(r)
	if _, err := s.store.CreateCounty(r.Context(), c); err != nil {
		s.render(w, "county_form", pageData{Title: "New County", Error: err.Error(), Data: c})
		return
	}
	http.Redirect(w, r, "/counties", http.StatusSeeOther)
}

func (s *Server) countyUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	c := countyFromForm(r)
	c.ID = id
	if err := s.store.UpdateCounty(r.Context(), c); err != nil {
		s.render(w, "county_form", pageData{Title: "Edit County", Error: err.Error(), Data: c})
		return
	}
	http.Redirect(w, r, "/counties", http.StatusSeeOther)
}

func (s *Server) countyDelete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteCounty(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func countyFromForm(r *http.Request) models.County {
	return models.County{
		Name:        formStr(r, "name"),
		ContactName: formStr(r, "contact_name"),
		Address:     formStr(r, "address"),
		City:        formStr(r, "city"),
		State:       formStr(r, "state"),
		Zip:         formStr(r, "zip"),
		Phone:       formStr(r, "phone"),
		Email:       formStr(r, "email"),
	}
}
