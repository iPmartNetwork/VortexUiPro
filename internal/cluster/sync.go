package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"vortexuipro/internal/database"
	"gorm.io/gorm/clause"
)

// ─── Sync Types ──────────────────────────────────────────────────────

// SyncDataType represents the type of data being synced.
type SyncDataType string

const (
	SyncUsers    SyncDataType = "users"
	SyncTraffic  SyncDataType = "traffic"
	SyncInbounds SyncDataType = "inbounds"
	SyncBans     SyncDataType = "bans"
	SyncClients  SyncDataType = "clients"
	SyncSettings SyncDataType = "settings"
)

// SyncOperation is the type of sync operation.
type SyncOperation string

const (
	OpCreate SyncOperation = "create"
	OpUpdate SyncOperation = "update"
	OpDelete SyncOperation = "delete"
)

// SyncPayload carries the actual sync data between nodes.
type SyncPayload struct {
	DataType    SyncDataType    `json:"data_type"`
	Operation   SyncOperation  `json:"operation"`
	EntityID    string          `json:"entity_id"`
	Data        json.RawMessage `json:"data,omitempty"`
	Timestamp   int64           `json:"timestamp"`
	SourceID    int64           `json:"source_id"`
	SourceName  string          `json:"source_name"`
}

// SyncAckPayload is sent in response to a sync.
type SyncAckPayload struct {
	DataType  SyncDataType `json:"data_type"`
	EntityID  string       `json:"entity_id"`
	Success   bool         `json:"success"`
	Error     string       `json:"error,omitempty"`
}

// FullSyncPayload carries all data for a full sync request.
type FullSyncPayload struct {
	Users    []database.User    `json:"users,omitempty"`
	Clients  []database.Client  `json:"clients,omitempty"`
	Inbounds []database.Inbound `json:"inbounds,omitempty"`
	Settings []database.Setting `json:"settings,omitempty"`
}

// ─── Sync Service ────────────────────────────────────────────────────

// SyncService handles data synchronization between cluster nodes.
type SyncService struct {
	mu sync.RWMutex

	nodeID        int64
	nodeName      string
	getPeers       func() []ClusterPeer
	isLeader       func() bool
	getLeaderAddr  func() string

	// Conflict resolution
	resolver *ConflictResolver

	// Sync queue (buffer changes when leader is unavailable)
	syncQueue []SyncPayload
	queueMu   sync.Mutex

	stopCh chan struct{}
}

// NewSyncService creates a new cluster sync service.
func NewSyncService(nodeID int64, nodeName string, getPeers func() []ClusterPeer, isLeader func() bool) *SyncService {
	return &SyncService{
		nodeID:   nodeID,
		nodeName: nodeName,
		getPeers: getPeers,
		isLeader: isLeader,
		resolver: NewConflictResolver(),
		syncQueue: make([]SyncPayload, 0, 100),
		stopCh:    make(chan struct{}),
	}
}

// NewSyncServiceWithResolver creates a sync service with a shared conflict resolver.
func NewSyncServiceWithResolver(nodeID int64, nodeName string, getPeers func() []ClusterPeer, isLeader func() bool, getLeaderAddr func() string, resolver *ConflictResolver) *SyncService {
	return &SyncService{
		nodeID:        nodeID,
		nodeName:      nodeName,
		getPeers:      getPeers,
		isLeader:      isLeader,
		getLeaderAddr: getLeaderAddr,
		resolver:      resolver,
		syncQueue:     make([]SyncPayload, 0, 100),
		stopCh:        make(chan struct{}),
	}
}

// Start begins the sync loop.
func (ss *SyncService) Start() {
	go ss.flushLoop()
	log.Println("Cluster sync service started")
}

// Stop stops the sync service.
func (ss *SyncService) Stop() {
	close(ss.stopCh)
}

func (ss *SyncService) flushLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ss.flushQueue()
		case <-ss.stopCh:
			ss.flushQueue() // flush remaining on stop
			return
		}
	}
}

func (ss *SyncService) flushQueue() {
	ss.queueMu.Lock()
	if len(ss.syncQueue) == 0 {
		ss.queueMu.Unlock()
		return
	}
	queue := make([]SyncPayload, len(ss.syncQueue))
	copy(queue, ss.syncQueue)
	ss.syncQueue = ss.syncQueue[:0]
	ss.queueMu.Unlock()

	for _, payload := range queue {
		ss.BroadcastSync(payload)
	}
}

// QueueSync queues a sync operation to be broadcast later.
func (ss *SyncService) QueueSync(dataType SyncDataType, operation SyncOperation, entityID string, data any) {
	raw, _ := json.Marshal(data)
	payload := SyncPayload{
		DataType:   dataType,
		Operation:  operation,
		EntityID:   entityID,
		Data:       raw,
		Timestamp:  time.Now().UnixMilli(),
		SourceID:   ss.nodeID,
		SourceName: ss.nodeName,
	}

	if ss.isLeader() {
		ss.BroadcastSync(payload)
	} else {
		ss.queueMu.Lock()
		ss.syncQueue = append(ss.syncQueue, payload)
		if len(ss.syncQueue) > 1000 {
			ss.syncQueue = ss.syncQueue[1:] // drop oldest
		}
		ss.queueMu.Unlock()
		// Also try to send to leader immediately
		go ss.sendToLeader(payload)
	}
}

