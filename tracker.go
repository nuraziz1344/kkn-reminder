package main

import (
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/types"
)

// checkmarkEmojis are the reaction emojis that count as "checked in".
var checkmarkEmojis = map[string]bool{
	"✅": true,
	"✔️": true,
	"☑️": true,
}

// CheckinTracker tracks which group members have checked in today and which
// message IDs belong to today's reminders. Everything resets at midnight WIB.
type CheckinTracker struct {
	mu          sync.Mutex
	loc         *time.Location
	date        string          // "YYYY-MM-DD" sentinel for daily reset
	checkedIn   map[string]bool // normalized JID string → true
	reminderIDs map[string]bool // sent message IDs for today's reminders
}

func NewCheckinTracker(timezone string) (*CheckinTracker, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("timezone %q tidak valid: %w", timezone, err)
	}
	t := &CheckinTracker{loc: loc}
	t.checkedIn = make(map[string]bool)
	t.reminderIDs = make(map[string]bool)
	t.date = t.today()
	return t, nil
}

func (t *CheckinTracker) today() string {
	return time.Now().In(t.loc).Format("2006-01-02")
}

// maybeReset clears all state when the calendar date has rolled over.
func (t *CheckinTracker) maybeReset() {
	if d := t.today(); d != t.date {
		t.date = d
		t.checkedIn = make(map[string]bool)
		t.reminderIDs = make(map[string]bool)
	}
}

// AddReminderID registers a sent reminder's message ID so reactions can be matched.
func (t *CheckinTracker) AddReminderID(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maybeReset()
	t.reminderIDs[id] = true
}

// IsReminderMessage reports whether id belongs to one of today's reminders.
func (t *CheckinTracker) IsReminderMessage(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maybeReset()
	return t.reminderIDs[id]
}

// MarkCheckedIn records that a member has checked in for today.
func (t *CheckinTracker) MarkCheckedIn(jid types.JID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maybeReset()
	t.checkedIn[jid.ToNonAD().String()] = true
}

// IsCheckedIn reports whether a member has already checked in today.
func (t *CheckinTracker) IsCheckedIn(jid types.JID) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maybeReset()
	return t.checkedIn[jid.ToNonAD().String()]
}

// IsCheckmark reports whether an emoji reaction counts as a check-in.
func IsCheckmark(emoji string) bool {
	return checkmarkEmojis[emoji]
}
