package handler

import (
	"github.com/google/uuid"
)

func ParseUUIDs(ids ...string) ([]uuid.UUID, error) {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		u, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, u)
	}
	return parsed, nil
}
