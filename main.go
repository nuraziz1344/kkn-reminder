package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func main() {
	listFlag := flag.Bool("list", false, "tampilkan semua grup beserta JID-nya, lalu keluar")
	testFlag := flag.Bool("test", false, "kirim satu reminder sekarang juga (untuk uji coba), lalu keluar")
	configPath := flag.String("config", "config.json", "path ke file konfigurasi")
	flag.Parse()

	ctx := context.Background()

	client, err := newClient(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := connect(ctx, client); err != nil {
		fmt.Println("Error koneksi:", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	// -list: print groups and exit (use this to find your group JID).
	if *listFlag {
		if err := listGroups(ctx, client); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println("Error konfigurasi:", err)
		os.Exit(1)
	}

	groupJID, err := types.ParseJID(cfg.GroupJID)
	if err != nil {
		fmt.Printf("group_jid %q tidak valid: %v\n", cfg.GroupJID, err)
		os.Exit(1)
	}

	tracker, err := NewCheckinTracker(cfg.Timezone)
	if err != nil {
		fmt.Println("Error tracker:", err)
		os.Exit(1)
	}

	// -test: send one reminder immediately, then exit.
	if *testFlag {
		if err := sendReminder(ctx, client, groupJID, cfg.Message, tracker); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	// Listen for emoji reactions on reminder messages.
	client.AddEventHandler(func(evt any) {
		msg, ok := evt.(*events.Message)
		if !ok {
			return
		}
		reaction := msg.Message.GetReactionMessage()
		if reaction == nil {
			return
		}
		targetID := reaction.GetKey().GetID()
		if !tracker.IsReminderMessage(targetID) {
			return
		}
		if IsCheckmark(reaction.GetText()) {
			tracker.MarkCheckedIn(msg.Info.Sender)
			fmt.Printf("☑️  %s sudah checkin hari ini\n", msg.Info.Sender.User)
		}
	})

	cronRunner, err := startScheduler(ctx, client, groupJID, cfg, tracker)
	if err != nil {
		fmt.Println("Error penjadwal:", err)
		os.Exit(1)
	}
	defer cronRunner.Stop()

	fmt.Println("Bot berjalan. Tekan Ctrl+C untuk berhenti.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nMenutup bot...")
}
