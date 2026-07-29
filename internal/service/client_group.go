package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"vortexuipro/internal/database"
	"vortexuipro/internal/events"
)

// ClientGroup represents a named group of clients.
type ClientGroup struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"uniqueIndex;not null;size:128" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	AdminID     int64  `gorm:"index;default:0" json:"admin_id,omitempty"`
	MemberCount int    `gorm:"default:0" json:"member_count"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (ClientGroup) TableName() string { return "client_groups" }

// ClientGroupMember represents a client's membership in a group.
type ClientGroupMember struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID   int64  `gorm:"uniqueIndex:idx_group_client;not null" json:"group_id"`
	ClientID  string `gorm:"uniqueIndex:idx_group_client;size:64;not null" json:"client_id"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
}

func (ClientGroupMember) TableName() string { return "client_group_members" }

// GroupedClient is a client with group membership info.
type GroupedClient struct {
	database.Client
	Groups []string `json:"groups,omitempty"`
}

// ClientGroupService manages client grouping and bulk operations.
type ClientGroupService struct {
	eventBus events.Publisher
}

// NewClientGroupService creates a new client group service.
func NewClientGroupService(bus events.Publisher) *ClientGroupService {
	if bus == nil {
		bus = events.Nop{}
	}
	return &ClientGroupService{eventBus: bus}
}

// ─── Group CRUD ──────────────────────────────────────────────────────

// CreateGroup creates a new client group.
func (s *ClientGroupService) CreateGroup(ctx context.Context, name, description string, adminID int64) (*ClientGroup, error) {
	g := &ClientGroup{
		Name:        name,
		Description: description,
		AdminID:     adminID,
		CreatedAt:   time.Now().UnixMilli(),
	}
	if err := database.DB.Create(g).Error; err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return g, nil
}

// GetGroup returns a group by ID.
func (s *ClientGroupService) GetGroup(id int64) (*ClientGroup, error) {
	var g ClientGroup
	if err := database.DB.First(&g, id).Error; err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}
	// Count members
	var count int64
	database.DB.Model(&ClientGroupMember{}).Where("group_id = ?", id).Count(&count)
	g.MemberCount = int(count)
	return &g, nil
}

// ListGroups returns all groups.
func (s *ClientGroupService) ListGroups(adminID int64) ([]ClientGroup, error) {
	var list []ClientGroup
	q := database.DB.Model(&ClientGroup{})
	if adminID > 0 {
		q = q.Where("admin_id = ?", adminID)
	}
	if err := q.Order("name asc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	// Get member counts
	for i := range list {
		var count int64
		database.DB.Model(&ClientGroupMember{}).Where("group_id = ?", list[i].ID).Count(&count)
		list[i].MemberCount = int(count)
	}
	return list, nil
}

// UpdateGroup updates a group's name and description.
func (s *ClientGroupService) UpdateGroup(ctx context.Context, id int64, name, description string) error {
	return database.DB.Model(&ClientGroup{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":       name,
		"description": description,
		"updated_at": time.Now().UnixMilli(),
	}).Error
}

// DeleteGroup deletes a group and all its memberships.
func (s *ClientGroupService) DeleteGroup(ctx context.Context, id int64) error {
	database.DB.Where("group_id = ?", id).Delete(&ClientGroupMember{})
	return database.DB.Delete(&ClientGroup{}, id).Error
}

// ─── Group Membership ────────────────────────────────────────────────

// AddClientToGroup adds a client to a group.
func (s *ClientGroupService) AddClientToGroup(ctx context.Context, groupID int64, clientID string) error {
	member := &ClientGroupMember{
		GroupID:   groupID,
		ClientID:  clientID,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := database.DB.Create(member).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
			return nil // already a member
		}
		return fmt.Errorf("add client to group: %w", err)
	}
	return nil
}

