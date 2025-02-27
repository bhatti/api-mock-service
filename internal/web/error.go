package web

import (
	"errors"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
)

func HandleError(c APIContext, err error) error {
	if err == nil {
		return nil
	}
	log.WithFields(log.Fields{
		"Method":  c.Request().Method,
		"Request": c.Request().URL,
		"Error":   err,
	}).Warnf("failed request")
	var validationErr *types.ValidationError
	var notFoundErr *types.NotFoundError
	if errors.As(err, &validationErr) {
		return c.String(400, err.Error())
	} else if errors.As(err, &notFoundErr) {
		return c.String(404, err.Error())
	}
	return c.String(500, err.Error())
}
