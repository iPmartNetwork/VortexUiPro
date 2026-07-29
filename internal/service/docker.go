package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"vortexuipro/internal/database"
)

// ─── Types ───────────────────────────────────────────────────────────

// DockerImage represents a Docker image.
type DockerImage struct {
	ID     string `json:"id"`
	Tags   string `json:"tags"`
	Size   int64  `json:"size"`
	Created string `json:"created"`
}

// DockerContainerInfo holds detailed container information.
type DockerContainerInfo struct {
	ContainerID string            `json:"container_id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	Ports       string            `json:"ports"`
	Created     string            `json:"created"`
	CPU         float64           `json:"cpu"`
	Memory      float64           `json:"memory"`
	State       string            `json:"state"`
	Health      string            `json:"health,omitempty"`
	IP          string            `json:"ip,omitempty"`
}

// ─── DockerService ───────────────────────────────────────────────────

// DockerService provides Docker container management from within the panel.
type DockerService struct {
	mu       sync.RWMutex
	enabled  bool
}

// NewDockerService creates a new Docker service.
func NewDockerService() *DockerService {
	// Check if Docker is available
	_, err := exec.LookPath("docker")
	return &DockerService{
		enabled: err == nil,
	}
}

// IsEnabled returns whether Docker is available on this system.
func (s *DockerService) IsEnabled() bool {
	return s.enabled
}

// ─── Container Operations ────────────────────────────────────────────

// ListContainers returns all containers from Docker.
func (s *DockerService) ListContainers(all bool) ([]DockerContainerInfo, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Docker not available")
	}

	args := []string{"ps", "--format", `{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.Ports}}|{{.CreatedAt}}|{{.State}}`, "--no-trunc"}
	if all {
		args = append(args, "-a")
	}

	output, err := s.runDocker(args...)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	var containers []DockerContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 7)
		if len(parts) >= 4 {
			info := DockerContainerInfo{
				ContainerID: parts[0][:12],
				Name:        parts[1],
				Image:       parts[2],
				Status:      parts[3],
				State:       parts[6],
			}
			if len(parts) > 4 {
				info.Ports = parts[4]
			}
			if len(parts) > 5 {
				info.Created = parts[5]
			}

			// Get container stats
			stats, _ := s.GetContainerStats(parts[0][:12])
			if stats != nil {
				info.CPU = stats.CPU
				info.Memory = stats.Memory
			}

			containers = append(containers, info)

			// Sync to database
			s.syncToDB(info)
		}
	}

	if containers == nil {
		containers = []DockerContainerInfo{}
	}
	return containers, nil
}

// CreateContainer creates a new Docker container.
func (s *DockerService) CreateContainer(name, image, envVars string, port int, network string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("Docker not available")
	}

	args := []string{"run", "-d", "--name", name, "--restart", "unless-stopped"}

	// Add environment variables
	if envVars != "" {
		var envs []string
		if err := json.Unmarshal([]byte(envVars), &envs); err == nil {
			for _, e := range envs {
				args = append(args, "-e", e)
			}
		}
	}

	// Add port mapping
	if port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", port, port))
	}

	// Add network
	if network != "" {
		args = append(args, "--network", network)
	}

	args = append(args, image)

	output, err := s.runDocker(args...)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	containerID := strings.TrimSpace(output)

	// Save to database
	dbContainer := &database.DockerContainer{
		ContainerID: containerID[:12],
		Name:        name,
		Image:       image,
		Status:      "running",
		Port:        port,
		EnvVars:     envVars,
		AutoRestart: true,
	}
	database.DB.Create(dbContainer)

	return containerID[:12], nil
}

// StartContainer starts a stopped container.
func (s *DockerService) StartContainer(containerID string) error {
	_, err := s.runDocker("start", containerID)
	return err
}

// StopContainer stops a running container.
func (s *DockerService) StopContainer(containerID string) error {
	_, err := s.runDocker("stop", containerID)
	return err
}

// RestartContainer restarts a container.
func (s *DockerService) RestartContainer(containerID string) error {
	_, err := s.runDocker("restart", containerID)
	return err
}

// RemoveContainer removes a container.
func (s *DockerService) RemoveContainer(containerID string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)
	_, err := s.runDocker(args...)
	if err == nil {
		database.DB.Where("container_id = ?", containerID).Delete(&database.DockerContainer{})
	}
	return err
}

// GetContainerLogs returns logs from a container.
func (s *DockerService) GetContainerLogs(containerID string, tail int) (string, error) {
	if tail <= 0 {
		tail = 100
	}
	return s.runDocker("logs", "--tail", fmt.Sprintf("%d", tail), containerID)
}

// ─── Image Operations ────────────────────────────────────────────────

// ListImages returns all Docker images.
func (s *DockerService) ListImages() ([]DockerImage, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Docker not available")
	}

	output, err := s.runDocker("images", "--format", `{{.ID}}|{{.Repository}}:{{.Tag}}|{{.Size}}|{{.CreatedAt}}`, "--no-trunc")
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	var images []DockerImage
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) >= 3 {
			img := DockerImage{
				ID:     parts[0][7:19],
				Tags:   parts[1],
				Size:   parseSize(parts[2]),
				Created: parts[3],
			}
			images = append(images, img)
		}
	}

	if images == nil {
		images = []DockerImage{}
	}
	return images, nil
}

// PullImage pulls a Docker image from registry.
func (s *DockerService) PullImage(image string) error {
	_, err := s.runDocker("pull", image)
	return err
}

// RemoveImage removes a Docker image.
func (s *DockerService) RemoveImage(imageID string, force bool) error {
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, imageID)
	_, err := s.runDocker(args...)
	return err
}

// PruneImages removes unused Docker images.
func (s *DockerService) PruneImages() (string, error) {
	return s.runDocker("image", "prune", "-f")
}

// ─── Stats & Monitoring ─────────────────────────────────────────────

// ContainerStats holds real-time container resource usage.
type ContainerStats struct {
	ContainerID string  `json:"container_id"`
	CPU         float64 `json:"cpu_percent"`
	Memory      float64 `json:"memory_mb"`
	MemoryPct   float64 `json:"memory_percent"`
	NetIO       string  `json:"net_io"`
	BlockIO     string  `json:"block_io"`
	PIDs        int     `json:"pids"`
}

// GetContainerStats returns stats for a single container.
func (s *DockerService) GetContainerStats(containerID string) (*ContainerStats, error) {
	output, err := s.runDocker("stats", "--no-stream", "--format", `{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}|{{.PIDs}}`, containerID)
	if err != nil {
		return nil, err
	}

	lines := strings.SplitN(strings.TrimSpace(output), "\n", 2)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no stats output")
	}

	parts := strings.Split(lines[0], "|")
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected stats format")
	}

	stats := &ContainerStats{
		ContainerID: containerID,
	}

	// Parse CPU (e.g., "2.50%")
	cpuStr := strings.TrimSuffix(strings.TrimSpace(parts[0]), "%")
	fmt.Sscanf(cpuStr, "%f", &stats.CPU)

	// Parse Memory (e.g., "128.5MiB / 1.952GiB")
	memParts := strings.Split(parts[1], "/")
	if len(memParts) >= 1 {
		memStr := strings.TrimSpace(memParts[0])
		if strings.HasSuffix(memStr, "MiB") {
			fmt.Sscanf(strings.TrimSuffix(memStr, "MiB"), "%f", &stats.Memory)
		} else if strings.HasSuffix(memStr, "GiB") {
			var gb float64
			fmt.Sscanf(strings.TrimSuffix(memStr, "GiB"), "%f", &gb)
			stats.Memory = gb * 1024
		}
	}

	// Parse Memory %
	memPctStr := strings.TrimSuffix(strings.TrimSpace(parts[2]), "%")
	fmt.Sscanf(memPctStr, "%f", &stats.MemoryPct)

	if len(parts) > 3 {
		stats.NetIO = strings.TrimSpace(parts[3])
	}
	if len(parts) > 4 {
		stats.BlockIO = strings.TrimSpace(parts[4])
	}
	if len(parts) > 5 {
		fmt.Sscanf(strings.TrimSpace(parts[5]), "%d", &stats.PIDs)
	}

	return stats, nil
}

// ─── Dashboard Stats ────────────────────────────────────────────────

// DockerDashboardStats holds an overview of the Docker environment.
type DockerDashboardStats struct {
	ContainersTotal   int `json:"containers_total"`
	ContainersRunning int `json:"containers_running"`
	ContainersStopped int `json:"containers_stopped"`
	ImagesTotal       int `json:"images_total"`
	Enabled           bool `json:"enabled"`
}

// GetDashboardStats returns Docker overview statistics.
func (s *DockerService) GetDashboardStats() (*DockerDashboardStats, error) {
	stats := &DockerDashboardStats{
		Enabled: s.enabled,
	}

	if !s.enabled {
		return stats, nil
	}

	// Count all containers
	output, err := s.runDocker("ps", "-aq", "--format", "{{.ID}}")
	if err == nil {
		stats.ContainersTotal = len(strings.Fields(output))
	}

	// Count running containers
	output, err = s.runDocker("ps", "-q", "--format", "{{.ID}}")
	if err == nil {
		stats.ContainersRunning = len(strings.Fields(output))
	}

	stats.ContainersStopped = stats.ContainersTotal - stats.ContainersRunning

	// Count images
	output, err = s.runDocker("images", "-q", "--format", "{{.ID}}")
	if err == nil {
		stats.ImagesTotal = len(strings.Fields(output))
	}

	return stats, nil
}

// ─── Internal ────────────────────────────────────────────────────────

// runDocker executes a Docker command and returns its output.
func (s *DockerService) runDocker(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker %s: %s (stderr: %s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}

// syncToDB syncs a container's state to the database.
func (s *DockerService) syncToDB(info DockerContainerInfo) {
	status := info.State
	if status == "" {
		// Parse status phrase
		if strings.HasPrefix(info.Status, "Up") {
			status = "running"
		} else if strings.HasPrefix(info.Status, "Exited") || strings.HasPrefix(info.Status, "Created") {
			status = "stopped"
		} else if strings.HasPrefix(info.Status, "Paused") {
			status = "paused"
		} else {
			status = "unknown"
		}
	}

	var existing database.DockerContainer
	if err := database.DB.Where("container_id = ?", info.ContainerID).First(&existing).Error; err == nil {
		database.DB.Model(&existing).Updates(map[string]any{
			"status":       status,
			"cpu":          info.CPU,
			"memory":       info.Memory,
			"updated_at":   time.Now().UnixMilli(),
		})
	} else {
		database.DB.Create(&database.DockerContainer{
			ContainerID: info.ContainerID,
			Name:        info.Name,
			Image:       info.Image,
			Status:      status,
		})
	}
}

// parseSize parses a Docker size string like "128MB" or "1.5GB".
func parseSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if strings.HasSuffix(sizeStr, "GB") {
		var val float64
		fmt.Sscanf(sizeStr, "%fGB", &val)
		return int64(val * 1024 * 1024 * 1024)
	}
	if strings.HasSuffix(sizeStr, "MB") {
		var val float64
		fmt.Sscanf(sizeStr, "%fMB", &val)
		return int64(val * 1024 * 1024)
	}
	if strings.HasSuffix(sizeStr, "kB") {
		var val float64
		fmt.Sscanf(sizeStr, "%fkB", &val)
		return int64(val * 1024)
	}
	if strings.HasSuffix(sizeStr, "B") {
		var val float64
		fmt.Sscanf(sizeStr, "%fB", &val)
		return int64(val)
	}
	return 0
}

// Ensure log import is used
var _ = log.Printf
var _ = io.Discard
