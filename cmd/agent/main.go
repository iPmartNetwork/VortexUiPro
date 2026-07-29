// ═══════════════════════════════════════════════════════════════════════
// VortexUiPro — Node Agent
//
// Lightweight agent binary that runs on remote nodes.
// Communicates with the panel via gRPC for heartbeat, config sync,
// traffic reporting, and core engine management.
//
// Usage:
//   vortexuipro-agent --panel-addr panel:50051 --grpc-addr :50051
//
// Build:
//   CGO_ENABLED=0 go build -o vortexuipro-agent ./cmd/agent
//
// ═══════════════════════════════════════════════════════════════════════

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vortexuipro/internal/agent"
)

// Set via ldflags at build time: -X main.version=<version>
var version = "dev"

var (
	panelAddr = flag.String("panel-addr", "", "Panel gRPC address (e.g., panel:50051)")
	grpcAddr  = flag.String("grpc-addr", ":50051", "Agent gRPC listen address")
	nodeName  = flag.String("node-name", "", "Unique node name (default: hostname)")
	nodeID    = flag.Int64("node-id", 0, "Node ID assigned by panel")
	logLevel  = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
)

func main() {
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("🚀 VortexUiPro Node Agent starting...")

	// ── Resolve node name ──────────────────────────────────────────
	if *nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("Failed to get hostname: %v", err)
		}
		*nodeName = hostname
	}

	log.Printf("  Node ID:    %d", *nodeID)
	log.Printf("  Node Name:  %s", *nodeName)
	log.Printf("  gRPC Addr:  %s", *grpcAddr)
	log.Printf("  Panel Addr: %s", *panelAddr)

	// ── Start gRPC server ──────────────────────────────────────────
	server := agent.NewNodeAgentServer(*grpcAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start node agent server: %v", err)
	}

	// ── Connect to panel (if panel address provided) ───────────────
	var client *agent.NodeAgentClient
	if *panelAddr != "" {
		client = agent.NewNodeAgentClient(*panelAddr)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := client.Connect(ctx)
		cancel()
		if err != nil {
			log.Printf("⚠️  Failed to connect to panel at %s: %v", *panelAddr, err)
			log.Printf("   Agent will run in standalone mode")
		} else {
			log.Printf("✅ Connected to panel at %s", *panelAddr)
			defer client.Close()
		}
	}

	// ── Start heartbeat goroutine ──────────────────────────────────
	heartbeatDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				heartbeat := agent.Heartbeat{
					NodeID:  *nodeID,
					Name:    *nodeName,
					Status:  "online",
					Address: *grpcAddr,
				}
				server.RegisterHeartbeat(heartbeat)
				log.Printf("💓 Heartbeat sent for node %s", *nodeName)
			case <-heartbeatDone:
				return
			}
		}
	}()

	// ── Graceful shutdown ──────────────────────────────────────────
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("\n🛑 Shutting down...")
	close(heartbeatDone)
	server.Stop()
	log.Println("✅ Node agent stopped gracefully")
}

func init() {
	// Override flags from environment variables
	if v := os.Getenv("NODE_GRPC_ADDR"); v != "" {
		*grpcAddr = v
	}
	if v := os.Getenv("PANEL_GRPC_ADDR"); v != "" {
		*panelAddr = v
	}
	if v := os.Getenv("NODE_NAME"); v != "" {
		*nodeName = v
	}
	if v := os.Getenv("NODE_ID"); v != "" {
		var id int64
		fmt.Sscanf(v, "%d", &id)
		*nodeID = id
	}
}
