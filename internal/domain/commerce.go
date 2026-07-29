package domain

// Wallet holds a user's financial balance.
type Wallet struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"user_id" db:"user_id"`
	Balance   int64  `json:"balance" db:"balance"` // in smallest currency unit (toman)
	CreatedAt int64  `json:"created_at" db:"created_at"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}

// Transaction records a wallet credit/debit operation.
type Transaction struct {
	ID          int64  `json:"id" db:"id"`
	UserID      int64  `json:"user_id" db:"user_id"`
	Amount      int64  `json:"amount" db:"amount"`
	Type        string `json:"type" db:"type"` // deposit, withdraw, payment, refund, commission
	Description string `json:"description,omitempty" db:"description"`
	ReferenceID string `json:"reference_id,omitempty" db:"reference_id"`
	Status      string `json:"status" db:"status"` // pending, confirmed, failed
	CreatedAt   int64  `json:"created_at" db:"created_at"`
}

// Plan defines a sellable service plan.
type Plan struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	Price       int64  `json:"price" db:"price"`
	DataLimit   int64  `json:"data_limit,omitempty" db:"data_limit"`
	SpeedLimit  int    `json:"speed_limit,omitempty" db:"speed_limit"`
	DeviceLimit int    `json:"device_limit,omitempty" db:"device_limit"`
	Duration    int    `json:"duration,omitempty" db:"duration"` // in days
	Protocol    string `json:"protocol,omitempty" db:"protocol"`
	NodeGroup   string `json:"node_group,omitempty" db:"node_group"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
}

// Order represents a user purchase of a plan.
type Order struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"user_id" db:"user_id"`
	PlanID    int64  `json:"plan_id" db:"plan_id"`
	Amount    int64  `json:"amount" db:"amount"`
	Status    string `json:"status" db:"status"` // pending, paid, cancelled, refunded
	ProofFile string `json:"proof_file,omitempty" db:"proof_file"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
	PaidAt    int64  `json:"paid_at,omitempty" db:"paid_at"`
}

// Ticket represents a support ticket.
type Ticket struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"user_id" db:"user_id"`
	Subject   string `json:"subject" db:"subject"`
	Message   string `json:"message" db:"message"`
	Status    string `json:"status" db:"status"` // open, answered, closed
	CreatedAt int64  `json:"created_at" db:"created_at"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}

// TicketReply is a response within a support ticket.
type TicketReply struct {
	ID        int64  `json:"id" db:"id"`
	TicketID  int64  `json:"ticket_id" db:"ticket_id"`
	UserID    int64  `json:"user_id" db:"user_id"`
	IsAdmin   bool   `json:"is_admin" db:"is_admin"`
	Message   string `json:"message" db:"message"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}
