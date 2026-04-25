package handler

import (
	"errors"
	"net/http"

	"ride-hailing/services/driver/internal/application"
	"ride-hailing/services/driver/internal/application/commands"
	"ride-hailing/services/driver/internal/application/queries"
	"ride-hailing/services/driver/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler holds the HTTP handlers for the driver resource.
// It translates HTTP concepts (request/response) into application commands/queries.
type Handler struct {
	svc    *application.DriverService
	logger *zap.Logger
}

func NewHandler(svc *application.DriverService, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// ── Request structs ───────────────────────────────────────────────────────────

type registerDriverRequest struct {
	UserID       string `json:"user_id"       binding:"required"`
	Name         string `json:"name"          binding:"required"`
	Phone        string `json:"phone"         binding:"required"`
	VehicleMake  string `json:"vehicle_make"  binding:"required"`
	VehicleModel string `json:"vehicle_model" binding:"required"`
	VehicleYear  int    `json:"vehicle_year"  binding:"required,min=1990"`
	PlateNumber  string `json:"plate_number"  binding:"required"`
	VehicleType  string `json:"vehicle_type"  binding:"required,oneof=motorcycle car van"`
}

type changeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=online offline on_trip"`
}

type updateLocationRequest struct {
	Latitude  float64 `json:"latitude"  binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// @Summary     Register a new driver
// @Tags        drivers
// @Accept      json
// @Produce     json
// @Param       body body registerDriverRequest true "Driver details"
// @Success     201 {object} domain.Driver
// @Failure     400 {object} map[string]string
// @Router      /drivers [post]
func (h *Handler) RegisterDriver(c *gin.Context) {
	var req registerDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver, err := h.svc.RegisterDriver(c.Request.Context(), commands.RegisterDriverCommand{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		Name:         req.Name,
		Phone:        req.Phone,
		VehicleMake:  req.VehicleMake,
		VehicleModel: req.VehicleModel,
		VehicleYear:  req.VehicleYear,
		PlateNumber:  req.PlateNumber,
		VehicleType:  req.VehicleType,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, driver)
}

// @Summary     Get a driver by ID
// @Tags        drivers
// @Produce     json
// @Param       id path string true "Driver ID"
// @Success     200 {object} domain.Driver
// @Failure     404 {object} map[string]string
// @Router      /drivers/{id} [get]
func (h *Handler) GetDriver(c *gin.Context) {
	driver, err := h.svc.GetDriver(c.Request.Context(), queries.GetDriverQuery{
		DriverID: c.Param("id"),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, driver)
}

// @Summary     Change driver status
// @Tags        drivers
// @Accept      json
// @Produce     json
// @Param       id   path string             true "Driver ID"
// @Param       body body changeStatusRequest true "New status"
// @Success     200 {object} domain.Driver
// @Failure     422 {object} map[string]string
// @Router      /drivers/{id}/status [put]
func (h *Handler) ChangeStatus(c *gin.Context) {
	var req changeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver, err := h.svc.ChangeStatus(c.Request.Context(), commands.ChangeStatusCommand{
		DriverID:  c.Param("id"),
		NewStatus: req.Status,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, driver)
}

// @Summary     Update driver location
// @Tags        drivers
// @Accept      json
// @Param       id   path string                true "Driver ID"
// @Param       body body updateLocationRequest true "GPS coordinates"
// @Success     204
// @Failure     422 {object} map[string]string
// @Router      /drivers/{id}/location [put]
func (h *Handler) UpdateLocation(c *gin.Context) {
	var req updateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateLocation(c.Request.Context(), commands.UpdateLocationCommand{
		DriverID:  c.Param("id"),
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// handleError translates domain errors to HTTP status codes.
// This is the only place in the whole service that knows about both domains and HTTP.
func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDriverNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrDriverAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrDriverOffline),
		errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidPhone):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		h.logger.Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
