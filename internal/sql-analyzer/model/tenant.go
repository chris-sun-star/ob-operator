package model

// Tenant represents a tenant with its ID.
type Tenant struct {
	ID uint64 `db:"tenant_id"`
}
