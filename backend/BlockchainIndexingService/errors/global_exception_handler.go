package errors

import (
	"BlockchainIndexingService/utils"
	"errors"
	"net/http"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrNotFound           = errors.New("requested resource not found")
	ErrInvalidAddress     = errors.New("invalid bitcoin address format")
	ErrNotImplemented     = errors.New("feature not implemented")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrInternalServer     = errors.New("internal server error")
)

func GlobalErrorHandler(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidAddress) || errors.Is(err, ErrInvalidArgument):
		utils.WriteErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		utils.WriteErr(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrNotImplemented):
		utils.WriteErr(w, http.StatusNotImplemented, err.Error())

	case errors.Is(err, ErrInternalServer):
	default:
		utils.WriteErr(w, http.StatusInternalServerError, "internal server error")
	}
}
