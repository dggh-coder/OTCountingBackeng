package repo

import (
	"context"
	"database/sql"
	"errors"
	"otcountingbackend/internal/service"
)

type SessionRepo struct{ DB *sql.DB }

func (r *SessionRepo) GetByID(ctx context.Context, sessionID int64) (*service.Session, error) {
	q := querier(ctx, r.DB)
	var s service.Session
	err := q.QueryRowContext(ctx, `SELECT session_id, work_date::text, period, status FROM work_session WHERE session_id=$1`, sessionID).
		Scan(&s.SessionID, &s.Date, &s.Period, &s.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepo) GetByDatePeriod(ctx context.Context, date, period string) (*service.Session, error) {
	q := querier(ctx, r.DB)
	var s service.Session
	err := q.QueryRowContext(ctx, `SELECT session_id, work_date::text, period, status FROM work_session WHERE work_date=$1 AND period=$2`, date, period).
		Scan(&s.SessionID, &s.Date, &s.Period, &s.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepo) Create(ctx context.Context, s service.Session) error {
	q := querier(ctx, r.DB)
	_, err := q.ExecContext(ctx, `INSERT INTO work_session (session_id, work_date, period, status) VALUES ($1,$2,$3,$4)`, s.SessionID, s.Date, s.Period, s.Status)
	return err
}
