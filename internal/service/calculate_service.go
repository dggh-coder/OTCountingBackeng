package service

import (
	"context"
	"fmt"
	"otcountingbackend/internal/engine"
	"time"
)

type CalculateService struct {
	Sessions   SessionRepo
	Entries    EntryRepo
	Results    ResultRepo
	Rendered   RenderedRepo
	Transactor Transactor
	Engine     Engine
}

func (c *CalculateService) CalculateAndPersist(ctx context.Context, sessionID int64) (engine.Output, error) {
	txCtx, tx, err := c.Transactor.BeginTx(ctx)
	if err != nil {
		return engine.Output{}, InternalError("failed to begin transaction")
	}
	defer tx.Rollback(txCtx)

	session, err := c.Sessions.GetByID(txCtx, sessionID)
	if err != nil {
		return engine.Output{}, InternalError("failed to read session")
	}
	if session == nil {
		return engine.Output{}, NotFoundError("session not found")
	}
	if session.Status != "OPEN" {
		return engine.Output{}, ConflictError("session is not OPEN")
	}

	entries, err := c.Entries.ListBySession(txCtx, sessionID)
	if err != nil {
		return engine.Output{}, InternalError("failed to read entries")
	}

	in := engine.Input{}
	for _, e := range entries {
		entry := engine.Entry{ID: e.ID, EmployeeID: e.EmployeeID, Date: session.Date, Period: session.Period, StartTime: e.StartTime, EndTime: e.EndTime}
		if e.EntryType == "OT" {
			in.OTEntries = append(in.OTEntries, entry)
		} else {
			in.BreakEntries = append(in.BreakEntries, entry)
		}
	}

	out, err := c.Engine.Calculate(in)
	if err != nil {
		return engine.Output{}, ValidationError("calculation failed", err.Error())
	}

	now := time.Now().UTC()
	records := make([]ResultRecord, 0, len(out.DailySummary))
	for _, r := range out.DailySummary {
		records = append(records, ResultRecord{
			SessionID:          sessionID,
			EmployeeID:         r.EmployeeID,
			DateLabel:          session.Date,
			Rate20Minutes:      r.Rate20Minutes,
			Rate20RoundedHours: r.Rate20RoundedHours,
			Rate15Minutes:      r.Rate15Minutes,
			Rate15RoundedHours: r.Rate15RoundedHours,
			CalculatedAt:       now,
		})
	}

	if err := c.Results.UpsertMany(txCtx, records); err != nil {
		return engine.Output{}, InternalError("failed to persist results")
	}

	for _, r := range out.DailySummary {
		html := fmt.Sprintf("<div class=\"daily-card\"><strong>%s</strong><span>%d mins</span></div>", r.EmployeeID, r.Rate20Minutes)
		_ = c.Rendered.Upsert(txCtx, RenderedFragment{SessionID: sessionID, EmployeeID: r.EmployeeID, FragmentType: "DAILY_CARD", FormatVersion: 1, ContentHTML: html, LastCalculated: now})
	}

	if err := tx.Commit(txCtx); err != nil {
		return engine.Output{}, InternalError("failed to commit transaction")
	}

	return out, nil
}

func (c *CalculateService) GetResults(ctx context.Context, sessionID int64) ([]ResultRecord, error) {
	rows, err := c.Results.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, InternalError("failed to list results")
	}
	if len(rows) == 0 {
		if s, err := c.Sessions.GetByID(ctx, sessionID); err == nil && s == nil {
			return nil, NotFoundError("session not found")
		}
	}
	return rows, nil
}
