package database

import (
	"fmt"
	"time"

	"vortexuipro/internal/auth"
)

// ─── Admin Repository ────────────────────────────────────────────────

func CreateAdmin(a *Admin) error {
	a.CreatedAt = time.Now().UnixMilli()
	a.UpdatedAt = a.CreatedAt
	return DB.Create(a).Error
}

func GetAdminByID(id int64) (*Admin, error) {
	var a Admin
	if err := DB.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func GetAdminByUsername(username string) (*Admin, error) {
	var a Admin
	if err := DB.Where("username = ?", username).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func ListAdmins() ([]Admin, error) {
	var list []Admin
	return list, DB.Find(&list).Error
}

func UpdateAdmin(a *Admin) error {
	a.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(a).Error
}

func DeleteAdmin(id int64) error {
	return DB.Delete(&Admin{}, id).Error
}

// GetAdminRole returns the role associated with an admin.
func GetAdminRole(adminID int64) (*AdminRole, error) {
	var admin Admin
	if err := DB.First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	if admin.GetRoleID() <= 0 {
		return nil, fmt.Errorf("admin has no role assigned")
	}
	var role AdminRole
	if err := DB.First(&role, admin.GetRoleID()).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// ─── User Repository ─────────────────────────────────────────────────

func CreateUser(u *User) error {
	u.CreatedAt = time.Now().UnixMilli()
	u.UpdatedAt = u.CreatedAt
	return DB.Create(u).Error
}

func GetUserByID(id int64) (*User, error) {
	var u User
	if err := DB.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByUsername(username string) (*User, error) {
	var u User
	if err := DB.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func ListUsers(adminID int64) ([]User, error) {
	var list []User
	q := DB.Model(&User{})
	if adminID > 0 {
		q = q.Where("admin_id = ?", adminID)
	}
	return list, q.Find(&list).Error
}

func UpdateUser(u *User) error {
	u.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(u).Error
}

func DeleteUser(id int64) error {
	// Delete related clients first
	DB.Where("user_id = ?", id).Delete(&Client{})
	return DB.Delete(&User{}, id).Error
}

// ─── Inbound Repository ──────────────────────────────────────────────

func CreateInbound(ib *Inbound) error {
	ib.CreatedAt = time.Now().UnixMilli()
	ib.UpdatedAt = ib.CreatedAt
	return DB.Create(ib).Error
}

func GetInboundByID(id int64) (*Inbound, error) {
	var ib Inbound
	if err := DB.First(&ib, id).Error; err != nil {
		return nil, err
	}
	return &ib, nil
}

func GetInboundByTag(tag string) (*Inbound, error) {
	var ib Inbound
	if err := DB.Where("tag = ?", tag).First(&ib).Error; err != nil {
		return nil, err
	}
	return &ib, nil
}

func ListInbounds(userID, nodeID int64) ([]Inbound, error) {
	var list []Inbound
	q := DB.Model(&Inbound{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if nodeID > 0 {
		q = q.Where("node_id = ?", nodeID)
	}
	return list, q.Order("id asc").Find(&list).Error
}

func UpdateInbound(ib *Inbound) error {
	ib.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(ib).Error
}

func DeleteInbound(id int64) error {
	return DB.Delete(&Inbound{}, id).Error
}

// ─── Client Repository ───────────────────────────────────────────────

func CreateClient(c *Client) error {
	now := time.Now().UnixMilli()
	c.CreatedAt = now
	c.UpdatedAt = now
	return DB.Create(c).Error
}

func GetClientByID(id string) (*Client, error) {
	var c Client
	if err := DB.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func GetClientByEmail(email string) (*Client, error) {
	var c Client
	if err := DB.Where("email = ?", email).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func ListClients(userID, inboundID int64) ([]Client, error) {
	var list []Client
	q := DB.Model(&Client{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if inboundID > 0 {
		q = q.Where("inbound_id = ?", inboundID)
	}
	return list, q.Order("email asc").Find(&list).Error
}

func UpdateClient(c *Client) error {
	c.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(c).Error
}

func DeleteClient(id string) error {
	return DB.Delete(&Client{}, "id = ?", id).Error
}

// ─── Node Repository ─────────────────────────────────────────────────

func CreateNode(n *Node) error {
	n.CreatedAt = time.Now().UnixMilli()
	n.UpdatedAt = n.CreatedAt
	return DB.Create(n).Error
}

func GetNodeByID(id int64) (*Node, error) {
	var n Node
	if err := DB.First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func GetNodeByName(name string) (*Node, error) {
	var n Node
	if err := DB.Where("name = ?", name).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func ListNodes() ([]Node, error) {
	var list []Node
	return list, DB.Order("name asc").Find(&list).Error
}

func UpdateNode(n *Node) error {
	n.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(n).Error
}

func DeleteNode(id int64) error {
	return DB.Delete(&Node{}, id).Error
}

// ─── Setting Repository ──────────────────────────────────────────────

func GetSetting(key string) (string, error) {
	var s Setting
	if err := DB.Where("\"key\" = ?", key).First(&s).Error; err != nil {
		return "", err
	}
	return s.Value, nil
}

func SetSetting(key, value string) error {
	var s Setting
	result := DB.Where("\"key\" = ?", key).First(&s)
	if result.Error != nil {
		return DB.Create(&Setting{Key: key, Value: value}).Error
	}
	s.Value = value
	return DB.Save(&s).Error
}

func DeleteSetting(key string) error {
	return DB.Where("\"key\" = ?", key).Delete(&Setting{}).Error
}

// ─── Ticket Repository ───────────────────────────────────────────────

func CreateTicket(t *Ticket) error {
	t.CreatedAt = time.Now().UnixMilli()
	t.UpdatedAt = t.CreatedAt
	return DB.Create(t).Error
}

func GetTicketByID(id int64) (*Ticket, error) {
	var t Ticket
	if err := DB.Preload("Replies").First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func ListTickets(userID int64) ([]Ticket, error) {
	var list []Ticket
	q := DB.Model(&Ticket{}).Preload("Replies")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	return list, q.Order("updated_at desc").Find(&list).Error
}

func UpdateTicket(t *Ticket) error {
	t.UpdatedAt = time.Now().UnixMilli()
	return DB.Save(t).Error
}

func DeleteTicket(id int64) error {
	DB.Where("ticket_id = ?", id).Delete(&TicketReply{})
	return DB.Delete(&Ticket{}, id).Error
}

// ─── Ticket Reply Repository ─────────────────────────────────────────

func CreateTicketReply(r *TicketReply) error {
	r.CreatedAt = time.Now().UnixMilli()
	if err := DB.Create(r).Error; err != nil {
		return err
	}
	// Update parent ticket's updated_at
	DB.Model(&Ticket{}).Where("id = ?", r.TicketID).Update("updated_at", r.CreatedAt)
	return nil
}

// ─── Outbound Repository ────────────────────────────────────────────────

func CreateOutbound(ob *Outbound) error {
	return DB.Create(ob).Error
}

func GetOutboundByID(id int64) (*Outbound, error) {
	var ob Outbound
	if err := DB.First(&ob, id).Error; err != nil {
		return nil, err
	}
	return &ob, nil
}

func GetOutboundByTag(tag string) (*Outbound, error) {
	var ob Outbound
	if err := DB.Where("tag = ?", tag).First(&ob).Error; err != nil {
		return nil, err
	}
	return &ob, nil
}

func ListOutbounds(nodeID int64) ([]Outbound, error) {
	var list []Outbound
	q := DB.Model(&Outbound{})
	if nodeID > 0 {
		q = q.Where("node_id = ?", nodeID)
	}
	return list, q.Order("id asc").Find(&list).Error
}

func UpdateOutbound(ob *Outbound) error {
	return DB.Save(ob).Error
}

func DeleteOutbound(id int64) error {
	return DB.Delete(&Outbound{}, id).Error
}

// ─── Notification Channel Repository ─────────────────────────────────

func CreateNotificationChannel(nc *NotificationChannel) error {
	nc.CreatedAt = time.Now().UnixMilli()
	return DB.Create(nc).Error
}

func ListNotificationChannels() ([]NotificationChannel, error) {
	var list []NotificationChannel
	return list, DB.Order("name asc").Find(&list).Error
}

func UpdateNotificationChannel(nc *NotificationChannel) error {
	return DB.Save(nc).Error
}

func DeleteNotificationChannel(id int64) error {
	return DB.Delete(&NotificationChannel{}, id).Error
}

// ─── Security Event Repository ───────────────────────────────────────

func CreateSecurityEvent(e *SecurityEvent) error {
	e.CreatedAt = time.Now().UnixMilli()
	return DB.Create(e).Error
}

func ListSecurityEvents(limit int) ([]SecurityEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	var list []SecurityEvent
	return list, DB.Order("created_at desc").Limit(limit).Find(&list).Error
}

// ─── Plan Repository ─────────────────────────────────────────────────

func CreatePlan(p *Plan) error {
	p.CreatedAt = time.Now().UnixMilli()
	return DB.Create(p).Error
}

func GetOrderByID(id int64) (*Order, error) {
	var o Order
	if err := DB.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func ListPlans() ([]Plan, error) {
	var list []Plan
	return list, DB.Order("price asc").Find(&list).Error
}

func GetPlanByID(id int64) (*Plan, error) {
	var p Plan
	if err := DB.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func DeletePlan(id int64) error {
	return DB.Delete(&Plan{}, id).Error
}

// ─── Order Repository ────────────────────────────────────────────────

func CreateOrder(o *Order) error {
	o.CreatedAt = time.Now().UnixMilli()
	return DB.Create(o).Error
}

func ListOrders(userID int64) ([]Order, error) {
	var list []Order
	q := DB.Model(&Order{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	return list, q.Order("created_at desc").Find(&list).Error
}

func UpdateOrder(o *Order) error {
	return DB.Save(o).Error
}

// ─── Transaction Repository ──────────────────────────────────────────

func CreateTransaction(t *Transaction) error {
	t.CreatedAt = time.Now().UnixMilli()
	return DB.Create(t).Error
}

func ListTransactions(userID int64, limit int) ([]Transaction, error) {
	if limit <= 0 {
		limit = 50
	}
	var list []Transaction
	q := DB.Model(&Transaction{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	return list, q.Order("created_at desc").Limit(limit).Find(&list).Error
}

// ─── Client By Sub ID ────────────────────────────────────────────────

func ListClientsBySubID(subID string) ([]Client, error) {
	var list []Client
	return list, DB.Where("sub_id = ?", subID).Order("email asc").Find(&list).Error
}

// ─── Subscription Profile Repository (Heimdall feature) ──────────────

func CreateSubscriptionProfile(sp *SubscriptionProfile) error {
	sp.CreatedAt = time.Now().UnixMilli()
	return DB.Create(sp).Error
}

func ListSubscriptionProfiles(inboundID int64) ([]SubscriptionProfile, error) {
	var list []SubscriptionProfile
	q := DB.Model(&SubscriptionProfile{})
	if inboundID > 0 {
		q = q.Where("inbound_id = ?", inboundID)
	}
	return list, q.Order("id asc").Find(&list).Error
}

func DeleteSubscriptionProfile(id int64) error {
	return DB.Delete(&SubscriptionProfile{}, id).Error
}

// ─── Utility ─────────────────────────────────────────────────────────

// SeedDefaults creates initial admin if no admin exists.
func SeedDefaults() error {
	var count int64
	DB.Model(&Admin{}).Count(&count)
	if count == 0 {
		// Generate proper Argon2id hash for "admin123"
		hash, err := auth.HashPassword("admin123")
		if err != nil {
			return fmt.Errorf("hash default password: %w", err)
		}
		admin := Admin{
			Username:     "admin",
			PasswordHash: hash,
			Role:         "super_admin",
		}
		if err := CreateAdmin(&admin); err != nil {
			return fmt.Errorf("seed default admin: %w", err)
		}
		fmt.Println("Default admin created: admin / admin123")
	}
	return nil
}
