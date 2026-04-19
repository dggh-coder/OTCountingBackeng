package service

import (
	"fmt"
	"strings"
	"time"
)

func normalizePeriod(period string) (string, error) {
	p := strings.ToUpper(strings.TrimSpace(period))
	if p != "AM" && p != "PM" {
		return "", ValidationError("period must be AM or PM")
	}
	return p, nil
}

func parseDate(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, nil
	}
	if t, err := time.Parse("01/02/2006", v); err == nil {
		return t, nil
	}
	return time.Time{}, ValidationError("date must be YYYY-MM-DD or MM/DD/YYYY")
}

func validateEmployee(v string) error {
	if v != "A" && v != "B" {
		return ValidationError("employeeId must be A or B")
	}
	return nil
}

func validateEntryType(v string) error {
	if v != "OT" && v != "BREAK" {
		return ValidationError("entryType must be OT or BREAK")
	}
	return nil
}

func validateHHMM(v string) error {
	if _, err := time.Parse("15:04", v); err != nil {
		return ValidationError(fmt.Sprintf("invalid time %q, expected HH:MM", v))
	}
	return nil
}

func sessionIDFrom(date time.Time, period string) int64 {
	suffix := "01"
	if period == "PM" {
		suffix = "02"
	}
	var id int64
	fmt.Sscanf(date.Format("20060102")+suffix, "%d", &id)
	return id
}
