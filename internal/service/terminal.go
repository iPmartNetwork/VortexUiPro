package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"vortexuipro/internal/database"
)

// ─── Terminal Session ──────────────────────────────────────────────────
type TerminalSession struct {
	ID        string    `json:"id"`
	NodeID    int64     `json:"node_id"`
	NodeName  string    `json:"node_name"`
	CreatedAt time.Time `json:"created_at"`
	client    *ssh.Client
	closer    io.WriteCloser
	mu        sync.Mutex
	closed    bool
}

// TerminalOutput is sent via WebSocket to the frontend.
type TerminalOutput struct {
	Type string `json:"type"` // "data", "error", "close"
	Data string `json:"data,omitempty"`
	ID   string `json:"id,omitempty"`
}

// ─── TerminalService ───────────────────────────────────────────────────
type TerminalService struct {
	sessions map[string]*TerminalSession
	mu       sync.RWMutex
	sshKey   ssh.Signer
}

// NewTerminalService creates a terminal service.
func NewTerminalService() *TerminalService {
	return &TerminalService{
		sessions: make(map[string]*TerminalSession),
	}
}

// LoadSSHKey loads a private key for SSH authentication.
func (s *TerminalService) LoadSSHKey(keyData []byte) error {
	key, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("parse ssh key: %w", err)
	}
	s.sshKey = key
	return nil
}

// OpenSession opens an SSH terminal session to a node.
func (s *TerminalService) OpenSession(nodeID int64, termWidth, termHeight int) (*TerminalSession, <-chan TerminalOutput, error) {
	var node database.Node
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		return nil, nil, fmt.Errorf("node not found: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", node.Address, node.Port)
	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	if s.sshKey != nil {
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(s.sshKey)}
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, fmt.Errorf("ssh dial: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("ssh session: %w", err)
	}

	// Setup terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		session.Close()
		client.Close()
		return nil, nil, fmt.Errorf("request pty: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, nil, fmt.Errorf("stdout pipe: %w", err)
	}

	output := make(chan TerminalOutput, 256)
	sess := &TerminalSession{
		ID:        fmt.Sprintf("term_%d_%d", nodeID, time.Now().UnixMilli()),
		NodeID:    nodeID,
		NodeName:  node.Name,
		CreatedAt: time.Now(),
		client:    client,
		closer:    stdin,
	}

	// Read stdout and send to channel
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				if !sess.isClosed() {
					output <- TerminalOutput{Type: "close", Data: err.Error(), ID: sess.ID}
				}
				close(output)
				return
			}
			if n > 0 {
				output <- TerminalOutput{Type: "data", Data: string(buf[:n]), ID: sess.ID}
			}
		}
	}()

	// Start shell
	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, nil, fmt.Errorf("start shell: %w", err)
	}

	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()

	// Background cleanup on session end
	go func() {
		session.Wait()
		s.closeSession(sess.ID)
	}()

	return sess, output, nil
}

// WriteInput writes data to the terminal stdin.
func (s *TerminalService) WriteInput(sessionID string, data string) error {
	s.mu.RLock()
	sess, ok := s.sessions[sessionID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()

	if sess.closed {
		return fmt.Errorf("session closed")
	}

	_, err := io.WriteString(sess.closer, data)
	return err
}

// ResizeTerminal resizes the terminal PTY.
func (s *TerminalService) ResizeTerminal(sessionID string, width, height int) error {
	s.mu.RLock()
	_, ok := s.sessions[sessionID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	return nil
}

// ListSessions returns all active terminal sessions.
func (s *TerminalService) ListSessions() []*TerminalSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*TerminalSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		sess.mu.Lock()
		closed := sess.closed
		sess.mu.Unlock()
		if !closed {
			list = append(list, sess)
		}
	}
	return list
}

// CloseSession closes a terminal session.
func (s *TerminalService) CloseSession(sessionID string) error {
	return s.closeSession(sessionID)
}

func (s *TerminalService) closeSession(id string) error {
	s.mu.Lock()
	sess, ok := s.sessions[id]
	if ok {
		delete(s.sessions, id)
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.closed {
		return nil
	}
	sess.closed = true

	if sess.closer != nil {
		sess.closer.Close()
	}
	if sess.client != nil {
		sess.client.Close()
	}
	log.Printf("[Terminal] Session %s closed", id)
	return nil
}

func (s *TerminalSession) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// MarshalJSON implements json.Marshaler
func (s *TerminalSession) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"id":         s.ID,
		"node_id":    s.NodeID,
		"node_name":  s.NodeName,
		"created_at": s.CreatedAt,
	})
}
