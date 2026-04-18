package repo

import (
	"context"
	"database/sql"
	"otcountingbackend/internal/service"
)

type ResultRepo struct{ DB *sql.DB }

func (r *ResultRepo) UpsertMany(ctx context.Context, records []service.ResultRecord) error {
	q := querier(ctx, r.DB)
	for _, rec := range records {
		_, err := q.ExecContext(ctx, `
INSERT INTO session_result (
	session_id, employee_id, date_label,
	rate20_minutes, rate20_rounded_hours,
	rate15_minutes, rate15_rounded_hours,
	total_ot_minutes, total_break_minutes, net_work_minutes,
	calculated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (session_id, employee_id)
DO UPDATE SET
	date_label=EXCLUDED.date_label,
	rate20_minutes=EXCLUDED.rate20_minutes,
	rate20_rounded_hours=EXCLUDED.rate20_rounded_hours,
	rate15_minutes=EXCLUDED.rate15_minutes,
	rate15_rounded_hours=EXCLUDED.rate15_rounded_hours,
	total_ot_minutes=EXCLUDED.total_ot_minutes,
	total_break_minutes=EXCLUDED.total_break_minutes,
	net_work_minutes=EXCLUDED.net_work_minutes,
	calculated_at=EXCLUDED.calculated_at`,
			rec.SessionID, rec.EmployeeID, rec.DateLabel,
			rec.Rate20Minutes, rec.Rate20RoundedHours,
			rec.Rate15Minutes, rec.Rate15RoundedHours,
			rec.TotalOTMinutes, rec.TotalBreakMinutes, rec.NetWorkMinutes,
			rec.CalculatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ResultRepo) ListBySession(ctx context.Context, sessionID int64) ([]service.ResultRecord, error) {
	q := querier(ctx, r.DB)
	rows, err := q.QueryContext(ctx, `SELECT session_id, employee_id, date_label, rate20_minutes, rate20_rounded_hours, rate15_minutes, rate15_rounded_hours, total_ot_minutes, total_break_minutes, net_work_minutes, calculated_at FROM session_result WHERE session_id=$1 ORDER BY employee_id`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []service.ResultRecord{}
	for rows.Next() {
		var r service.ResultRecord
		if err := rows.Scan(&r.SessionID, &r.EmployeeID, &r.DateLabel, &r.Rate20Minutes, &r.Rate20RoundedHours, &r.Rate15Minutes, &r.Rate15RoundedHours, &r.TotalOTMinutes, &r.TotalBreakMinutes, &r.NetWorkMinutes, &r.CalculatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

var _ = sql.ErrNoRows
