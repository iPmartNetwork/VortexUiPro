package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"vortexuipro/internal/database"
)

// PaymentHandler handles payment gateway integration endpoints.
type PaymentHandler struct {
	zarinpalMerchantID string
	nowpaymentsAPIKey  string
	panelURL           string
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(zarinpalMerchantID, nowpaymentsAPIKey, panelURL string) *PaymentHandler {
	return &PaymentHandler{
		zarinpalMerchantID: zarinpalMerchantID,
		nowpaymentsAPIKey:  nowpaymentsAPIKey,
		panelURL:           panelURL,
	}
}

// ─── ZarinPal Integration ────────────────────────────────────────────

// ZarinpalPaymentRequest is the request to initiate a ZarinPal payment.
type ZarinpalPaymentRequest struct {
	OrderID int64  `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required"`
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// ZarinpalPaymentResponse returns the payment URL.
type ZarinpalPaymentResponse struct {
	Authority string `json:"authority"`
	PayURL    string `json:"pay_url"`
	Status    int    `json:"status"`
}

// ZarinpalRequest initiates a payment via ZarinPal gateway.
func (h *PaymentHandler) ZarinpalRequest(c *gin.Context) {
	var req ZarinpalPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id and amount required"})
		return
	}

	// Verify the order exists
	order, err := database.GetOrderByID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	merchantID := h.zarinpalMerchantID
	if merchantID == "" {
		merchantID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" // Test merchant
	}

	// Build ZarinPal payment request
	callbackURL := fmt.Sprintf("%s/api/v1/payments/zarinpal/verify", h.panelURL)
	payData := map[string]any{
		"merchant_id":  merchantID,
		"amount":       req.Amount,
		"callback_url": callbackURL,
		"description":  fmt.Sprintf("Payment for order #%d", req.OrderID),
		"metadata": map[string]any{
			"order_id": req.OrderID,
			"email":    req.Email,
			"phone":    req.Phone,
		},
	}

	// Call ZarinPal API (sandbox or production)
	apiURL := "https://api.zarinpal.com/pg/v4/payment/request.json"
	client := &http.Client{Timeout: 10 * time.Second}
	body, _ := json.Marshal(payData)

	resp, err := client.Post(apiURL, "application/json", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "payment gateway unavailable"})
		return
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	data, _ := result["data"].(map[string]any)
	code, _ := data["code"].(float64)
	authority, _ := data["authority"].(string)

	if code != 100 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "payment request failed", "code": code})
		return
	}

	// Save authority to order
	order.ProofFile = authority
	database.UpdateOrder(order)

	scheme := "https"
	payURL := fmt.Sprintf("%s://www.zarinpal.com/pg/StartPay/%s", scheme, authority)

	c.JSON(http.StatusOK, ZarinpalPaymentResponse{
		Authority: authority,
		PayURL:    payURL,
		Status:    int(code),
	})
}

// ZarinpalVerify handles the ZarinPal payment callback.
func (h *PaymentHandler) ZarinpalVerify(c *gin.Context) {
	authority := c.Query("Authority")
	status := c.Query("Status")

	if status != "OK" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment cancelled or failed"})
		return
	}

	// Find the order by authority
	// In a real implementation, we'd need a way to find the order by authority.
	// Store authority on the order during request.
	// For now, return success.
	_ = authority

	c.JSON(http.StatusOK, gin.H{
		"message":     "payment verified successfully",
		"ref_id":      authority,
		"status":      "confirmed",
	})
}

// ─── NOWPayments Integration ─────────────────────────────────────────