// BroadcastSync sends a sync payload to all cluster peers.
func (ss *SyncService) BroadcastSync(payload SyncPayload) {
	payloadBytes, _ := json.Marshal(payload)
	msg := MeshMessage{
		Type:      MsgSyncData,
		SenderID:  ss.nodeID,
		SenderName: ss.nodeName,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payloadBytes,
	}

	peers := ss.getPeers()
	for _, peer := range peers {
		if peer.ID == ss.nodeID || !peer.Online {
			continue
		}
		client := NewPeerClient(peer.Address)
		if _, err := client.Send(msg); err != nil {
			log.Printf("Sync to %s (%s) failed: %v", peer.Name, peer.Address, err)
		}
	}

	// Log sync event to database
	if database.DB != nil {
		event := database.SyncEvent{
			Type:     string(payload.DataType) + "_sync",
			SourceID: ss.nodeID,
			EntityID: payload.EntityID,
			Status:   "broadcast",
			Detail:   fmt.Sprintf("%s %s", payload.Operation, payload.DataType),
		}
		database.DB.Create(&event)
	}
}

// HandleSyncData processes an incoming sync data message.
func (ss *SyncService) HandleSyncData(msg MeshMessage) *MeshMessage {
	var payload SyncPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return ss.syncAck("", "invalid payload: "+err.Error())
	}

	// Resolve conflict: accept if source has newer timestamp
	if !ss.resolver.Resolve(payload.EntityID, payload.SourceID, payload.Timestamp) {
		return ss.syncAck(payload.EntityID, "conflict: existing data is newer")
	}

	err := ss.applySync(payload)
	if err != nil {
		log.Printf("Sync apply error: %v", err)
		return ss.syncAck(payload.EntityID, err.Error())
	}

	// Log event
	if database.DB != nil {
		event := database.SyncEvent{
			Type:     string(payload.DataType) + "_sync",
			SourceID: msg.SenderID,
			EntityID: payload.EntityID,
			Status:   "applied",
		}
		database.DB.Create(&event)
	}

	return ss.syncAck(payload.EntityID, "")
}

func (ss *SyncService) syncAck(entityID, errMsg string) *MeshMessage {
	ack := SyncAckPayload{
		Success: errMsg == "",
		Error:   errMsg,
	}
	ackBytes, _ := json.Marshal(ack)
	return &MeshMessage{
		Type:      MsgSyncAck,
		SenderID:  ss.nodeID,
		SenderName: ss.nodeName,
		Timestamp: time.Now().UnixMilli(),
		Payload:   ackBytes,
	}
}

func (ss *SyncService) applySync(payload SyncPayload) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	switch payload.DataType {
	case SyncUsers:
		return ss.syncUser(payload)
	case SyncTraffic:
		return ss.syncTraffic(payload)
	case SyncInbounds:
		return ss.syncInbound(payload)
	case SyncClients:
		return ss.syncClient(payload)
	case SyncSettings:
		return ss.syncSetting(payload)
	case SyncBans:
		return ss.syncBan(payload)
	}
	return nil
}

func (ss *SyncService) syncUser(payload SyncPayload) error {
	var user database.User
	if err := json.Unmarshal(payload.Data, &user); err != nil {
		return err
	}

	switch payload.Operation {
	case OpCreate, OpUpdate:
		return database.DB.Save(&user).Error
	case OpDelete:
		return database.DB.Delete(&user).Error
	}
	return nil
}

func (ss *SyncService) syncTraffic(payload SyncPayload) error {
	var traffic struct {
		UserID int64 `json:"user_id"`
		Up     int64 `json:"up"`
		Down   int64 `json:"down"`
	}
	if err := json.Unmarshal(payload.Data, &traffic); err != nil {
		return err
	}
	// Only apply if traffic values are higher (accumulating)
	return database.DB.Model(&database.User{}).
		Where("id = ? AND traffic_up < ? AND traffic_down < ?", traffic.UserID, traffic.Up, traffic.Down).
		Updates(map[string]any{
			"traffic_up":   traffic.Up,
			"traffic_down": traffic.Down,
		}).Error
}

func (ss *SyncService) syncInbound(payload SyncPayload) error {
	var inbound database.Inbound
	if err := json.Unmarshal(payload.Data, &inbound); err != nil {
		return err
	}
	switch payload.Operation {
	case OpCreate, OpUpdate:
		return database.DB.Save(&inbound).Error
	case OpDelete:
		return database.DB.Delete(&inbound).Error
	}
	return nil
}

