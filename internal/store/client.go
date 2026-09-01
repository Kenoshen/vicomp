package store

import (
	"context"

	"github.com/mwingfield/vicomp/internal/models"
)

const clientSelect = `
	SELECT cl.id, cl.first_name, cl.last_name, cl.date_of_birth, cl.county_id,
		cl.clinician_id, cl.claim_number, cl.notes, cl.created_at,
		COALESCE(co.name, ''),
		COALESCE(cli.first_name || ' ' || cli.last_name, '')
	FROM clients cl
	LEFT JOIN counties co ON co.id = cl.county_id
	LEFT JOIN clinicians cli ON cli.id = cl.clinician_id`

func scanClient(row scanner) (models.Client, error) {
	var c models.Client
	err := row.Scan(&c.ID, &c.FirstName, &c.LastName, &c.DateOfBirth, &c.CountyID,
		&c.ClinicianID, &c.ClaimNumber, &c.Notes, &c.CreatedAt,
		&c.CountyName, &c.ClinicianName)
	return c, err
}

func (s *Store) ListClients(ctx context.Context) ([]models.Client, error) {
	rows, err := s.pool.Query(ctx, clientSelect+` ORDER BY cl.last_name, cl.first_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Client
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListClientsByClinician backs the "clinicians have clients" view.
func (s *Store) ListClientsByClinician(ctx context.Context, clinicianID int64) ([]models.Client, error) {
	rows, err := s.pool.Query(ctx, clientSelect+`
		WHERE cl.clinician_id = $1 ORDER BY cl.last_name, cl.first_name`, clinicianID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Client
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetClient(ctx context.Context, id int64) (models.Client, error) {
	return scanClient(s.pool.QueryRow(ctx, clientSelect+` WHERE cl.id = $1`, id))
}

func (s *Store) CreateClient(ctx context.Context, c models.Client) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO clients (first_name, last_name, date_of_birth, county_id,
			clinician_id, claim_number, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		c.FirstName, c.LastName, c.DateOfBirth, c.CountyID, c.ClinicianID,
		c.ClaimNumber, c.Notes).Scan(&id)
	return id, err
}

func (s *Store) UpdateClient(ctx context.Context, c models.Client) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE clients SET first_name=$1, last_name=$2, date_of_birth=$3,
			county_id=$4, clinician_id=$5, claim_number=$6, notes=$7 WHERE id=$8`,
		c.FirstName, c.LastName, c.DateOfBirth, c.CountyID, c.ClinicianID,
		c.ClaimNumber, c.Notes, c.ID)
	return err
}

func (s *Store) DeleteClient(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM clients WHERE id = $1`, id)
	return err
}
