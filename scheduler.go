package main

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// startScheduler registers the two daily reminder slots and starts the cron runner.
// 06:00 — shows today's piket; 21:00 — shows tomorrow's piket.
// The returned *cron.Cron is already running; call Stop on it.
func startScheduler(ctx context.Context, client *whatsmeow.Client, groupJID types.JID, cfg *Config, tracker *CheckinTracker, duty *DutySchedule) (*cron.Cron, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("timezone %q tidak valid: %w", cfg.Timezone, err)
	}

	c := cron.New(cron.WithLocation(loc))

	// 06:00 — today's piket
	if _, err := c.AddFunc("0 6 * * *", func() {
		msg := buildMessage(cfg.Message, duty, time.Now().In(loc), 0)
		if err := sendReminder(ctx, client, groupJID, msg, tracker); err != nil {
			fmt.Printf("⚠️  gagal kirim reminder (06:00): %v\n", err)
		}
	}); err != nil {
		return nil, fmt.Errorf("daftarkan jadwal 06:00: %w", err)
	}

	// 21:00 — tomorrow's piket
	if _, err := c.AddFunc("0 21 * * *", func() {
		msg := buildMessage(cfg.Message, duty, time.Now().In(loc), 1)
		if err := sendReminder(ctx, client, groupJID, msg, tracker); err != nil {
			fmt.Printf("⚠️  gagal kirim reminder (21:00): %v\n", err)
		}
	}); err != nil {
		return nil, fmt.Errorf("daftarkan jadwal 21:00: %w", err)
	}

	c.Start()
	fmt.Printf("⏰ Penjadwal aktif (%s). Reminder: 06:00 (piket hari ini), 21:00 (piket besok).\n", cfg.Timezone)
	return c, nil
}

// buildMessage appends the duty block for the given day offset to the base message.
func buildMessage(base string, duty *DutySchedule, now time.Time, dayOffset int) string {
	if duty == nil {
		return base
	}
	block := duty.ForDate(now.AddDate(0, 0, dayOffset))
	if block == "" {
		return base
	}
	return base + "\n\n" + block
}
