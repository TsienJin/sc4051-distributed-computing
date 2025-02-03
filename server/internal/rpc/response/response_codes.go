package response

import "net/http"

type StatusCode uint16

const (
	StatusOk StatusCode = http.StatusOK

	StatusBadRequest StatusCode = http.StatusBadRequest
	StatusNotFound   StatusCode = http.StatusNotFound

	StatusInternalServerError StatusCode = http.StatusInternalServerError
)
