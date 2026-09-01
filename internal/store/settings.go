package store

import (
	"context"

	"github.com/mwingfield/vicomp/internal/models"
)

// GetSettings returns the singleton settings row, creating it if the migration
// somehow left it absent.
func (s *Store) GetSettings(ctx context.Context) (models.Settings, error) {
	var st models.Settings
	err := s.pool.QueryRow(ctx, `
		SELECT make_checks_payable_to, default_session_rate,
			default_session_minutes, default_cpt_code,
			practice_name, address, city, state, zip, phone, email, tax_id, npi
		FROM settings WHERE id = 1`).
		Scan(&st.MakeChecksPayableTo, &st.DefaultSessionRate,
			&st.DefaultSessionMinutes, &st.DefaultCPTCode,
			&st.PracticeName, &st.Address, &st.City, &st.State, &st.Zip,
			&st.Phone, &st.Email, &st.TaxID, &st.NPI)
	return st, err
}

func (s *Store) UpdateSettings(ctx context.Context, st models.Settings) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE settings SET make_checks_payable_to = $1, default_session_rate = $2,
			default_session_minutes = $3, default_cpt_code = $4,
			practice_name = $5, address = $6, city = $7, state = $8, zip = $9,
			phone = $10, email = $11, tax_id = $12, npi = $13
		WHERE id = 1`,
		st.MakeChecksPayableTo, st.DefaultSessionRate,
		st.DefaultSessionMinutes, st.DefaultCPTCode,
		st.PracticeName, st.Address, st.City, st.State, st.Zip,
		st.Phone, st.Email, st.TaxID, st.NPI)
	return err
}
