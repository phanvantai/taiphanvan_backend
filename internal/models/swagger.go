package models

import "time"

// SwaggerDeletedAt is a custom type for Swagger documentation
// @Description A timestamp for soft-deleted records (null if not deleted)
type SwaggerDeletedAt struct {
	Time  time.Time `json:"time,omitempty"`
	Valid bool      `json:"valid,omitempty"`
}
