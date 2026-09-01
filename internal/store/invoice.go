package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mwingfield/vicomp/internal/models"
)

const invoiceSelect = `
	SELECT i.id, i.client_id, i.invoice_number, i.status, i.total, i.created_at,
		(c.first_name || ' ' || c.last_name) AS client_name
	FROM invoices i
	JOIN clients c ON c.id = i.client_id`

func scanInvoice(row scanner) (models.Invoice, error) {
	var i models.Invoice
	err := row.Scan(&i.ID, &i.ClientID, &i.InvoiceNumber, &i.Status, &i.Total,
		&i.CreatedAt, &i.ClientName)
	return i, err
}

func (s *Store) ListInvoices(ctx context.Context) ([]models.Invoice, error) {
	rows, err := s.pool.Query(ctx, invoiceSelect+` ORDER BY i.created_at DESC, i.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (s *Store) GetInvoice(ctx context.Context, id int64) (models.Invoice, error) {
	return scanInvoice(s.pool.QueryRow(ctx, invoiceSelect+` WHERE i.id = $1`, id))
}

// ErrNoSessions is returned when an invoice generation is attempted for a client
// that has no uninvoiced sessions.
var ErrNoSessions = errors.New("client has no uninvoiced sessions")

// CreateInvoiceFromSessions gathers every uninvoiced session for the client,
// creates an invoice for their sum, and stamps each of those sessions with the
// new invoice id so they can never be invoiced again. All in one transaction.
func (s *Store) CreateInvoiceFromSessions(ctx context.Context, clientID int64) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var (
		count int
		total float64
	)
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM sessions WHERE client_id = $1 AND invoice_id IS NULL`, clientID).
		Scan(&count, &total)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, ErrNoSessions
	}

	number := fmt.Sprintf("INV-%s-%d", time.Now().Format("20060102-150405"), clientID)

	var invoiceID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO invoices (client_id, invoice_number, status, total)
		VALUES ($1, $2, 'draft', $3) RETURNING id`,
		clientID, number, total).Scan(&invoiceID)
	if err != nil {
		return 0, err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE sessions SET invoice_id = $1
		WHERE client_id = $2 AND invoice_id IS NULL`, invoiceID, clientID)
	if err != nil {
		return 0, err
	}
	if tag.RowsAffected() != int64(count) {
		return 0, errors.New("session count changed during invoice creation")
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return invoiceID, nil
}

func (s *Store) UpdateInvoiceStatus(ctx context.Context, id int64, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE invoices SET status = $1 WHERE id = $2`, status, id)
	return err
}

// DeleteInvoice removes the invoice; its sessions fall back to uninvoiced via the
// ON DELETE SET NULL foreign key, so they become eligible for invoicing again.
func (s *Store) DeleteInvoice(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM invoices WHERE id = $1`, id)
	return err
}
