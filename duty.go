package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"
)

type DutySchedule struct {
	cooking      map[string]string
	cleaningCowo map[string]string
	cleaningCewe map[string]string
}

func LoadDutySchedule(cookPath, cleanPath string) (*DutySchedule, error) {
	d := &DutySchedule{
		cooking:      make(map[string]string),
		cleaningCowo: make(map[string]string),
		cleaningCewe: make(map[string]string),
	}

	if err := d.loadCooking(cookPath); err != nil {
		return nil, fmt.Errorf("jadwal masak: %w", err)
	}
	if err := d.loadCleaning(cleanPath); err != nil {
		return nil, fmt.Errorf("jadwal kebersihan: %w", err)
	}
	return d, nil
}

func (d *DutySchedule) loadCooking(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}
	for _, row := range rows[1:] { // skip header
		if len(row) >= 2 {
			d.cooking[row[0]] = row[1]
		}
	}
	return nil
}

func (d *DutySchedule) loadCleaning(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}
	for _, row := range rows[1:] { // skip header
		if len(row) >= 3 {
			d.cleaningCowo[row[0]] = row[1]
			d.cleaningCewe[row[0]] = row[2]
		}
	}
	return nil
}

// ForDate returns a formatted duty block for t, or "" if no data for that date.
func (d *DutySchedule) ForDate(t time.Time) string {
	key := t.Format("02-01-2006")
	var lines []string

	if cook := d.cooking[key]; cook != "" {
		lines = append(lines, "👨‍🍳 *Piket Masak:* "+cook)
	}

	cowo := d.cleaningCowo[key]
	cewe := d.cleaningCewe[key]
	cleaning := strings.Join(filter(cowo, cewe), ", ")
	if cleaning != "" {
		lines = append(lines, "🧹 *Piket Kebersihan:* "+cleaning)
	}

	return strings.Join(lines, "\n")
}

func filter(parts ...string) []string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
