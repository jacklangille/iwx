package exchangecorehttp

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"iwx/go_backend/internal/authcontext"
	"iwx/go_backend/internal/store"
)

func (s *Server) handleAccountMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleAccountMeShow)(w, r)
}

func (s *Server) handleAccountMeShow(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	account, err := s.service.GetCashAccount(r.Context(), claims.UserID, r.URL.Query().Get("currency"))
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeCashAccount(account))
}

func (s *Server) handleAccountLedger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleAccountLedgerIndex)(w, r)
}

func (s *Server) handleAccountLedgerIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	limit := 50
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid limit"})
			return
		}
		limit = parsed
	}

	entries, err := s.service.ListLedgerEntries(r.Context(), claims.UserID, r.URL.Query().Get("currency"), limit)
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeLedgerEntries(entries))
}

func (s *Server) handleAccountCollateralLocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleAccountCollateralLocksIndex)(w, r)
}

func (s *Server) handleAccountCollateralLocksIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	locks, err := s.service.ListCollateralLocks(r.Context(), claims.UserID, r.URL.Query().Get("currency"))
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeCollateralLocks(locks))
}

func (s *Server) handleAccountCashReservations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	requireAuth(s.config, s.handleAccountCashReservationsIndex)(w, r)
}

func (s *Server) handleAccountCashReservationsIndex(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	reservations, err := s.service.ListOrderCashReservations(r.Context(), claims.UserID, r.URL.Query().Get("currency"))
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, serializeOrderCashReservations(reservations))
}

func (s *Server) handleAccountDeposits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleAccountDepositCreate)(w, r)
}

func (s *Server) handleAccountDepositCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	request, err := decodeAccountMoneyRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	account, entry, err := s.service.DepositCash(r.Context(), store.DepositCashInput{
		UserID:        claims.UserID,
		Currency:      request.Currency,
		AmountCents:   request.AmountCents,
		ReferenceID:   request.ReferenceID,
		CorrelationID: request.CorrelationID,
		Description:   request.Description,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core deposit user_id=%d amount_cents=%d currency=%s account_id=%d", claims.UserID, request.AmountCents, account.Currency, account.ID)
	writeJSON(w, http.StatusCreated, accountMutationResponse{
		Account:     serializeCashAccount(account),
		LedgerEntry: ptrLedgerEntryResponse(serializeLedgerEntry(entry)),
	})
}

func (s *Server) handleAccountWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleAccountWithdrawalCreate)(w, r)
}

func (s *Server) handleAccountWithdrawalCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	request, err := decodeAccountMoneyRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	account, entry, err := s.service.WithdrawCash(r.Context(), store.WithdrawCashInput{
		UserID:        claims.UserID,
		Currency:      request.Currency,
		AmountCents:   request.AmountCents,
		ReferenceID:   request.ReferenceID,
		CorrelationID: request.CorrelationID,
		Description:   request.Description,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core withdrawal user_id=%d amount_cents=%d currency=%s account_id=%d", claims.UserID, request.AmountCents, account.Currency, account.ID)
	writeJSON(w, http.StatusCreated, accountMutationResponse{
		Account:     serializeCashAccount(account),
		LedgerEntry: ptrLedgerEntryResponse(serializeLedgerEntry(entry)),
	})
}

func (s *Server) handleCollateralLocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleCollateralLockCreate)(w, r)
}

func (s *Server) handleCollateralLockCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	request, err := decodeCollateralLockRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	lock, account, entry, err := s.service.CreateCollateralLock(r.Context(), store.CreateCollateralLockInput{
		UserID:        claims.UserID,
		ContractID:    request.ContractID,
		Currency:      request.Currency,
		AmountCents:   request.AmountCents,
		ReferenceID:   request.ReferenceID,
		CorrelationID: request.CorrelationID,
		Description:   request.Description,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core collateral_lock_created user_id=%d contract_id=%d amount_cents=%d lock_id=%d", claims.UserID, lock.ContractID, lock.AmountCents, lock.ID)
	writeJSON(w, http.StatusCreated, collateralLockMutationResponse{
		CollateralLock: serializeCollateralLock(lock),
		Account:        serializeCashAccount(account),
		LedgerEntry:    ptrLedgerEntryResponse(serializeLedgerEntry(entry)),
	})
}

func (s *Server) handleCollateralLockActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleCollateralLockRelease)(w, r)
}

