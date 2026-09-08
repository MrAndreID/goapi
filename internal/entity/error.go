package entity

import "errors"

type ErrorStatus struct {
	Err    error
	Status int
}

func StatusForError(err error, table []ErrorStatus, fallback int) int {
	for _, entry := range table {
		if errors.Is(err, entry.Err) {
			return entry.Status
		}
	}

	return fallback
}
