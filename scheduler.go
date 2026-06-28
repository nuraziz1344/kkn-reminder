package main

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// checkpointSchedules are the daily KKN checkpoint times (cron format: min hour * * *).
var checkpointSchedules = []string{
	"0 6 * * *",  // 06:00
	"0 12 * * *", // 12:00
	"0 18 * * *", // 18:00
	"0 22 * * *", // 22:00
	"45 23 * * *", // 23:45
}

// startScheduler wires the checkpoint times to sendReminder, pinned to the
// configured timezone. The returned *cron.Cron is already running; call Stop on it.
func startScheduler(ctx context.Context, client *whatsmeow.Client, groupJID types.JID, cfg *Config, tracker *CheckinTracker) (*cron.Cron, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("timezone %q tidak valid: %w", cfg.Timezone, err)
	}

	c := cron.New(cron.WithLocation(loc))
	for _, spec := range checkpointSchedules {
		spec := spec // capture
		if _, err := c.AddFunc(spec, func() {
			if err := sendReminder(ctx, client, groupJID, cfg.Message, tracker); err != nil {
				fmt.Printf("⚠️  gagal kirim reminder (%s): %v\n", spec, err)
			}
		}); err != nil {
			return nil, fmt.Errorf("daftarkan jadwal %q: %w", spec, err)
		}
	}

	c.Start()
	fmt.Printf("⏰ Penjadwal aktif (%s). Reminder: 06:00, 12:00, 18:00, 22:00, 23:45.\n", cfg.Timezone)
	return c, nil
}
