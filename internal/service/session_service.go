package service

import (
	"context"

	"github.com/google/uuid"
)

type SessionService struct {
	Sessions   SessionRepo
	Entries    EntryRepo
	Transactor Transactor
}

func (s *SessionService) CreateOrGetSession(ctx context.Context, dateInput, periodInput string) (*Session, error) {
	period, err := normalizePeriod(periodInput)
	if err != nil {
		return nil, err
	}
	d, err := parseDate(dateInput)
	if err != nil {
		return nil, err
	}
	date := d.Format("2006-01-02")

	existing, err := s.Sessions.GetByDatePeriod(ctx, date, period)
	if err != nil {
		return nil, InternalError("failed to get session")
	}
	if existing != nil {
		return existing, nil
	}

	session := Session{SessionID: sessionIDFrom(d, period), Date: date, Period: period, Status: "OPEN"}
	if err := s.Sessions.Create(ctx, session); err != nil {
		return nil, InternalError("failed to create session")
	}
	return &session, nil
}

func (s *SessionService) ReplaceEntries(ctx context.Context, sessionID int64, payload []EntryPayload) error {
	txCtx, tx, err := s.Transactor.BeginTx(ctx)
	if err != nil {
		return InternalError("failed to begin transaction")
	}
	defer tx.Rollback(txCtx)

	session, err := s.Sessions.GetByID(txCtx, sessionID)
	if err != nil {
		return InternalError("failed to read session")
	}
	if session == nil {
		return NotFoundError("session not found")
	}
	if session.Status != "OPEN" {
		return ConflictError("session is not OPEN")
	}

	rows := make([]SessionEntry, 0, len(payload))
	for _, p := range payload {
		if err := validateEmployee(p.Employee); err != nil {
			return err
		}
		if err := validateEntryType(p.EntryType); err != nil {
			return err
		}
		if err := validateHHMM(p.StartTime); err != nil {
			return err
		}
		if err := validateHHMM(p.EndTime); err != nil {
			return err
		}

		id := p.ID
		if id == "" {
			id = uuid.NewString()
		} else if _, err := uuid.Parse(id); err != nil {
			return ValidationError("id must be a valid UUID")
		}

		rows = append(rows, SessionEntry{SessionID: sessionID, ID: id, EmployeeID: p.Employee, EntryType: p.EntryType, StartTime: p.StartTime, EndTime: p.EndTime})
	}
	if err := s.Entries.ReplaceBySession(txCtx, sessionID, rows); err != nil {
		return InternalError("failed to replace entries")
	}

	if err := tx.Commit(txCtx); err != nil {
		return InternalError("failed to commit entries transaction")
	}
	return nil
}
