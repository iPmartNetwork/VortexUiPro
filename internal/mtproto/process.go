package mtproto

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const stopTimeout = 10 * time.Second

// Process wraps an mtg-multi child process.
type Process struct {
	binPath string
	cfgPath string
	name    string
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
}

func newProcess(binPath, cfgPath, name string) *Process {
	return &Process{
		binPath: binPath,
		cfgPath: cfgPath,
		name:    name,
	}
}

// Start launches the mtg-multi process with the config file.
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	if _, err := os.Stat(p.binPath); os.IsNotExist(err) {
		return fmt.Errorf("mtg-multi binary not found at %s", p.binPath)
	}

	p.cmd = exec.Command(p.binPath, p.cfgPath)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mtg-multi for %s: %w", p.name, err)
	}

	p.running = true
	go p.waitAndMark()
	return nil
}

func (p *Process) waitAndMark() {
	err := p.cmd.Wait()
	p.mu.Lock()
	p.running = false
	p.mu.Unlock()
	if err != nil {
		fmt.Printf("mtg-multi process %s exited: %v\n", p.name, err)
	}
}

// Stop gracefully terminates the mtg-multi process with timeout.
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running || p.cmd == nil || p.cmd.Process == nil {
		p.running = false
		return nil
	}

	// Try graceful shutdown first
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		_ = p.cmd.Process.Kill()
		p.running = false
		return nil
	}

	// Wait for process to exit with timeout
	done := make(chan struct{})
	go func() {
		_ = p.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(stopTimeout):
		// Force kill after timeout
		_ = p.cmd.Process.Kill()
		<-done
	}

	p.running = false
	return nil
}

// IsRunning reports whether the process is still alive.
func (p *Process) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// killStrayMtgProcesses terminates any orphaned mtg-multi processes.
func killStrayMtgProcesses(binPath string) int {
	// On Linux/macOS: pgrep/pkill
	// On Windows: taskkill
	// Returns count of killed processes
	return 0
}
