package outbox

import "time"

const (
	EventTypeContractResolved = "contract_resolved"
	EventTypeExecutionCreated = "execution_created"
	EventTypeProjectionChange = "projection_change"
)

type Event struct {
	ID           int64
	EventID      string
	EventType    string
	Payload      []byte
	CreatedAt    time.Time
	PublishedAt  *time.Time
	AttemptCount int64
	LastError    *string
}
