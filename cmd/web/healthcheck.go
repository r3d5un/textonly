package main

import (
	"net/http"

	"textonly.islandwind.me/internal/utils"
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
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	healthCheckMessage := HealthCheckMessage{
		Status:      "available",
		Environment: app.config.Server.ENV,
		Version:     version,
	}

	logger.Info("writing response", "response", healthCheckMessage)
	err := app.writeJSON(w, http.StatusOK, healthCheckMessage, nil)
	if err != nil {
		logger.Error(
			"an error occurred while returning healthcheck response",
			"error", err,
		)
		app.serverErrorResponse(w, r, err)
		return
	}
}
