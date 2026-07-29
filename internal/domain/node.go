package domain

// NodeStatus represents the operational state of a node.
type NodeStatus string

const (
	NodeOnline  NodeStatus = "online"
	NodeOffline NodeStatus = "offline"
	NodeError   NodeStatus = "error"
)

// Node represents a proxy server node.
type Node struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Address     string     `json:"address" db:"address"`
	Port        int        `json:"port,omitempty" db:"port"`
	APIPort     int        `json:"api_port,omitempty" db:"api_port"`
	Status      NodeStatus `json:"status" db:"status"`
	CoreType    CoreType   `json:"core_type" db:"core_type"`
	Enabled     bool       `json:"enable" db:"enable"`
	Country     string     `json:"country,omitempty" db:"country"`
	Location    string     `json:"location,omitempty" db:"location"`
	CPULoad     float64    `json:"cpu_load,omitempty" db:"cpu_load"`
	MemoryUsed  float64    `json:"memory_used,omitempty" db:"memory_used"`
	Uplink      int64      `json:"uplink,omitempty" db:"uplink"`
	Downlink    int64      `json:"downlink,omitempty" db:"downlink"`
	TrafficUp   int64      `json:"traffic_up,omitempty" db:"traffic_up"`
	TrafficDown int64      `json:"traffic_down,omitempty" db:"traffic_down"`
	LastHeartbeat int64    `json:"last_heartbeat,omitempty" db:"last_heartbeat"`
	Remark      string     `json:"remark,omitempty" db:"remark"`
	CreatedAt   int64      `json:"created_at" db:"created_at"`
	UpdatedAt   int64      `json:"updated_at" db:"updated_at"`
}

// NodeHeartbeat is a lightweight status update from a node agent.
type NodeHeartbeat struct {
	NodeID      int64   `json:"node_id"`
	CPULoad     float64 `json:"cpu_load"`
	MemoryUsed  float64 `json:"memory_used"`
	Uplink      int64   `json:"uplink"`
	Downlink    int64   `json:"downlink"`
	TrafficUp   int64   `json:"traffic_up"`
	TrafficDown int64   `json:"traffic_down"`
	LoadAvg     float64 `json:"load_avg"`
	Uptime      int64   `json:"uptime"`
}
