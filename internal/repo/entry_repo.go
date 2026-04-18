package repo

import (
	"context"
	"database/sql"
	"otcountingbackend/internal/service"
)

type EntryRepo struct{ DB *sql.DB }

func (r *EntryRepo) ReplaceBySession(ctx context.Context, sessionID int64, entries []service.SessionEntry) error {
	q := querier(ctx, r.DB)
	if _, err := q.ExecContext(ctx, `DELETE FROM time_entry WHERE session_id=$1`, sessionID); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := q.ExecContext(ctx, `INSERT INTO time_entry (entry_id, session_id, employee_id, entry_type, start_time, end_time) VALUES ($1,$2,$3,$4,$5,$6)`, e.ID, sessionID, e.EmployeeID, e.EntryType, e.StartTime, e.EndTime); err != nil {
			return err
		}
	}
	return nil
}

func (r *EntryRepo) ListBySession(ctx context.Context, sessionID int64) ([]service.SessionEntry, error) {
	q := querier(ctx, r.DB)
	rows, err := q.QueryContext(ctx, `SELECT session_id, entry_id, employee_id, entry_type, start_time, end_time FROM time_entry WHERE session_id=$1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []service.SessionEntry
	for rows.Next() {
		var e service.SessionEntry
		if err := rows.Scan(&e.SessionID, &e.ID, &e.EmployeeID, &e.EntryType, &e.StartTime, &e.EndTime); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

var _ = sql.ErrNoRows
