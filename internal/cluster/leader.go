package cluster

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// ─── Leader Election (Bully Algorithm with Raft-style terms) ─────────

// NodeRole represents the role of a cluster node.
type NodeRole string

const (
	RoleLeader    NodeRole = "leader"
	RoleFollower  NodeRole = "follower"
	RoleCandidate NodeRole = "candidate"
)

// LeaderElection implements a hybrid leader election combining
// the Bully algorithm (higher priority wins) with Raft-style terms.
type LeaderElection struct {
	mu sync.RWMutex

	// Node identity
	nodeID   int64
	nodeName string
	priority int

	// Election state
	role      NodeRole
	term      int64
	votedFor  int64
	leaderID  int64
	leaderName string

	// Cluster peers
	peers     []ClusterPeer

	// Callbacks
	onLeaderElected func(leaderID int64, term int64)
	onVoteRequest   func(candidateID int64, term int64) bool

	// Heartbeat tracking
	lastHeartbeat time.Time
	heartbeatTimeout time.Duration

	// Control
	stopCh    chan struct{}
	running   atomic.Bool
}

// ClusterPeer represents a known peer in the cluster for election purposes.
type ClusterPeer struct {
	ID       int64
	Name     string
	Address  string
	Priority int
	Term     int64
	Online   bool
}

// NewLeaderElection creates a new leader election instance.
func NewLeaderElection(nodeID int64, nodeName string, priority int) *LeaderElection {
	return &LeaderElection{
		nodeID:           nodeID,
		nodeName:         nodeName,
		priority:         priority,
		role:             RoleFollower,
		lastHeartbeat:    time.Now(),
		heartbeatTimeout: 15 * time.Second,
		stopCh:           make(chan struct{}),
	}
}

// OnLeaderElected registers a callback for when a leader is elected.
func (le *LeaderElection) OnLeaderElected(fn func(leaderID int64, term int64)) {
	le.mu.Lock()
	le.onLeaderElected = fn
	le.mu.Unlock()
}

// OnVoteRequest registers a callback to decide whether to grant a vote.
func (le *LeaderElection) OnVoteRequest(fn func(candidateID int64, term int64) bool) {
	le.mu.Lock()
	le.onVoteRequest = fn
	le.mu.Unlock()
}

// SetPeers updates the list of known cluster peers.
func (le *LeaderElection) SetPeers(peers []ClusterPeer) {
	le.mu.Lock()
	le.peers = peers
	le.mu.Unlock()
}

// Start begins the leader election loop.
func (le *LeaderElection) Start() {
	if !le.running.CompareAndSwap(false, true) {
		return
	}
	go le.electionLoop()
	log.Printf("Leader election started (node=%s priority=%d)", le.nodeName, le.priority)
}

// Stop stops the election loop.
func (le *LeaderElection) Stop() {
	if le.running.CompareAndSwap(true, false) {
		close(le.stopCh)
	}
}

func (le *LeaderElection) electionLoop() {
	checkTicker := time.NewTicker(5 * time.Second)
	defer checkTicker.Stop()

	for {
		select {
		case <-checkTicker.C:
			le.checkHeartbeat()
		case <-le.stopCh:
			return
		}
	}
}

func (le *LeaderElection) checkHeartbeat() {
	le.mu.RLock()
	role := le.role
	timeSinceHeartbeat := time.Since(le.lastHeartbeat)
	timeout := le.heartbeatTimeout
	le.mu.RUnlock()

	// Only followers start elections when heartbeat times out
	if role != RoleFollower {
		return
	}

	if timeSinceHeartbeat > timeout {
		log.Printf("Heartbeat timeout (%v), starting election...", timeSinceHeartbeat)
		le.StartElection()
	}
}

// StartElection begins the election process (Bully algorithm).
func (le *LeaderElection) StartElection() {
	le.mu.Lock()
	le.role = RoleCandidate
	atomic.AddInt64(&le.term, 1)
	currentTerm := le.term
	le.votedFor = le.nodeID // vote for self
	le.mu.Unlock()

	log.Printf("Starting election: node=%s term=%d priority=%d", le.nodeName, currentTerm, le.priority)

	// Count self-vote
	votes := 1
	totalPeers := 0

	// Snapshot peers under lock, then release before making HTTP calls
	le.mu.RLock()
	peers := make([]ClusterPeer, len(le.peers))
	copy(peers, le.peers)
	le.mu.RUnlock()

	for _, peer := range peers {
		if !peer.Online {
			continue
		}
		totalPeers++

		// In Bully algorithm: only peers with higher priority need to be asked
		if peer.Priority > le.priority {
			// Request vote from higher-priority peer
			msg := MeshMessage{
				Type:       MsgVoteRequest,
				SenderID:   le.nodeID,
				SenderName: le.nodeName,
				Term:       currentTerm,
				Timestamp:  time.Now().UnixMilli(),
			}
			payload, _ := json.Marshal(VotePayload{
				CandidateID:   le.nodeID,
				CandidateName: le.nodeName,
				Priority:      le.priority,
				Term:          currentTerm,
			})
			msg.Payload = payload

			client := NewPeerClient(peer.Address)
			resp, err := client.Send(msg)
			if err != nil {
				log.Printf("  Vote request to %s failed: %v", peer.Name, err)
				continue
			}
			if resp != nil && resp.Type == MsgVoteResponse {
				var voteResult VoteResultPayload
				if json.Unmarshal(resp.Payload, &voteResult) == nil && voteResult.Granted {
					votes++
				}
			}
		} else {
			// Lower-priority peer automatically votes for us
			votes++
		}
	}

	// Need majority to win
	majority := (totalPeers + 1) / 2 + 1
	if votes >= majority {
		log.Printf("Election WON: node=%s term=%d votes=%d/%d", le.nodeName, currentTerm, votes, totalPeers+1)
		le.DeclareLeader(le.nodeID, le.nodeName, currentTerm)
	} else {
		log.Printf("Election LOST: node=%s term=%d votes=%d/%d needed=%d", le.nodeName, currentTerm, votes, totalPeers+1, majority)
		le.mu.Lock()
		le.role = RoleFollower
		le.mu.Unlock()
	}
}

