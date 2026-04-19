package engine

import (
	"fmt"
	"sort"
	"time"
)

func parseHHMM(v string) (time.Time, error) {
	return time.Parse("15:04", v)
}

func roundHours(minutes, unit int) int {
	if minutes <= 0 {
		return 0
	}
	if minutes%unit == 0 {
		return minutes / 60
	}
	return (minutes + unit - 1) / 60
}

func Calculate(input Input) (Output, error) {
	otMinutes := map[string]int{}
	breakMinutes := map[string]int{}
	dateLabelByEmployee := map[string]string{}

	for _, e := range input.OTEntries {
		s, err := parseHHMM(e.StartTime)
		if err != nil {
			return Output{}, fmt.Errorf("invalid OT startTime for employee %s: %w", e.EmployeeID, err)
		}
		en, err := parseHHMM(e.EndTime)
		if err != nil {
			return Output{}, fmt.Errorf("invalid OT endTime for employee %s: %w", e.EmployeeID, err)
		}
		otMinutes[e.EmployeeID] += minutesDiff(s, en)
		if e.Date != "" {
			dateLabelByEmployee[e.EmployeeID] = e.Date
		}
	}

	for _, e := range input.BreakEntries {
		s, err := parseHHMM(e.StartTime)
		if err != nil {
			return Output{}, fmt.Errorf("invalid BREAK startTime for employee %s: %w", e.EmployeeID, err)
		}
		en, err := parseHHMM(e.EndTime)
		if err != nil {
			return Output{}, fmt.Errorf("invalid BREAK endTime for employee %s: %w", e.EmployeeID, err)
		}
		breakMinutes[e.EmployeeID] += minutesDiff(s, en)
		if e.Date != "" {
			dateLabelByEmployee[e.EmployeeID] = e.Date
		}
	}

	employeesMap := map[string]struct{}{}
	for k := range otMinutes {
		employeesMap[k] = struct{}{}
	}
	for k := range breakMinutes {
		employeesMap[k] = struct{}{}
	}

	employees := make([]string, 0, len(employeesMap))
	for e := range employeesMap {
		employees = append(employees, e)
	}
	sort.Strings(employees)

	result := make([]EmployeeSummary, 0, len(employees))
	now := time.Now().Unix()
	for _, emp := range employees {
		net := otMinutes[emp] - breakMinutes[emp]
		if net < 0 {
			net = 0
		}
		s := EmployeeSummary{
			EmployeeID:          emp,
			DateLabel:           dateLabelByEmployee[emp],
			Rate20Minutes:       net,
			Rate20RoundedHours:  roundHours(net, 20),
			Rate15Minutes:       net,
			Rate15RoundedHours:  roundHours(net, 15),
			TotalOTMinutes:      otMinutes[emp],
			TotalBreakMinutes:   breakMinutes[emp],
			NetWorkMinutes:      net,
			CalculatedAtUnixSec: now,
		}
		result = append(result, s)
	}

	return Output{
		DailySummary:   result,
		MonthlySummary: result,
	}, nil
}
