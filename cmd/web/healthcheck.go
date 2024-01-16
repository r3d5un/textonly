package main

import (
	"net/http"
)

type HealthCheckMessage struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

// @Summary		Healthcheck
// @Description	Endpoint to check if the API is running
// @Tags			healthcheck
// @Produce		json
// @Success		200	{object}	HealthCheckMessage
// @Failure		500	{object}	ErrorMessage
// @Router			/v1/healthcheck [get]
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	healthCheckMessage := HealthCheckMessage{
		Status:  "available",
		Version: version,
	}

	err := app.writeJSON(w, http.StatusOK, healthCheckMessage, nil)
	if err != nil {
		app.logger.ErrorContext(
			r.Context(),
			"an error occurred while returning healthcheck response",
			"request",
			r,
			"error",
			err,
		)
		app.serverErrorResponse(w, r, err)
		return
	}
}
