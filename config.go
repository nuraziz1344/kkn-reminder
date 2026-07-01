package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the bot's runtime settings, loaded from config.json.
type Config struct {
	GroupJID      string `json:"group_jid"`
	Timezone      string `json:"timezone"`
	Message       string `json:"message"`
	CookSchedule  string `json:"cook_schedule"`
	CleanSchedule string `json:"clean_schedule"`
}

// loadConfig reads and validates config.json from the given path.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("baca %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Jakarta"
	}
	if cfg.CookSchedule == "" {
		cfg.CookSchedule = "jadwal-masak.csv"
	}
	if cfg.CleanSchedule == "" {
		cfg.CleanSchedule = "jadwal-kebersihan.csv"
	}
	if cfg.GroupJID == "" || cfg.GroupJID == "PASTE_FROM_-list@g.us" {
		return nil, fmt.Errorf("group_jid belum diisi di %s — jalankan dulu dengan flag -list untuk mendapatkan JID grup", path)
	}
	if cfg.Message == "" {
		return nil, fmt.Errorf("message kosong di %s", path)
	}

	return &cfg, nil
}
