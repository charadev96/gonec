package shared

import (
	"github.com/google/uuid"
)

func UUIDToPB(id uuid.UUID) string {
	return id.String()
}

func UUIDFromPB(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