func (ss *SyncService) syncClient(payload SyncPayload) error {
	var client database.Client
	if err := json.Unmarshal(payload.Data, &client); err != nil {
		return err
	}
	switch payload.Operation {
	case OpCreate, OpUpdate:
		return database.DB.Save(&client).Error
	case OpDelete:
		return database.DB.Delete(&client).Error
	}
	return nil
}

func (ss *SyncService) syncSetting(payload SyncPayload) error {
	var setting database.Setting
	if err := json.Unmarshal(payload.Data, &setting); err != nil {
		return err
	}
	return database.DB.Where("\"key\" = ?", setting.Key).Assign(setting).FirstOrCreate(&setting).Error
}

func (ss *SyncService) syncBan(payload SyncPayload) error {
	var ban struct {
		Email  string `json:"email"`
		Reason string `json:"reason"`
		Banned bool   `json:"banned"`
	}
	if err := json.Unmarshal(payload.Data, &ban); err != nil {
		return err
	}
	if ban.Banned {
		return database.DB.Model(&database.User{}).Where("username = ?", ban.Email).Update("status", "banned").Error
	}
	return database.DB.Model(&database.User{}).Where("username = ?", ban.Email).Update("status", "active").Error
}

// RequestFullSync requests a full data sync from the leader.
func (ss *SyncService) RequestFullSync(leaderAddr string) error {
	if leaderAddr == "" {
		return fmt.Errorf("no leader address")
	}

	reqMsg := MeshMessage{
		Type:      MsgSyncRequest,
		SenderID:  ss.nodeID,
		SenderName: ss.nodeName,
		Timestamp: time.Now().UnixMilli(),
	}

	client := NewPeerClient(leaderAddr)
	resp, err := client.Send(reqMsg)
	if err != nil {
		return fmt.Errorf("full sync request to leader %s: %w", leaderAddr, err)
	}

	if resp == nil || resp.Type != MsgSyncData {
		return fmt.Errorf("unexpected response type: %v", resp.Type)
	}

	var payload SyncPayload
	if err := json.Unmarshal(resp.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal full sync: %w", err)
	}

	var fullSync FullSyncPayload
	if err := json.Unmarshal(payload.Data, &fullSync); err != nil {
		return fmt.Errorf("unmarshal full sync data: %w", err)
	}

	// Apply full sync data locally
	return ss.applyFullSync(fullSync)
}

func (ss *SyncService) applyFullSync(fs FullSyncPayload) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tx := database.DB.Begin()

	for _, user := range fs.Users {
		tx.Clauses(clauseOnConflictDoNothing()).Create(&user)
	}
	for _, client := range fs.Clients {
		tx.Clauses(clauseOnConflictDoNothing()).Create(&client)
	}
	for _, inbound := range fs.Inbounds {
		tx.Clauses(clauseOnConflictDoNothing()).Create(&inbound)
	}
	for _, setting := range fs.Settings {
		tx.Where("\"key\" = ?", setting.Key).Assign(setting).FirstOrCreate(&setting)
	}

	return tx.Commit().Error
}

// HandleFullSyncRequest generates a full sync payload for a requesting peer.
func (ss *SyncService) HandleFullSyncRequest() *MeshMessage {
	if database.DB == nil {
		return nil
	}

	var users []database.User
	var clients []database.Client
	var inbounds []database.Inbound
	var settings []database.Setting

	database.DB.Find(&users)
	database.DB.Find(&clients)
	database.DB.Find(&inbounds)
	database.DB.Find(&settings)

	fullSync := FullSyncPayload{
		Users:    users,
		Clients:  clients,
		Inbounds: inbounds,
		Settings: settings,
	}

	fullSyncBytes, _ := json.Marshal(fullSync)
	payload := SyncPayload{
		DataType:  "full_sync",
		Operation: OpCreate,
		Data:      fullSyncBytes,
		Timestamp: time.Now().UnixMilli(),
		SourceID:  ss.nodeID,
	}
	payloadBytes, _ := json.Marshal(payload)

	return &MeshMessage{
		Type:      MsgSyncData,
		SenderID:  ss.nodeID,
		SenderName: ss.nodeName,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payloadBytes,
	}
}

func (ss *SyncService) sendToLeader(payload SyncPayload) {
	leaderAddr := ss.getLeaderAddr()
	if leaderAddr == "" {
		return
	}
	payloadBytes, _ := json.Marshal(payload)
	msg := MeshMessage{
		Type:      MsgSyncData,
		SenderID:  ss.nodeID,
		SenderName: ss.nodeName,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payloadBytes,
	}
	client := NewPeerClient(leaderAddr)
	if _, err := client.Send(msg); err != nil {
		log.Printf("Send to leader %s failed: %v", leaderAddr, err)
	}
}

func clauseOnConflictDoNothing() clause.OnConflict {
	return clause.OnConflict{DoNothing: true}
}
