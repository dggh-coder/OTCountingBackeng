package engine

import "time"

type Entry struct {
	ID         string `json:"id"`
	EmployeeID string `json:"employeeId"`
	Date       string `json:"date,omitempty"`
	Period     string `json:"period,omitempty"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type Input struct {
	OTEntries    []Entry `json:"otEntries"`
	BreakEntries []Entry `json:"breakEntries"`
}

type EmployeeSummary struct {
	EmployeeID          string `json:"employeeId"`
	DateLabel           string `json:"dateLabel"`
	Rate20Minutes       int    `json:"rate20Minutes"`
	Rate20RoundedHours  int    `json:"rate20RoundedHours"`
	Rate15Minutes       int    `json:"rate15Minutes"`
	Rate15RoundedHours  int    `json:"rate15RoundedHours"`
	TotalOTMinutes      int    `json:"totalOTMinutes"`
	TotalBreakMinutes   int    `json:"totalBreakMinutes"`
	NetWorkMinutes      int    `json:"netWorkMinutes"`
	CalculatedAtUnixSec int64  `json:"calculatedAtUnixSec,omitempty"`
}

type Output struct {
	DailySummary   []EmployeeSummary `json:"dailySummary"`
	MonthlySummary []EmployeeSummary `json:"monthlySummary"`
}

func minutesDiff(start, end time.Time) int {
	d := end.Sub(start)
	if d < 0 {
		d += 24 * time.Hour
	}
	return int(d.Minutes())
}
