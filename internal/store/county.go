package store

import (
	"context"

	"github.com/mwingfield/vicomp/internal/models"
)

func (s *Store) ListCounties(ctx context.Context) ([]models.County, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, contact_name, address, city, state, zip, phone, email, created_at
		FROM counties ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.County
	for rows.Next() {
		var c models.County
		if err := rows.Scan(&c.ID, &c.Name, &c.ContactName, &c.Address, &c.City,
			&c.State, &c.Zip, &c.Phone, &c.Email, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCounty(ctx context.Context, id int64) (models.County, error) {
	var c models.County
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, contact_name, address, city, state, zip, phone, email, created_at
		FROM counties WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.ContactName, &c.Address, &c.City,
			&c.State, &c.Zip, &c.Phone, &c.Email, &c.CreatedAt)
	return c, err
}

func (s *Store) CreateCounty(ctx context.Context, c models.County) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO counties (name, contact_name, address, city, state, zip, phone, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		c.Name, c.ContactName, c.Address, c.City, c.State, c.Zip, c.Phone, c.Email).Scan(&id)
	return id, err
}

func (s *Store) UpdateCounty(ctx context.Context, c models.County) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE counties SET name=$1, contact_name=$2, address=$3, city=$4,
			state=$5, zip=$6, phone=$7, email=$8 WHERE id=$9`,
		c.Name, c.ContactName, c.Address, c.City, c.State, c.Zip, c.Phone, c.Email, c.ID)
	return err
}

func (s *Store) DeleteCounty(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM counties WHERE id = $1`, id)
	return err
}
