package service

import "time"

type Session struct {
	SessionID int64  `json:"sessionId"`
	Date      string `json:"date"`
	Period    string `json:"period"`
	Status    string `json:"status"`
}

type EntryPayload struct {
	ID        string `json:"id"`
	Employee  string `json:"employeeId"`
	EntryType string `json:"entryType"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type SessionEntry struct {
	SessionID  int64
	ID         string
	EmployeeID string
	EntryType  string
	StartTime  string
	EndTime    string
}

type ResultRecord struct {
	SessionID          int64
	EmployeeID         string
	DateLabel          string
	Rate20Minutes      int
	Rate20RoundedHours int
	Rate15Minutes      int
	Rate15RoundedHours int
	TotalOTMinutes     int
	TotalBreakMinutes  int
	NetWorkMinutes     int
	CalculatedAt       time.Time
}

type RenderedFragment struct {
	SessionID      int64
	EmployeeID     string
	FragmentType   string
	FormatVersion  int
	ContentHTML    string
	LastCalculated time.Time
}
