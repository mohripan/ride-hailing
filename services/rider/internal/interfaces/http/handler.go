package handler

import (
	"errors"
	"net/http"

	"ride-hailing/services/rider/internal/application"
	"ride-hailing/services/rider/internal/application/commands"
	"ride-hailing/services/rider/internal/application/queries"
	"ride-hailing/services/rider/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	svc    *application.RiderService
	logger *zap.Logger
}

func NewHandler(svc *application.RiderService, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

type registerRiderRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Name   string `json:"name"    binding:"required"`
	Phone  string `json:"phone"   binding:"required"`
	Email  string `json:"email"`
}

type updateProfileRequest struct {
	Name            string `json:"name"              binding:"required"`
	Phone           string `json:"phone"             binding:"required"`
	Email           string `json:"email"`
	ProfilePhotoURL string `json:"profile_photo_url"`
}

type topUpRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type addAddressRequest struct {
	Label     string  `json:"label"     binding:"required,oneof=home work other"`
	Address   string  `json:"address"   binding:"required"`
	Latitude  float64 `json:"latitude"  binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// @Summary     Register a new rider
// @Tags        riders
// @Accept      json
// @Produce     json
// @Param       body body registerRiderRequest true "Rider details"
// @Success     201 {object} domain.Rider
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /riders [post]
func (h *Handler) RegisterRider(c *gin.Context) {
	var req registerRiderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rider, err := h.svc.RegisterRider(c.Request.Context(), commands.RegisterRiderCommand{
		ID:     uuid.New().String(),
		UserID: req.UserID,
		Name:   req.Name,
		Phone:  req.Phone,
		Email:  req.Email,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rider)
}

// @Summary     Get a rider by ID
// @Tags        riders
// @Produce     json
// @Param       id path string true "Rider ID"
// @Success     200 {object} domain.Rider
// @Failure     404 {object} map[string]string
// @Router      /riders/{id} [get]
func (h *Handler) GetRider(c *gin.Context) {
	rider, err := h.svc.GetRider(c.Request.Context(), queries.GetRiderQuery{
		RiderID: c.Param("id"),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rider)
}

// @Summary     Update rider profile
// @Tags        riders
// @Accept      json
// @Produce     json
// @Param       id   path string               true "Rider ID"
// @Param       body body updateProfileRequest true "Profile details"
// @Success     200 {object} domain.Rider
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /riders/{id} [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rider, err := h.svc.UpdateProfile(c.Request.Context(), commands.UpdateProfileCommand{
		RiderID:         c.Param("id"),
		Name:            req.Name,
		Phone:           req.Phone,
		Email:           req.Email,
		ProfilePhotoURL: req.ProfilePhotoURL,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rider)
}

// @Summary     Top up rider wallet
// @Tags        riders
// @Accept      json
// @Produce     json
// @Param       id   path string       true "Rider ID"
// @Param       body body topUpRequest true "Top up amount"
// @Success     200 {object} domain.Rider
// @Failure     400 {object} map[string]string
// @Failure     422 {object} map[string]string
// @Router      /riders/{id}/wallet/topup [post]
func (h *Handler) TopUpWallet(c *gin.Context) {
	var req topUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rider, err := h.svc.TopUpWallet(c.Request.Context(), commands.TopUpWalletCommand{
		RiderID: c.Param("id"),
		Amount:  req.Amount,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rider)
}

// @Summary     Add a saved address
// @Tags        riders
// @Accept      json
// @Produce     json
// @Param       id   path string            true "Rider ID"
// @Param       body body addAddressRequest true "Address details"
// @Success     201 {object} domain.Rider
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /riders/{id}/addresses [post]
func (h *Handler) AddSavedAddress(c *gin.Context) {
	var req addAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rider, err := h.svc.AddSavedAddress(c.Request.Context(), commands.AddSavedAddressCommand{
		RiderID:   c.Param("id"),
		AddressID: uuid.New().String(),
		Label:     req.Label,
		Address:   req.Address,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rider)
}

// @Summary     Remove a saved address
// @Tags        riders
// @Produce     json
// @Param       id        path string true "Rider ID"
// @Param       addressId path string true "Address ID"
// @Success     200 {object} domain.Rider
// @Failure     404 {object} map[string]string
// @Router      /riders/{id}/addresses/{addressId} [delete]
func (h *Handler) RemoveSavedAddress(c *gin.Context) {
	rider, err := h.svc.RemoveSavedAddress(c.Request.Context(), commands.RemoveSavedAddressCommand{
		RiderID:   c.Param("id"),
		AddressID: c.Param("addressId"),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rider)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRiderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrRiderAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidPhone),
		errors.Is(err, domain.ErrInsufficientWallet),
		errors.Is(err, domain.ErrInvalidTopUpAmount),
		errors.Is(err, domain.ErrAddressNotFound):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		h.logger.Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