// DeclareLeader announces a new leader to the cluster.
func (le *LeaderElection) DeclareLeader(newLeaderID int64, newLeaderName string, term int64) {
	le.mu.Lock()
	le.role = RoleLeader
	le.leaderID = newLeaderID
	le.leaderName = newLeaderName
	le.term = term
	le.lastHeartbeat = time.Now()
	le.mu.Unlock()

	log.Printf("Leader declared: node=%s id=%d term=%d", newLeaderName, newLeaderID, term)

	if le.onLeaderElected != nil {
		le.onLeaderElected(newLeaderID, term)
	}

	// Announce to all peers
	le.mu.RLock()
	peers := le.peers
	le.mu.RUnlock()

	announceMsg := MeshMessage{
		Type:       MsgLeaderAnnounce,
		SenderID:   newLeaderID,
		SenderName: newLeaderName,
		Term:       term,
		Timestamp:  time.Now().UnixMilli(),
	}
	payload, _ := json.Marshal(LeaderAnnouncePayload{
		LeaderID:   newLeaderID,
		LeaderName: newLeaderName,
		Term:       term,
	})
	announceMsg.Payload = payload

	for _, peer := range peers {
		if !peer.Online {
			continue
		}
		client := NewPeerClient(peer.Address)
		if _, err := client.Send(announceMsg); err != nil {
			log.Printf("  Announce to %s failed: %v", peer.Name, err)
		}
	}
}

// HandleVoteRequest processes an incoming vote request.
func (le *LeaderElection) HandleVoteRequest(msg MeshMessage) *MeshMessage {
	var voteReq VotePayload
	if err := json.Unmarshal(msg.Payload, &voteReq); err != nil {
		return nil
	}

	le.mu.Lock()
	defer le.mu.Unlock()

	granted := false
	if voteReq.Term > le.term {
		le.term = voteReq.Term
		le.votedFor = voteReq.CandidateID
		// Bully: higher priority responds; we step down if candidate has higher priority
		if voteReq.Priority > le.priority {
			le.role = RoleFollower
			granted = true
		}
	} else if voteReq.Term == le.term && le.votedFor == 0 {
		le.votedFor = voteReq.CandidateID
		granted = true
	}

	resultPayload, _ := json.Marshal(VoteResultPayload{
		VoterID: le.nodeID,
		Term:    le.term,
		Granted: granted,
	})

	return &MeshMessage{
		Type:       MsgVoteResponse,
		SenderID:   le.nodeID,
		SenderName: le.nodeName,
		Term:       le.term,
		Timestamp:  time.Now().UnixMilli(),
		Payload:    resultPayload,
	}
}

// HandleLeaderAnnounce processes a leader announcement.
func (le *LeaderElection) HandleLeaderAnnounce(msg MeshMessage) {
	var announce LeaderAnnouncePayload
	if err := json.Unmarshal(msg.Payload, &announce); err != nil {
		return
	}

	le.mu.Lock()
	// Accept if term is higher or equal
	if announce.Term >= le.term {
		le.role = RoleFollower
		le.leaderID = announce.LeaderID
		le.leaderName = announce.LeaderName
		le.term = announce.Term
		le.lastHeartbeat = time.Now()
		log.Printf("Accepted leader: %s (id=%d term=%d)", announce.LeaderName, announce.LeaderID, announce.Term)

		le.mu.Unlock()

		if le.onLeaderElected != nil {
			le.onLeaderElected(announce.LeaderID, announce.Term)
		}
	} else {
		le.mu.Unlock()
	}
}

// HandleHeartbeat processes a heartbeat from the leader.
func (le *LeaderElection) HandleHeartbeat() {
	le.mu.Lock()
	le.lastHeartbeat = time.Now()
	le.mu.Unlock()
}

// IsLeader returns whether this node is the leader.
func (le *LeaderElection) IsLeader() bool {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.role == RoleLeader
}

// GetRole returns the current role.
func (le *LeaderElection) GetRole() NodeRole {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.role
}

// GetLeaderID returns the current leader's node ID.
func (le *LeaderElection) GetLeaderID() int64 {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.leaderID
}

// GetLeaderName returns the current leader's name.
func (le *LeaderElection) GetLeaderName() string {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.leaderName
}

// GetTerm returns the current election term.
func (le *LeaderElection) GetTerm() int64 {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return le.term
}

// Stats returns election statistics.
func (le *LeaderElection) Stats() map[string]any {
	le.mu.RLock()
	defer le.mu.RUnlock()
	return map[string]any{
		"node_id":      le.nodeID,
		"node_name":    le.nodeName,
		"role":         string(le.role),
		"term":         le.term,
		"leader_id":    le.leaderID,
		"leader_name":  le.leaderName,
		"priority":     le.priority,
		"voted_for":    le.votedFor,
		"is_leader":    le.role == RoleLeader,
		"last_heartbeat": le.lastHeartbeat.UnixMilli(),
	}
}
