package models

import "time"

type Customer struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CustomerUser struct {
	ID         int       `db:"id"`
	CustomerID int       `db:"customer_id"`
	Email      string    `db:"email"`
	CreatedAt  time.Time `db:"created_at"`
}

type VPSInstance struct {
	ID            int       `db:"id"`
	CustomerID    int       `db:"customer_id"`
	Hostname      string    `db:"hostname"`
	IPAddress     string    `db:"ip_address"`
	ServerDetails string    `db:"server_details"` // JSON string
	OS            string    `db:"os"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
