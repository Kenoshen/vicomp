package web

import (
	"net/http"
	"time"

	"github.com/mwingfield/vicomp/internal/models"
)

type sessionFormData struct {
	Session models.Session
	Clients []models.Client
}

func (s *Server) sessionsList(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "sessions_list", pageData{Title: "Sessions", Data: list})
}

func (s *Server) sessionFormPage(w http.ResponseWriter, r *http.Request, title string, sess models.Session, errMsg string) {
	clients, err := s.store.ListClients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "session_form", pageData{
		Title: title,
		Error: errMsg,
		Data:  sessionFormData{Session: sess, Clients: clients},
	})
}

func (s *Server) sessionNew(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.GetSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sessionFormPage(w, r, "New Session", models.Session{
		SessionDate:     time.Now(),
		DurationMinutes: st.DefaultSessionMinutes,
		Amount:          st.DefaultSessionRate,
		CPTCode:         st.DefaultCPTCode,
	}, "")
}

func (s *Server) sessionEdit(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	sess, err := s.store.GetSession(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if sess.InvoiceID != nil {
		http.Error(w, "session is already invoiced and cannot be edited", http.StatusConflict)
		return
	}
	s.sessionFormPage(w, r, "Edit Session", sess, "")
}

func (s *Server) sessionCreate(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromForm(r)
	if _, err := s.store.CreateSession(r.Context(), sess); err != nil {
		s.sessionFormPage(w, r, "New Session", sess, err.Error())
		return
	}
	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

func (s *Server) sessionUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	sess := sessionFromForm(r)
	sess.ID = id
	if err := s.store.UpdateSession(r.Context(), sess); err != nil {
		s.sessionFormPage(w, r, "Edit Session", sess, err.Error())
		return
	}
	http.Redirect(w, r, "/sessions", http.StatusSeeOther)
}

func (s *Server) sessionDelete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteSession(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func sessionFromForm(r *http.Request) models.Session {
	cpt := formStr(r, "cpt_code")
	if cpt == "" {
		cpt = "90834"
	}
	return models.Session{
		ClientID:        int64(formInt(r, "client_id")),
		SessionDate:     formDate(r, "session_date"),
		DurationMinutes: formInt(r, "duration_minutes"),
		Amount:          formFloat(r, "amount"),
		CPTCode:         cpt,
		Notes:           formStr(r, "notes"),
	}
}
