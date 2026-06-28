package main

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// newClient builds a whatsmeow client backed by a persistent SQLite session.
// On the first run there is no stored session, so the caller must handle QR login.
func newClient(ctx context.Context) (*whatsmeow.Client, error) {
	if err := os.MkdirAll("data", 0700); err != nil {
		return nil, fmt.Errorf("buat direktori data: %w", err)
	}
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:data/session.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, fmt.Errorf("buka session store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("ambil device: %w", err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	return whatsmeow.NewClient(deviceStore, clientLog), nil
}

// connect connects the client, performing a QR-code login if no session exists yet.
// On a new login it waits for the Connected event so callers can safely use the
// client (list groups, send messages) right away.
func connect(ctx context.Context, client *whatsmeow.Client) error {
	if client.Store.ID != nil {
		// Existing session — just reconnect.
		return client.Connect()
	}

	// New session: register a handler that signals when the connection is ready.
	connectedCh := make(chan struct{}, 1)
	handlerID := client.AddEventHandler(func(evt any) {
		if _, ok := evt.(*events.Connected); ok {
			select {
			case connectedCh <- struct{}{}:
			default:
			}
		}
	})
	defer client.RemoveEventHandler(handlerID)

	qrChan, _ := client.GetQRChannel(ctx)
	if err := client.Connect(); err != nil {
		return err
	}

	for evt := range qrChan {
		switch evt.Event {
		case "code":
			fmt.Println("\nScan QR ini di WhatsApp > Perangkat Tertaut (Linked Devices):")
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
		case "success":
			fmt.Println("✅ Berhasil dipasangkan! Menunggu sinkronisasi dengan WhatsApp...")
		default:
			fmt.Println("Event login:", evt.Event)
		}
	}

	// Wait until the server confirms we're connected, or 60 s timeout.
	fmt.Println("   Mohon tunggu, sedang menyinkronkan pesan...")
	select {
	case <-connectedCh:
		fmt.Println("✅ Terhubung dan siap digunakan!")
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(60 * time.Second):
		fmt.Println("⚠️  Timeout sinkronisasi, tapi koneksi mungkin masih berjalan.")
	}
	return nil
}

// listGroups prints every group the account belongs to, with its JID.
// Copy the desired JID into config.json's "group_jid".
func listGroups(ctx context.Context, client *whatsmeow.Client) error {
	groups, err := client.GetJoinedGroups(ctx)
	if err != nil {
		return fmt.Errorf("ambil daftar grup: %w", err)
	}
	fmt.Printf("\nKamu anggota dari %d grup:\n\n", len(groups))
	for _, g := range groups {
		fmt.Printf("  %-40s  %s\n", g.Name, g.JID.String())
	}
	fmt.Println("\nSalin JID grup KKN (yang diakhiri @g.us) ke field \"group_jid\" di config.json.")
	return nil
}

// sendReminder posts the configured message to the group, @mentioning members
// who have not yet checked in today. Members who reacted with a checkmark to an
// earlier reminder today are skipped. Pass tracker=nil to mention everyone.
func sendReminder(ctx context.Context, client *whatsmeow.Client, groupJID types.JID, message string, tracker *CheckinTracker) error {
	info, err := client.GetGroupInfo(ctx, groupJID)
	if err != nil {
		return fmt.Errorf("ambil info grup: %w", err)
	}

	mentions := make([]string, 0, len(info.Participants))
	skipped := 0
	for _, p := range info.Participants {
		if tracker != nil && tracker.IsCheckedIn(p.JID) {
			skipped++
			continue
		}
		mentions = append(mentions, p.JID.ToNonAD().String())
	}

	if len(mentions) == 0 {
		fmt.Printf("🎉 Semua anggota %s sudah checkin hari ini, reminder dilewati.\n", info.Name)
		return nil
	}

	resp, err := client.SendMessage(ctx, groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(message),
			ContextInfo: &waE2E.ContextInfo{MentionedJID: mentions},
		},
	})
	if err != nil {
		return fmt.Errorf("kirim pesan: %w", err)
	}

	if tracker != nil {
		tracker.AddReminderID(resp.ID)
	}

	fmt.Printf("✅ Reminder terkirim ke %s (%d di-mention, %d sudah checkin)\n", info.Name, len(mentions), skipped)
	return nil
}
