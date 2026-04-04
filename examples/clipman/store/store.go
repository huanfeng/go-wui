package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ClipEntry represents a single clipboard history entry.
type ClipEntry struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Source    string    `json:"source"`      // source application window title
	Category  string    `json:"category"`    // user-assigned category tag
	Pinned    bool      `json:"pinned"`      // pinned entries survive auto-cleanup
	CreatedAt time.Time `json:"created_at"`
	UsedCount int       `json:"used_count"`  // number of times re-pasted
}

// Config holds persistent settings for the clipboard manager.
type Config struct {
	MaxEntries int    `json:"max_entries"` // max history entries (default 500)
	HotkeyMod  uint32 `json:"hotkey_mod"`  // modifier keys for quick panel
	HotkeyKey  uint32 `json:"hotkey_key"`  // virtual key code for quick panel
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		MaxEntries: 500,
		HotkeyMod:  0x0002 | 0x0004, // Ctrl+Shift
		HotkeyKey:  0x56,             // 'V'
	}
}

// Store manages clipboard history entries in memory with JSON persistence.
type Store struct {
	mu         sync.RWMutex
	entries    []ClipEntry
	config     Config
	dataDir    string
	dirty      bool
	saveTimer  *time.Timer
	nextID     int
}

// New creates a new Store with the given data directory.
// If dataDir is empty, it defaults to %APPDATA%/clipman.
func New(dataDir string) (*Store, error) {
	if dataDir == "" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = "."
		}
		dataDir = filepath.Join(appData, "clipman")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	s := &Store{
		dataDir: dataDir,
		config:  DefaultConfig(),
	}

	// Load existing data
	s.loadConfig()
	s.loadHistory()

	return s, nil
}

// Add inserts a new clipboard entry. Duplicates (same text) are moved to top.
func (s *Store) Add(text, source string) *ClipEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deduplicate: if same text exists, move to top and update
	for i, e := range s.entries {
		if e.Text == text {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			e.Source = source
			e.CreatedAt = time.Now()
			e.UsedCount++
			s.entries = append([]ClipEntry{e}, s.entries...)
			s.markDirty()
			return &s.entries[0]
		}
	}

	// New entry
	s.nextID++
	entry := ClipEntry{
		ID:        fmt.Sprintf("clip_%d", s.nextID),
		Text:      text,
		Source:    source,
		CreatedAt: time.Now(),
	}
	s.entries = append([]ClipEntry{entry}, s.entries...)

	// Auto-cleanup if over limit
	s.cleanup()
	s.markDirty()

	return &s.entries[0]
}

// Delete removes an entry by ID.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.entries {
		if e.ID == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			s.markDirty()
			return true
		}
	}
	return false
}

// Pin toggles the pinned state of an entry.
func (s *Store) Pin(id string, pinned bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.entries {
		if s.entries[i].ID == id {
			s.entries[i].Pinned = pinned
			s.markDirty()
			return true
		}
	}
	return false
}

// SetCategory sets the category tag for an entry.
func (s *Store) SetCategory(id, category string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.entries {
		if s.entries[i].ID == id {
			s.entries[i].Category = category
			s.markDirty()
			return true
		}
	}
	return false
}

// Use increments the use count and moves the entry to top.
func (s *Store) Use(id string) *ClipEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.entries {
		if s.entries[i].ID == id {
			s.entries[i].UsedCount++
			entry := s.entries[i]
			// Move to top
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			s.entries = append([]ClipEntry{entry}, s.entries...)
			s.markDirty()
			return &s.entries[0]
		}
	}
	return nil
}

// GetAll returns all entries (newest first).
func (s *Store) GetAll() []ClipEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ClipEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// GetRecent returns the most recent n entries.
func (s *Store) GetRecent(n int) []ClipEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.entries) {
		n = len(s.entries)
	}
	result := make([]ClipEntry, n)
	copy(result, s.entries[:n])
	return result
}

