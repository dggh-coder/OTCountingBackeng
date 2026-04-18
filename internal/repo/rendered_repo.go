package repo

import (
	"context"
	"database/sql"
	"errors"
	"otcountingbackend/internal/service"
)

type RenderedRepo struct{ DB *sql.DB }

func (r *RenderedRepo) Upsert(ctx context.Context, row service.RenderedFragment) error {
	q := querier(ctx, r.DB)
	_, err := q.ExecContext(ctx, `
INSERT INTO session_rendered_fragment (session_id, employee_id, fragment_type, format_version, content_html, last_calculated)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (session_id, employee_id, fragment_type, format_version)
DO UPDATE SET content_html=EXCLUDED.content_html, last_calculated=EXCLUDED.last_calculated`,
		row.SessionID, row.EmployeeID, row.FragmentType, row.FormatVersion, row.ContentHTML, row.LastCalculated)
	return err
}

func (r *RenderedRepo) Get(ctx context.Context, sessionID int64, employeeID, fragmentType string, formatVersion int) (*service.RenderedFragment, error) {
	q := querier(ctx, r.DB)
	var row service.RenderedFragment
	err := q.QueryRowContext(ctx, `SELECT session_id, employee_id, fragment_type, format_version, content_html, last_calculated FROM session_rendered_fragment WHERE session_id=$1 AND employee_id=$2 AND fragment_type=$3 AND format_version=$4`, sessionID, employeeID, fragmentType, formatVersion).
		Scan(&row.SessionID, &row.EmployeeID, &row.FragmentType, &row.FormatVersion, &row.ContentHTML, &row.LastCalculated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
