package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
)

// TicketHandler handles support ticket endpoints.
type TicketHandler struct{}

func NewTicketHandler() *TicketHandler {
	return &TicketHandler{}
}

// List returns all tickets for the authenticated user.
func (h *TicketHandler) List(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	tickets, err := database.ListTickets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets, "total": len(tickets)})
}

// Get returns a single ticket with its replies.
func (h *TicketHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ticket, err := database.GetTicketByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	c.JSON(http.StatusOK, ticket)
}

// CreateRequest is the ticket creation payload.
type CreateTicketRequest struct {
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// Create opens a new support ticket.
func (h *TicketHandler) Create(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject and message required"})
		return
	}

	adminID, _ := c.Get("admin_id")
	ticket := &database.Ticket{
		UserID:  adminID.(int64),
		Subject: req.Subject,
		Message: req.Message,
		Status:  "open",
	}
	if err := database.CreateTicket(ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ticket)
}

// ReplyRequest adds a reply to a ticket.
type ReplyTicketRequest struct {
	Message string `json:"message" binding:"required"`
	IsAdmin bool   `json:"is_admin"`
}

// Reply adds a response to a ticket.
func (h *TicketHandler) Reply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ReplyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message required"})
		return
	}

	adminID, _ := c.Get("admin_id")
	reply := &database.TicketReply{
		TicketID: id,
		UserID:   adminID.(int64),
		IsAdmin:  req.IsAdmin,
		Message:  req.Message,
	}
	if err := database.CreateTicketReply(reply); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update ticket status
	ticket, err := database.GetTicketByID(id)
	if err == nil {
		ticket.Status = "answered"
		database.UpdateTicket(ticket)
	}

	c.JSON(http.StatusCreated, reply)
}

// Close changes a ticket's status to closed.
func (h *TicketHandler) Close(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ticket, err := database.GetTicketByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	ticket.Status = "closed"
	database.UpdateTicket(ticket)
	c.JSON(http.StatusOK, gin.H{"message": "ticket closed"})
}

// Delete removes a ticket and its replies.
func (h *TicketHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := database.DeleteTicket(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ticket deleted"})
}

// Stats returns ticket statistics.
func (h *TicketHandler) Stats(c *gin.Context) {
	tickets, err := database.ListTickets(0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var open, answered, closed int
	for _, t := range tickets {
		switch t.Status {
		case "open":
			open++
		case "answered":
			answered++
		case "closed":
			closed++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"total":    len(tickets),
		"open":     open,
		"answered": answered,
		"closed":   closed,
	})
}