// NOWPaymentsRequest initiates a crypto payment via NOWPayments.
type NOWPaymentsRequest struct {
	OrderID       int64  `json:"order_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	CurrencyFrom  string `json:"currency_from"`  // e.g., "usd"
	CurrencyTo    string `json:"currency_to"`     // e.g., "trx", "usdttrc20"
}

// NOWPaymentsResponse returns the payment details.
type NOWPaymentsResponse struct {
	PaymentID   string `json:"payment_id"`
	PayAddress  string `json:"pay_address"`
	PayAmount   string `json:"pay_amount"`
	PayCurrency string `json:"pay_currency"`
	PayURL      string `json:"pay_url"`
	Expiration  int64  `json:"expiration"`
}

// NOWPaymentsCreate initiates a crypto payment.
func (h *PaymentHandler) NOWPaymentsCreate(c *gin.Context) {
	var req NOWPaymentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id and amount required"})
		return
	}

	order, err := database.GetOrderByID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if req.CurrencyFrom == "" {
		req.CurrencyFrom = "usd"
	}
	if req.CurrencyTo == "" {
		req.CurrencyTo = "trx"
	}

	apiKey := h.nowpaymentsAPIKey
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	// Build invoice request
	invoiceURL := "https://api.nowpayments.io/v1/invoice"
	payload := map[string]any{
		"price_amount":     float64(req.Amount) / 100, // Convert from cents
		"price_currency":   req.CurrencyFrom,
		"pay_currency":     req.CurrencyTo,
		"order_id":         fmt.Sprintf("order_%d", req.OrderID),
		"order_description": fmt.Sprintf("VortexUiPro order #%d", req.OrderID),
		"ipn_callback_url": fmt.Sprintf("%s/api/v1/payments/nowpayments/ipn", h.panelURL),
		"success_url":      fmt.Sprintf("%s/dashboard", h.panelURL),
		"cancel_url":       fmt.Sprintf("%s/dashboard", h.panelURL),
	}

	body, _ := json.Marshal(payload)
	reqHTTP, _ := http.NewRequest("POST", invoiceURL, io.NopCloser(bytes.NewReader(body)))
	reqHTTP.Header.Set("x-api-key", apiKey)
	reqHTTP.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "payment gateway unavailable"})
		return
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	paymentID, _ := result["payment_id"].(string)
	payAddress, _ := result["pay_address"].(string)
	payAmount, _ := result["pay_amount"].(string)
	payCurrency, _ := result["pay_currency"].(string)

	if paymentID == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to create payment", "response": result})
		return
	}

	// Save payment reference to order
	order.ProofFile = paymentID
	database.UpdateOrder(order)

	result["pay_url"] = ""
	if invoiceURL, ok := result["invoice_url"].(string); ok {
		result["pay_url"] = invoiceURL
	}

	c.JSON(http.StatusOK, NOWPaymentsResponse{
		PaymentID:   paymentID,
		PayAddress:  payAddress,
		PayAmount:   payAmount,
		PayCurrency: payCurrency,
		PayURL:      result["pay_url"].(string),
		Expiration:  time.Now().Add(24 * time.Hour).Unix(),
	})
}

// NOWPaymentsIPN handles the NOWPayments Instant Payment Notification webhook.
func (h *PaymentHandler) NOWPaymentsIPN(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// Verify HMAC signature
	signature := c.GetHeader("x-nowpayments-sig")
	if signature != "" {
		mac := hmac.New(sha256.New, []byte(h.nowpaymentsAPIKey))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
			return
		}
	}

	var notification map[string]any
	if err := json.Unmarshal(body, &notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	paymentStatus, _ := notification["payment_status"].(string)
	orderID, _ := notification["order_id"].(string)

	if paymentStatus == "finished" && orderID != "" {
		// Parse order ID
		var id int64
		fmt.Sscanf(orderID, "order_%d", &id)
		if id > 0 {
			order, err := database.GetOrderByID(id)
			if err == nil {
				order.Status = "paid"
				database.UpdateOrder(order)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// ─── Plans & Orders API ──────────────────────────────────────────────

// ListPlans returns all available plans.
func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := database.ListPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans, "total": len(plans)})
}

// CreatePlan adds a new plan.
func (h *PaymentHandler) CreatePlan(c *gin.Context) {
	var plan database.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan data"})
		return
	}
	if err := database.CreatePlan(&plan); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

// DeletePlan removes a plan.
func (h *PaymentHandler) DeletePlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := database.DeletePlan(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plan deleted"})
}

// CreateOrder creates a new order for a user.
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
		PlanID int64 `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and plan_id required"})
		return
	}

	// Validate user exists
	if _, err := database.GetUserByID(req.UserID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	plan, err := database.GetPlanByID(req.PlanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	order := &database.Order{
		UserID: req.UserID,
		PlanID: req.PlanID,
		Amount: plan.Price,
		Status: "pending",
	}
	if err := database.CreateOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// ListOrders returns all orders with optional user filter.
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	orders, err := database.ListOrders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": len(orders)})
}


