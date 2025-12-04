package uuid

import "github.com/google/uuid"

// UUIDGenerator abstracts UUID generation for testability
type UUIDGenerator interface {
	New() string
}

// SystemUUID is the production implementation of UUIDGenerator
type SystemUUID struct{}

func (u *SystemUUID) New() string {
	return uuid.New().String()
}

// New creates a new UUIDGenerator instance
func New() UUIDGenerator {
	return &SystemUUID{}
}
