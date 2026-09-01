package store

import (
	"context"

	"github.com/mwingfield/vicomp/internal/models"
)

func (s *Store) ListClinicians(ctx context.Context) ([]models.Clinician, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, first_name, last_name, credentials, npi, tax_id, address, city,
			state, zip, phone, email, created_at
		FROM clinicians ORDER BY last_name, first_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Clinician
	for rows.Next() {
		c, err := scanClinician(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetClinician(ctx context.Context, id int64) (models.Clinician, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, first_name, last_name, credentials, npi, tax_id, address, city,
			state, zip, phone, email, created_at
		FROM clinicians WHERE id = $1`, id)
	return scanClinician(row)
}

func (s *Store) CreateClinician(ctx context.Context, c models.Clinician) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO clinicians (first_name, last_name, credentials, npi, tax_id,
			address, city, state, zip, phone, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
		c.FirstName, c.LastName, c.Credentials, c.NPI, c.TaxID,
		c.Address, c.City, c.State, c.Zip, c.Phone, c.Email).Scan(&id)
	return id, err
}

func (s *Store) UpdateClinician(ctx context.Context, c models.Clinician) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE clinicians SET first_name=$1, last_name=$2, credentials=$3, npi=$4,
			tax_id=$5, address=$6, city=$7, state=$8, zip=$9, phone=$10, email=$11
		WHERE id=$12`,
		c.FirstName, c.LastName, c.Credentials, c.NPI, c.TaxID,
		c.Address, c.City, c.State, c.Zip, c.Phone, c.Email, c.ID)
	return err
}

func (s *Store) DeleteClinician(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM clinicians WHERE id = $1`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanClinician(row scanner) (models.Clinician, error) {
	var c models.Clinician
	err := row.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Credentials, &c.NPI, &c.TaxID,
		&c.Address, &c.City, &c.State, &c.Zip, &c.Phone, &c.Email, &c.CreatedAt)
	return c, err
}