// RemoveClientFromGroup removes a client from a group.
func (s *ClientGroupService) RemoveClientFromGroup(ctx context.Context, groupID int64, clientID string) error {
	return database.DB.Where("group_id = ? AND client_id = ?", groupID, clientID).Delete(&ClientGroupMember{}).Error
}

// GetGroupClients returns all clients in a group.
func (s *ClientGroupService) GetGroupClients(groupID int64) ([]database.Client, error) {
	var clients []database.Client
	if err := database.DB.Raw(`
		SELECT c.* FROM clients c
		INNER JOIN client_group_members cgm ON cgm.client_id = c.id
		WHERE cgm.group_id = ?
		ORDER BY c.email ASC
	`, groupID).Scan(&clients).Error; err != nil {
		return nil, fmt.Errorf("get group clients: %w", err)
	}
	return clients, nil
}

// GetClientGroups returns all groups a client belongs to.
func (s *ClientGroupService) GetClientGroups(clientID string) ([]ClientGroup, error) {
	var groups []ClientGroup
	if err := database.DB.Raw(`
		SELECT cg.* FROM client_groups cg
		INNER JOIN client_group_members cgm ON cgm.group_id = cg.id
		WHERE cgm.client_id = ?
		ORDER BY cg.name ASC
	`, clientID).Scan(&groups).Error; err != nil {
		return nil, fmt.Errorf("get client groups: %w", err)
	}
	return groups, nil
}

// ─── Bulk Operations ────────────────────────────────────────────────

// BulkAddClientsToGroup adds multiple clients to a group at once.
func (s *ClientGroupService) BulkAddClientsToGroup(ctx context.Context, groupID int64, clientIDs []string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, clientID := range clientIDs {
			member := &ClientGroupMember{
				GroupID:   groupID,
				ClientID:  clientID,
				CreatedAt: time.Now().UnixMilli(),
			}
			if err := tx.Create(member).Error; err != nil {
				if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
					continue
				}
				return err
			}
		}
		return nil
	})
}

// BulkRemoveClientsFromGroup removes multiple clients from a group at once.
func (s *ClientGroupService) BulkRemoveClientsFromGroup(ctx context.Context, groupID int64, clientIDs []string) error {
	return database.DB.Where("group_id = ? AND client_id IN ?", groupID, clientIDs).Delete(&ClientGroupMember{}).Error
}

// BulkAttachGroupToInbounds attaches a group's clients to inbounds.
func (s *ClientGroupService) BulkAttachGroupToInbounds(ctx context.Context, groupID int64, inboundIDs []int64) error {
	clients, err := s.GetGroupClients(groupID)
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, client := range clients {
			for _, inboundID := range inboundIDs {
				// Check if already exists
				var count int64
				tx.Model(&database.Client{}).Where("id = ? AND inbound_id = ?", client.ID, inboundID).Count(&count)
				if count > 0 {
					continue
				}
				// Clone client to new inbound
				newClient := client
				newClient.InboundID = inboundID
				newClient.ID = "" // let DB auto-generate new ID or use UUID
				if err := tx.Create(&newClient).Error; err != nil {
					continue // skip duplicates
				}
			}
		}
		return nil
	})
}

// GetClientsWithGroups returns all clients with their group memberships.
func (s *ClientGroupService) GetClientsWithGroups(adminID int64) ([]GroupedClient, error) {
	var clients []database.Client
	q := database.DB.Model(&database.Client{})
	if adminID > 0 {
		q = q.Where("admin_id = ?", adminID)
	}
	if err := q.Order("email ASC").Find(&clients).Error; err != nil {
		return nil, err
	}

	var result []GroupedClient
	for _, c := range clients {
		groups, _ := s.GetClientGroups(c.ID)
		groupNames := make([]string, len(groups))
		for i, g := range groups {
			groupNames[i] = g.Name
		}
		result = append(result, GroupedClient{
			Client: c,
			Groups: groupNames,
		})
	}
	return result, nil
}
