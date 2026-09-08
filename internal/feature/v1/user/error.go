package user

import (
	"errors"
	"net/http"

	"github.com/MrAndreID/goapi/v2/internal/entity"
)

var (
	ErrDuplicateEmail          = errors.New("DUPLICATE_EMAIL")
	ErrFailedToCreateUser      = errors.New("FAILED_TO_CREATE_USER")
	ErrFailedToCreateEmail     = errors.New("FAILED_TO_CREATE_EMAIL")
	ErrFailedToReadUserData    = errors.New("FAILED_TO_READ_USER_DATA")
	ErrFailedToUpdateUserData  = errors.New("FAILED_TO_UPDATE_USER_DATA")
	ErrFailedToDeleteUserData  = errors.New("FAILED_TO_DELETE_USER_DATA")
	ErrFailedToDeleteEmailData = errors.New("FAILED_TO_DELETE_EMAIL_DATA")
)

var errorStatuses = []entity.ErrorStatus{
	{Err: ErrDuplicateEmail, Status: http.StatusConflict},
	{Err: ErrFailedToReadUserData, Status: http.StatusNotFound},
}

func StatusFor(err error) int {
	return entity.StatusForError(err, errorStatuses, http.StatusInternalServerError)
}
