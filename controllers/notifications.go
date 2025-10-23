package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"notification/channels"
	"notification/models"
	"notification/services/notifier"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	svc *notifier.NotifierService
}

func NewNotificationController(svc *notifier.NotifierService) *NotificationController {
	return &NotificationController{svc: svc}
}

type createNotificationDTO struct {
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	ChannelName string         `json:"channel_name"`
	Meta        map[string]any `json:"meta"`
}

func (dto *createNotificationDTO) normalizeMeta() map[string]string {
	normalizedMeta := make(map[string]string, len(dto.Meta))
	for k, v := range dto.Meta {
		switch val := v.(type) {
		case string:
			normalizedMeta[k] = val
		default:
			b, _ := json.Marshal(val)
			normalizedMeta[k] = string(b)
		}
	}
	return normalizedMeta
}

// @Summary Create notification
// @Description Create and enqueue a notification. Supports multiple channels: email, sms, and push.
// @Description
// @Description **Email Channel** - See channels.ValidEmailMeta for required meta fields
// @Description **SMS Channel** - See channels.ValidSMSMeta for required meta fields
// @Description **Push Channel** - See channels.ValidPushMeta for required meta fields
// @Description
// @Description **Example:** {"title":"Welcome","content":"Welcome message","channel_name":"email","meta":{"to":"user@example.com","subject":"Welcome!"}}
// @Tags notifications
// @Accept json
// @Produce json
// @Param data body models.CreateNotificationRequest true "Notification data"
// @Success 202 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /notifications [post]
func (nc *NotificationController) CreateNotification(c *gin.Context) {
	var dto createNotificationDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	req := notifier.NotificationRequest{
		Title:       dto.Title,
		Content:     dto.Content,
		ChannelName: dto.ChannelName,
		Meta:        dto.normalizeMeta(),
		UserID:      user.(models.User).ID,
	}

	if err := nc.svc.CreateAndEnqueue(c.Request.Context(), req); err != nil {
		if errors.Is(err, notifier.ErrInvalidChannel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel name"})
			return
		}
		if errors.Is(err, notifier.ErrInvalidMetadata) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "Notification created and enqueued"})
}

// @Summary List notifications
// @Description List user notifications
// @Tags notifications
// @Produce json
// @Success 200 {array} models.Notification
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /notifications [get]
func (nc *NotificationController) ListNotifications(c *gin.Context) {
	var list []models.Notification
	list, err := nc.svc.ListNotifications(c.Request.Context(), c.GetInt("user_id"), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// @Summary Get notification
// @Description Get a notification by ID
// @Tags notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} models.Notification
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /notifications/{id} [get]
func (nc *NotificationController) GetNotification(c *gin.Context) {
	idParam := c.Param("id")
	id, _ := strconv.Atoi(idParam)
	n, err := nc.svc.GetNotification(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, notifier.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, n)
}

// @Summary Update notification
// @Description Update allowed fields of a notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Param patch body map[string]any true "Partial update"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /notifications/{id} [patch]
func (nc *NotificationController) UpdateNotification(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := nc.svc.UpdateNotification(c.Request.Context(), id, patch); err != nil {
		if errors.Is(err, notifier.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Delete notification
// @Description Mark a notification as deleted
// @Tags notifications
// @Param id path int true "Notification ID"
// @Success 204 "No Content"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /notifications/{id} [delete]
func (nc *NotificationController) DeleteNotification(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := nc.svc.DeleteNotification(c.Request.Context(), id); err != nil {
		if errors.Is(err, notifier.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Get channel schemas
// @Description Get the required meta field schemas for each notification channel
// @Tags notifications
// @Produce json
// @Success 200 {object} models.ChannelSchemasResponse
// @Router /notifications/channels/schemas [get]
func (nc *NotificationController) GetChannelSchemas(c *gin.Context) {
	schemas := models.ChannelSchemasResponse{
		Email: channels.ValidEmailMeta{
			To:       "user@example.com",
			Subject:  "Email subject",
			Template: "titled",
		},
		SMS: channels.ValidSMSMeta{
			Phone:    "+1234567890",
			SendDate: "2024-10-21",
		},
		Push: channels.ValidPushMeta{
			Token:    "device_token_xyz123",
			Platform: "android",
			Data:     map[string]string{"message_id": "123"},
			Options:  map[string]string{"priority": "high"},
		},
	}
	c.JSON(http.StatusOK, schemas)
}
