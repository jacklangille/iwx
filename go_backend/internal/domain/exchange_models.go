package domain

import "time"

type CashAccount struct {
	ID             int64
	UserID         int64
	Currency       string
	AvailableCents int64
	LockedCents    int64
	TotalCents     int64
	UpdatedAt      time.Time
}

type LedgerEntry struct {
	ID            int64
	AccountID     int64
	UserID        int64
	EntryType     LedgerEntryType
	AmountCents   int64
	ReferenceType string
	ReferenceID   string
	CorrelationID string
	Description   string
	OccurredAt    time.Time
}

type CollateralLock struct {
	ID                  int64
	UserID              int64
	ContractID          int64
	Currency            string
	AmountCents         int64
	Status              CollateralLockStatus
	ReferenceID         string
	Description         string
	ReferenceIssuanceID *int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ReleasedAt          *time.Time
}

type IssuanceBatch struct {
	ID               int64
	ContractID       int64
	CreatorUserID    int64
	CollateralLockID int64
	AboveQuantity    int64
	BelowQuantity    int64
	Status           IssuanceBatchStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Position struct {
	ID                int64
	UserID            int64
	ContractID        int64
	Side              ClaimSide
	AvailableQuantity int64
	LockedQuantity    int64
	TotalQuantity     int64
	UpdatedAt         time.Time
}

type PositionLock struct {
	ID            int64
	UserID        int64
	ContractID    int64
	Side          ClaimSide
	Quantity      int64
	Status        PositionLockStatus
	OrderID       *int64
	ReferenceType string
	ReferenceID   string
	CorrelationID string
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReleasedAt    *time.Time
}

type OrderCashReservation struct {
	ID            int64
	UserID        int64
	ContractID    int64
	Currency      string
	AmountCents   int64
	Status        CashReservationStatus
	ReferenceType string
	ReferenceID   string
	CorrelationID string
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReleasedAt    *time.Time
}

type OracleObservation struct {
	ID                     int64
	ContractID             int64
	ProviderName           string
	StationID              string
	ObservedMetric         string
	ObservationWindowStart time.Time
	ObservationWindowEnd   time.Time
	ObservedValue          string
	NormalizedValue        string
	ObservedAt             time.Time
	RecordedAt             time.Time
}

type ContractResolution struct {
	ID                     int64
	EventID                string
	ContractID             int64
	ProviderName           string
	StationID              string
	ObservedMetric         string
	ObservationWindowStart time.Time
	ObservationWindowEnd   time.Time
	RuleVersion            string
	ResolvedValue          string
	Outcome                ResolutionOutcome
	ResolvedAt             time.Time
	PublishedAt            *time.Time
}

type SettlementEntry struct {
	ID          int64
	ContractID  int64
	UserID      int64
	EntryType   SettlementEntryType
	Outcome     ResolutionOutcome
	AmountCents int64
	Quantity    int64
	ReferenceID string
	CreatedAt   time.Time
}

type ContractRule struct {
	ID                      int64
	ContractID              int64
	RuleVersion             string
	Metric                  string
	Threshold               *int64
	MeasurementUnit         string
	ResolutionInclusiveSide ClaimSide
	CreatedAt               time.Time
}