func (s *Server) handleCollateralLockRelease(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	lockID, err := parseReleasePathID(r.URL.Path, "/api/accounts/collateral_locks/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid collateral lock path"})
		return
	}

	request, err := decodeReleaseRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	lock, account, entry, err := s.service.ReleaseCollateralLock(r.Context(), store.ReleaseCollateralLockInput{
		UserID:        claims.UserID,
		LockID:        lockID,
		CorrelationID: request.CorrelationID,
		Description:   request.Description,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core collateral_lock_released user_id=%d lock_id=%d", claims.UserID, lock.ID)
	writeJSON(w, http.StatusOK, collateralLockMutationResponse{
		CollateralLock: serializeCollateralLock(lock),
		Account:        serializeCashAccount(account),
		LedgerEntry:    ptrLedgerEntryResponse(serializeLedgerEntry(entry)),
	})
}

func (s *Server) handleCashReservations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleCashReservationCreate)(w, r)
}

func (s *Server) handleCashReservationCreate(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	request, err := decodeCashReservationRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	reservation, account, entry, err := s.service.CreateOrderCashReservation(r.Context(), store.CreateOrderCashReservationInput{
		UserID:        claims.UserID,
		ContractID:    request.ContractID,
		Currency:      request.Currency,
		AmountCents:   request.AmountCents,
		ReferenceType: request.ReferenceType,
		ReferenceID:   request.ReferenceID,
		CorrelationID: request.CorrelationID,
		Description:   request.Description,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core cash_reservation_created user_id=%d contract_id=%d reservation_id=%d amount_cents=%d", claims.UserID, reservation.ContractID, reservation.ID, reservation.AmountCents)
	writeJSON(w, http.StatusCreated, cashReservationMutationResponse{
		CashReservation: serializeOrderCashReservation(reservation),
		Account:         serializeCashAccount(account),
		LedgerEntry:     ptrLedgerEntryResponse(serializeLedgerEntry(entry)),
	})
}

func (s *Server) handleCashReservationActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	requireAuth(s.config, s.handleCashReservationRelease)(w, r)
}

func (s *Server) handleCashReservationRelease(w http.ResponseWriter, r *http.Request) {
	claims, err := authcontext.ClaimsFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing auth context"})
		return
	}

	reservationID, err := parseReleasePathID(r.URL.Path, "/api/accounts/cash_reservations/")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid cash reservation path"})
		return
	}

	request, err := decodeReleaseRequest(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	reservation, account, entry, err := s.service.ReleaseOrderCashReservation(r.Context(), store.ReleaseOrderCashReservationInput{
		UserID:        claims.UserID,
		ReservationID: reservationID,
		CorrelationID: request.CorrelationID,
		Description:   request.Description,
	})
	if err != nil {
		writeExchangeCoreError(w, err)
		return
	}

	log.Printf("exchange-core cash_reservation_released user_id=%d reservation_id=%d", claims.UserID, reservation.ID)
	writeJSON(w, http.StatusOK, cashReservationMutationResponse{
		CashReservation: serializeOrderCashReservation(reservation),
		Account:         serializeCashAccount(account),
		LedgerEntry:     ptrLedgerEntryResponse(serializeLedgerEntry(entry)),
	})
}

func parseReleasePathID(path, prefix string) (int64, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[1] != "release" {
		return 0, errors.New("invalid release path")
	}

	return strconv.ParseInt(parts[0], 10, 64)
}
