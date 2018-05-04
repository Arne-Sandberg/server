package utils

import (
	"github.com/freecloudio/freecloud/models"
	"net/http"
	"fmt"
)

func PbOK() *models.DefaultResponse {
	return &models.DefaultResponse{ResponseCode: http.StatusOK, ErrorMessage: "ok"}
}

func PbCreated() *models.DefaultResponse {
	return &models.DefaultResponse{ ResponseCode: http.StatusCreated, ErrorMessage: "created" }
}

func PbBadRequest(msg string, a ...interface{}) *models.DefaultResponse {
	return &models.DefaultResponse{ ResponseCode: http.StatusBadRequest, ErrorMessage: fmt.Sprintf(msg, a) }
}

func PbForbidden() *models.DefaultResponse {
	return &models.DefaultResponse{ ResponseCode: http.StatusForbidden, ErrorMessage: "admin rights required" }
}

func PbUnauthorized(msg string, a ...interface{}) *models.DefaultResponse {
	return &models.DefaultResponse{ ResponseCode: http.StatusUnauthorized, ErrorMessage: fmt.Sprintf(msg, a) }
}

func PbInternalServerError(msg string, a ...interface{}) *models.DefaultResponse {
	return &models.DefaultResponse{ ResponseCode: http.StatusInternalServerError, ErrorMessage: fmt.Sprintf(msg, a) }
}