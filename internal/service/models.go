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
	SessionID  int64  `json:"sessionId"`
	ID         string `json:"id"`
	EmployeeID string `json:"employeeId"`
	EntryType  string `json:"entryType"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type ResultRecord struct {
	SessionID          int64     `json:"sessionId"`
	EmployeeID         string    `json:"employeeId"`
	DateLabel          string    `json:"dateLabel"`
	Rate20Minutes      int       `json:"rate20Minutes"`
	Rate20RoundedHours int       `json:"rate20RoundedHours"`
	Rate15Minutes      int       `json:"rate15Minutes"`
	Rate15RoundedHours int       `json:"rate15RoundedHours"`
	CalculatedAt       time.Time `json:"calculatedAt"`
}

type RenderedFragment struct {
	SessionID      int64     `json:"sessionId"`
	EmployeeID     string    `json:"employeeId"`
	FragmentType   string    `json:"fragmentType"`
	FormatVersion  int       `json:"formatVersion"`
	ContentHTML    string    `json:"contentHtml"`
	LastCalculated time.Time `json:"lastCalculated"`
}
