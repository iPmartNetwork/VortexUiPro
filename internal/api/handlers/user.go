package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vortexuipro/internal/database"
	"vortexuipro/internal/rbac"
	"vortexuipro/internal/service"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// getAdminID extracts the admin ID from the context.
func getAdminID(c *gin.Context) int64 {
	id, _ := c.Get("admin_id")
	switch v := id.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}

func (h *UserHandler) List(c *gin.Context) {
	adminID, _ := strconv.ParseInt(c.Query("admin_id"), 10, 64)
	users, err := h.userSvc.ListUsers(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": len(users)})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user, err := h.userSvc.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type CreateUserRequest struct {
	Username   string `json:"username" binding:"required"`
	Email      string `json:"email,omitempty"`
	DataLimit  int64  `json:"data_limit,omitempty"`
	ExpiryTime int64  `json:"expiry_time,omitempty"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	user, err := h.userSvc.CreateUser(req.Username, req.Email, req.DataLimit, req.ExpiryTime)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// UpdateUserRequest is the user update payload.
type UpdateUserRequest struct {
	Username   string `json:"username,omitempty"`
	Email      string `json:"email,omitempty"`
	DataLimit  *int64 `json:"data_limit,omitempty"`
	ExpiryTime *int64 `json:"expiry_time,omitempty"`
	Status     string `json:"status,omitempty"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.userSvc.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.DataLimit != nil {
		user.DataLimit = *req.DataLimit
	}
	if req.ExpiryTime != nil {
		user.ExpiryTime = *req.ExpiryTime
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := h.userSvc.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.userSvc.DeleteUser(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

type AddClientRequest struct {
	InboundID int64  `json:"inbound_id" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Group     string `json:"group,omitempty"`
	DataLimit int64  `json:"total_gb,omitempty"`
	ExpiryTime int64 `json:"expiry_time,omitempty"`
}

func (h *UserHandler) AddClient(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req AddClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "inbound_id and email required"})
		return
	}

	adminID := getAdminID(c)

	// RBAC: Check create permission
	if !rbac.CanCreateClient(adminID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "client create permission denied"})
		return
	}

	// RBAC: Validate group access
	if req.Group != "" {
		if err := rbac.ValidateGroupAccess(adminID, req.Group); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
	}

	// RBAC: Validate inbound access
	if err := rbac.ValidateInboundAccess(adminID, req.InboundID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// RBAC: Validate client limits (data limit, expiry)
	if err := rbac.ValidateClientCreate(adminID, req.DataLimit, req.ExpiryTime); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// RBAC: Verify the admin has access to this user
	user, err := h.userSvc.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	scope := rbac.GetClientAccessScopeForAdmin(adminID, "create")
	if scope.Mode != rbac.ClientAccessAll && user.AdminID > 0 && user.AdminID != adminID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this user"})
		return
	}

	// Create client with UUID and ownership set upfront (avoid race condition)
	client := &database.Client{
		ID:           uuid.New().String(),
		UserID:       userID,
		InboundID:    req.InboundID,
		Email:        req.Email,
		Group:        req.Group,
		Enable:       true,
		OwnerAdminID: adminID,
		TotalGB:      req.DataLimit,
		ExpiryTime:   req.ExpiryTime,
	}
	if err := database.CreateClient(client); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, client)
}

func (h *UserHandler) ListClients(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Get all clients for this user
	clients, err := h.userSvc.ListClients(userID, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adminID := getAdminID(c)

	// RBAC: Filter clients by access scope
	scope := rbac.GetClientAccessScopeForAdmin(adminID, "view")
	if scope.Mode == rbac.ClientAccessOwn {
		filtered := make([]database.Client, 0, len(clients))
		for _, cl := range clients {
			if cl.OwnerAdminID == adminID {
				filtered = append(filtered, cl)
			}
		}
		clients = filtered
	} else if scope.Mode == rbac.ClientAccessNone {
		clients = []database.Client{}
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients, "total": len(clients)})
}

func (h *UserHandler) DeleteClient(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client id required"})
		return
	}

	// RBAC: Check client exists and is accessible
	client, err := h.userSvc.GetClient(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	adminID := getAdminID(c)
	scope := rbac.GetClientAccessScopeForAdmin(adminID, "delete")
	if !scope.ClientAllowed(client) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this client"})
		return
	}

	if err := h.userSvc.DeleteClient(clientID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client deleted"})
}

func (h *UserHandler) ResetTraffic(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// RBAC: Check reset permission
	adminID := getAdminID(c)
	role, roleErr := database.GetAdminRole(adminID)
	if roleErr != nil || (!role.OwnerRole && !rbac.CheckPermission(role, "users", "resetUsage")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reset traffic permission denied"})
		return
	}

	_ = userID // Will implement in future
	c.JSON(http.StatusOK, gin.H{"message": "traffic reset initiated"})
}

// ─── User Portal Endpoints ─────────────────────────────────────────

// ListOwnClients returns clients for the currently authenticated admin.
func (h *UserHandler) ListOwnClients(c *gin.Context) {
	adminID := getAdminID(c)
	if adminID == 0 {
		c.JSON(http.StatusOK, gin.H{"clients": []any{}, "total": 0})
		return
	}

	// RBAC: filter by own scope
	scope := rbac.GetClientAccessScopeForAdmin(adminID, "view")
	var clients []database.Client
	if scope.Mode != rbac.ClientAccessNone {
		var err error
		clients, err = h.userSvc.ListClients(adminID, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Filter by ownership
		if scope.Mode == rbac.ClientAccessOwn {
			filtered := make([]database.Client, 0, len(clients))
			for _, cl := range clients {
				if cl.OwnerAdminID == adminID {
					filtered = append(filtered, cl)
				}
			}
			clients = filtered
		}
	}
	if clients == nil {
		clients = []database.Client{}
	}
	c.JSON(http.StatusOK, gin.H{"clients": clients, "total": len(clients)})
}

// GetClientDetail returns detailed info about a specific client.
func (h *UserHandler) GetClientDetail(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client id required"})
		return
	}

	client, err := h.userSvc.GetClient(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	// RBAC: Check client access
	adminID := getAdminID(c)
	scope := rbac.GetClientAccessScopeForAdmin(adminID, "view")
	if !scope.ClientAllowed(client) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	user, err := h.userSvc.GetUser(client.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	usage := user.TrafficUp + user.TrafficDown
	usagePercent := 0.0
	if user.DataLimit > 0 {
		usagePercent = float64(usage) / float64(user.DataLimit) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"client":        client,
		"email":         client.Email,
		"inbound_id":    client.InboundID,
		"group":         client.Group,
		"enable":        client.Enable,
		"sub_id":        client.SubID,
		"traffic_up":    user.TrafficUp,
		"traffic_down":  user.TrafficDown,
		"traffic_total": usage,
		"data_limit":    user.DataLimit,
		"usage_percent": usagePercent,
		"expiry_time":   user.ExpiryTime,
		"user_status":   user.Status,
	})
}

// GetOwnTraffic returns traffic stats for the authenticated user.
func (h *UserHandler) GetOwnTraffic(c *gin.Context) {
	adminID := getAdminID(c)

	user, err := h.userSvc.GetUser(adminID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"traffic_up":   0,
			"traffic_down": 0,
			"data_limit":   0,
			"usage":        0,
			"remaining":    0,
			"expiry_time":  0,
			"status":       "unknown",
		})
		return
	}

	usage := user.TrafficUp + user.TrafficDown
	remaining := user.DataLimit - usage
	if remaining < 0 {
		remaining = 0
	}
	usagePercent := 0.0
	if user.DataLimit > 0 {
		usagePercent = float64(usage) / float64(user.DataLimit) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"traffic_up":    user.TrafficUp,
		"traffic_down":  user.TrafficDown,
		"data_limit":    user.DataLimit,
		"usage":         usage,
		"remaining":     remaining,
		"usage_percent": usagePercent,
		"expiry_time":   user.ExpiryTime,
		"status":        user.Status,
	})
}