// Search returns entries matching the query string (case-insensitive).
func (s *Store) Search(query string) []ClipEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	var result []ClipEntry
	for _, e := range s.entries {
		if strings.Contains(strings.ToLower(e.Text), query) ||
			strings.Contains(strings.ToLower(e.Source), query) ||
			strings.Contains(strings.ToLower(e.Category), query) {
			result = append(result, e)
		}
	}
	return result
}

// GetByCategory returns entries with the given category.
func (s *Store) GetByCategory(category string) []ClipEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []ClipEntry
	for _, e := range s.entries {
		if e.Category == category {
			result = append(result, e)
		}
	}
	return result
}

// GetPinned returns all pinned entries.
func (s *Store) GetPinned() []ClipEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []ClipEntry
	for _, e := range s.entries {
		if e.Pinned {
			result = append(result, e)
		}
	}
	return result
}

// Categories returns a list of all unique category names.
func (s *Store) Categories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var cats []string
	for _, e := range s.entries {
		if e.Category != "" && !seen[e.Category] {
			seen[e.Category] = true
			cats = append(cats, e.Category)
		}
	}
	return cats
}

// Count returns the total number of entries.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// Config returns the current configuration.
func (s *Store) GetConfig() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetConfig updates the configuration and saves it.
func (s *Store) SetConfig(cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = cfg
	s.saveConfigLocked()
}

// ClearAll removes all non-pinned entries.
func (s *Store) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var pinned []ClipEntry
	for _, e := range s.entries {
		if e.Pinned {
			pinned = append(pinned, e)
		}
	}
	s.entries = pinned
	s.markDirty()
}

// SaveNow forces an immediate save to disk.
func (s *Store) SaveNow() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.saveTimer != nil {
		s.saveTimer.Stop()
		s.saveTimer = nil
	}
	s.saveHistoryLocked()
}

// Close saves any pending changes and stops the save timer.
func (s *Store) Close() {
	s.SaveNow()
}

// cleanup removes oldest non-pinned entries when over the max limit.
// Must be called with mu held.
func (s *Store) cleanup() {
	if len(s.entries) <= s.config.MaxEntries {
		return
	}

	// Keep pinned entries and remove oldest unpinned entries
	var kept []ClipEntry
	unpinnedCount := 0
	for _, e := range s.entries {
		if e.Pinned || unpinnedCount < s.config.MaxEntries {
			kept = append(kept, e)
			if !e.Pinned {
				unpinnedCount++
			}
		}
	}
	s.entries = kept
}

// markDirty schedules a debounced save (2 seconds).
func (s *Store) markDirty() {
	s.dirty = true
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}
	s.saveTimer = time.AfterFunc(2*time.Second, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.saveHistoryLocked()
	})
}

// historyData is the serialization wrapper for JSON persistence.
type historyData struct {
	NextID  int        `json:"next_id"`
	Entries []ClipEntry `json:"entries"`
}

func (s *Store) historyPath() string {
	return filepath.Join(s.dataDir, "history.json")
}

func (s *Store) configPath() string {
	return filepath.Join(s.dataDir, "config.json")
}

func (s *Store) loadHistory() {
	data, err := os.ReadFile(s.historyPath())
	if err != nil {
		return // file doesn't exist yet
	}

	var hd historyData
	if err := json.Unmarshal(data, &hd); err != nil {
		return
	}
	s.entries = hd.Entries
	s.nextID = hd.NextID
}

func (s *Store) saveHistoryLocked() {
	if !s.dirty {
		return
	}

	hd := historyData{
		NextID:  s.nextID,
		Entries: s.entries,
	}

	data, err := json.MarshalIndent(hd, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(s.historyPath(), data, 0644)
	s.dirty = false
}

func (s *Store) loadConfig() {
	data, err := os.ReadFile(s.configPath())
	if err != nil {
		return
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	s.config = cfg
}

func (s *Store) saveConfigLocked() {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(s.configPath(), data, 0644)
}
