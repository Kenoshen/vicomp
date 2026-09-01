package models

import "time"

type County struct {
	ID          int64
	Name        string
	ContactName string
	Address     string
	City        string
	State       string
	Zip         string
	Phone       string
	Email       string
	CreatedAt   time.Time
}

type Clinician struct {
	ID          int64
	FirstName   string
	LastName    string
	Credentials string
	NPI         string
	TaxID       string
	Address     string
	City        string
	State       string
	Zip         string
	Phone       string
	Email       string
	CreatedAt   time.Time
}

func (c Clinician) FullName() string {
	name := c.FirstName + " " + c.LastName
	if c.Credentials != "" {
		name += ", " + c.Credentials
	}
	return name
}

type Client struct {
	ID          int64
	FirstName   string
	LastName    string
	DateOfBirth *time.Time
	CountyID    *int64
	ClinicianID *int64
	ClaimNumber string
	Notes       string
	CreatedAt   time.Time

	// Optional joined data, populated by some queries.
	CountyName    string
	ClinicianName string
}

func (c Client) FullName() string { return c.FirstName + " " + c.LastName }

type Session struct {
	ID              int64
	ClientID        int64
	SessionDate     time.Time
	DurationMinutes int
	Amount          float64
	CPTCode         string
	Notes           string
	InvoiceID       *int64
	CreatedAt       time.Time

	ClientName string
}

// Settings is the single-row configuration for the app.
type Settings struct {
	MakeChecksPayableTo   string
	DefaultSessionRate    float64
	DefaultSessionMinutes int
	DefaultCPTCode        string

	// Practice / billing identity — the invoice "From" block. A clinician's own
	// non-empty value for the same field overrides these on their invoices.
	PracticeName string
	Address      string
	City         string
	State        string
	Zip          string
	Phone        string
	Email        string
	TaxID        string
	NPI          string
}

type Invoice struct {
	ID            int64
	ClientID      int64
	InvoiceNumber string
	Status        string
	Total         float64
	CreatedAt     time.Time

	ClientName string
}
