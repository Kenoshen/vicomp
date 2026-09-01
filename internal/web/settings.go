package web

import (
	"net/http"

	"github.com/mwingfield/vicomp/internal/models"
)

func (s *Server) settingsEdit(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.GetSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "settings_form", pageData{Title: "Settings", Data: st})
}

func (s *Server) settingsUpdate(w http.ResponseWriter, r *http.Request) {
	st := models.Settings{
		MakeChecksPayableTo:   formStr(r, "make_checks_payable_to"),
		DefaultSessionRate:    formFloat(r, "default_session_rate"),
		DefaultSessionMinutes: formInt(r, "default_session_minutes"),
		DefaultCPTCode:        formStr(r, "default_cpt_code"),
		PracticeName:          formStr(r, "practice_name"),
		Address:               formStr(r, "address"),
		City:                  formStr(r, "city"),
		State:                 formStr(r, "state"),
		Zip:                   formStr(r, "zip"),
		Phone:                 formStr(r, "phone"),
		Email:                 formStr(r, "email"),
		TaxID:                 formStr(r, "tax_id"),
		NPI:                   formStr(r, "npi"),
	}
	if st.DefaultCPTCode == "" {
		st.DefaultCPTCode = "90834"
	}
	if err := s.store.UpdateSettings(r.Context(), st); err != nil {
		s.render(w, "settings_form", pageData{Title: "Settings", Error: err.Error(), Data: st})
		return
	}
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
