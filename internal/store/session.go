package store

import (
	"context"

	"github.com/mwingfield/vicomp/internal/models"
)

const sessionSelect = `
	SELECT s.id, s.client_id, s.session_date, s.duration_minutes, s.amount,
		s.cpt_code, s.notes, s.invoice_id, s.created_at,
		(c.first_name || ' ' || c.last_name) AS client_name
	FROM sessions s
	JOIN clients c ON c.id = s.client_id`

func scanSession(row scanner) (models.Session, error) {
	var s models.Session
	err := row.Scan(&s.ID, &s.ClientID, &s.SessionDate, &s.DurationMinutes,
		&s.Amount, &s.CPTCode, &s.Notes, &s.InvoiceID, &s.CreatedAt, &s.ClientName)
	return s, err
}

func (s *Store) ListSessions(ctx context.Context) ([]models.Session, error) {
	rows, err := s.pool.Query(ctx, sessionSelect+` ORDER BY s.session_date DESC, s.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectSessions(rows)
}

// ListUninvoicedSessions returns the sessions eligible to be rolled into a new
// invoice for a client: those with no invoice_id yet.
func (s *Store) ListUninvoicedSessions(ctx context.Context, clientID int64) ([]models.Session, error) {
	rows, err := s.pool.Query(ctx, sessionSelect+`
		WHERE s.client_id = $1 AND s.invoice_id IS NULL
		ORDER BY s.session_date`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectSessions(rows)
}

func (s *Store) ListSessionsByInvoice(ctx context.Context, invoiceID int64) ([]models.Session, error) {
	rows, err := s.pool.Query(ctx, sessionSelect+`
		WHERE s.invoice_id = $1 ORDER BY s.session_date`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectSessions(rows)
}

func collectSessions(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]models.Session, error) {
	var out []models.Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *Store) GetSession(ctx context.Context, id int64) (models.Session, error) {
	return scanSession(s.pool.QueryRow(ctx, sessionSelect+` WHERE s.id = $1`, id))
}

func (s *Store) CreateSession(ctx context.Context, sess models.Session) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sessions (client_id, session_date, duration_minutes, amount, cpt_code, notes)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		sess.ClientID, sess.SessionDate, sess.DurationMinutes, sess.Amount,
		sess.CPTCode, sess.Notes).Scan(&id)
	return id, err
}

// UpdateSession refuses to touch a session that has already been invoiced so the
// invoice total stays consistent.
func (s *Store) UpdateSession(ctx context.Context, sess models.Session) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sessions SET client_id=$1, session_date=$2, duration_minutes=$3,
			amount=$4, cpt_code=$5, notes=$6
		WHERE id=$7 AND invoice_id IS NULL`,
		sess.ClientID, sess.SessionDate, sess.DurationMinutes, sess.Amount,
		sess.CPTCode, sess.Notes, sess.ID)
	return err
}

func (s *Store) DeleteSession(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1 AND invoice_id IS NULL`, id)
	return err
}
