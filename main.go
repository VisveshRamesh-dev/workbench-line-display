package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ---------- config types ----------

// Step is one instruction panel shown on one monitor. A part's procedure is an
// ordered list of these; monitor i shows the product's step (i+1).
type Step struct {
	Image       string `json:"image"`
	Description string `json:"description"`
}

type PartSpec struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

// DisplayConfig is one monitor / bench position in the line. The NUMBER of
// entries here is the number of screens — fully configurable.
type DisplayConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Config struct {
	CodePattern string              `json:"code_pattern"`
	ServerPort  int                 `json:"server_port"`
	Displays    []DisplayConfig     `json:"displays"`
	Parts       map[string]PartSpec `json:"parts"`
}

// ---------- events ----------

type Event struct {
	Type        string    `json:"type"` // "state" | "error"
	State       string    `json:"state,omitempty"`
	PartCode    string    `json:"part_code,omitempty"`
	PartName    string    `json:"part_name,omitempty"`
	Image       string    `json:"image,omitempty"`
	Description string    `json:"description,omitempty"`
	StepIndex   int       `json:"step_index,omitempty"`  // 1-based
	TotalSteps  int       `json:"total_steps,omitempty"`
	Message     string    `json:"message,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

func idleEvent() Event {
	return Event{Type: "state", State: "idle", Timestamp: time.Now()}
}

// ---------- display (one monitor; a view sink with its own SSE subscribers) ----------

type Display struct {
	Config DisplayConfig

	subMu       sync.Mutex
	subscribers map[chan Event]bool

	curMu   sync.Mutex
	current Event // replayed to a browser that connects mid-stream
}

func NewDisplay(dc DisplayConfig) *Display {
	return &Display{
		Config:      dc,
		subscribers: make(map[chan Event]bool),
		current:     idleEvent(),
	}
}

func (d *Display) Subscribe() chan Event {
	ch := make(chan Event, 8)
	d.subMu.Lock()
	d.subscribers[ch] = true
	d.subMu.Unlock()

	d.curMu.Lock()
	ev := d.current
	d.curMu.Unlock()
	ch <- ev // render correctly the instant the browser connects
	return ch
}

func (d *Display) Unsubscribe(ch chan Event) {
	d.subMu.Lock()
	if _, ok := d.subscribers[ch]; ok {
		delete(d.subscribers, ch)
		close(ch)
	}
	d.subMu.Unlock()
}

func (d *Display) broadcast(ev Event) {
	d.subMu.Lock()
	defer d.subMu.Unlock()
	for ch := range d.subscribers {
		select {
		case ch <- ev:
		default: // slow subscriber: drop rather than block the line
		}
	}
}

// Push updates the persistent state (replayed to new subscribers) and broadcasts.
func (d *Display) Push(ev Event) {
	d.curMu.Lock()
	d.current = ev
	d.curMu.Unlock()
	d.broadcast(ev)
}

// Flash broadcasts a transient event (e.g. an error) without changing state.
func (d *Display) Flash(ev Event) { d.broadcast(ev) }

func (d *Display) Current() Event {
	d.curMu.Lock()
	defer d.curMu.Unlock()
	return d.current
}

// ---------- the line (scan-advanced shift register) ----------

// slot is the product currently at one position in the line.
type slot struct {
	code  string
	name  string
	steps []Step
}

// Line is a shift register of length len(displayList). Each scan pushes a new
// product onto position 0 and shifts the rest forward; the product off the end
// is complete. Position i renders the product's step (i+1).
type Line struct {
	mu    sync.Mutex
	slots []*slot
}

var line Line

func (l *Line) init(n int) {
	l.slots = make([]*slot, n)
}

// Advance pushes a scanned product onto the line.
func (l *Line) Advance(code, name string, steps []Step) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(l.slots)
	if n == 0 {
		return
	}
	for i := n - 1; i >= 1; i-- {
		l.slots[i] = l.slots[i-1]
	}
	l.slots[0] = &slot{code: code, name: name, steps: steps}
	l.renderLocked()
}

// Reset clears the whole line (e.g. at shift start).
func (l *Line) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range l.slots {
		l.slots[i] = nil
	}
	l.renderLocked()
}

// renderLocked pushes each position's current step to its monitor.
func (l *Line) renderLocked() {
	now := time.Now()
	for i, d := range displayList {
		s := l.slots[i]
		if s == nil || i >= len(s.steps) {
			// empty, or this product has fewer steps than this position (done)
			ev := idleEvent()
			ev.Timestamp = now
			d.Push(ev)
			continue
		}
		st := s.steps[i]
		d.Push(Event{
			Type: "state", State: "display",
			PartCode: s.code, PartName: s.name,
			Image: st.Image, Description: st.Description,
			StepIndex: i + 1, TotalSteps: len(s.steps),
			Timestamp: now,
		})
	}
}

// PosSnapshot is the supervisor's view of one position in the line.
type PosSnapshot struct {
	Position    int    `json:"position"`
	DisplayID   string `json:"display_id"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	PartCode    string `json:"part_code,omitempty"`
	PartName    string `json:"part_name,omitempty"`
	Description string `json:"description,omitempty"`
	StepIndex   int    `json:"step_index,omitempty"`
	TotalSteps  int    `json:"total_steps,omitempty"`
}

func (l *Line) Snapshot() []PosSnapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]PosSnapshot, len(displayList))
	for i, d := range displayList {
		ps := PosSnapshot{Position: i + 1, DisplayID: d.Config.ID, DisplayName: d.Config.Name, State: "idle"}
		if s := l.slots[i]; s != nil && i < len(s.steps) {
			st := s.steps[i]
			ps.State = "display"
			ps.PartCode = s.code
			ps.PartName = s.name
			ps.Description = st.Description
			ps.StepIndex = i + 1
			ps.TotalSteps = len(s.steps)
		}
		out[i] = ps
	}
	return out
}

// ---------- globals ----------

var (
	cfg         Config
	codeRegex   *regexp.Regexp
	displays    = make(map[string]*Display)
	displayList []*Display
	logMu       sync.Mutex
	logFilePath string
)

// ---------- main ----------

func main() {
	if err := loadConfig("config.json"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	re, err := regexp.Compile(cfg.CodePattern)
	if err != nil {
		log.Fatalf("invalid code_pattern: %v", err)
	}
	codeRegex = re

	if len(cfg.Displays) < 1 {
		log.Fatalf("config error: at least one display is required")
	}
	for _, dc := range cfg.Displays {
		d := NewDisplay(dc)
		displays[dc.ID] = d
		displayList = append(displayList, d)
		log.Printf("display registered: %s (%s)", dc.ID, dc.Name)
	}
	line.init(len(displayList))

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("logs dir: %v", err)
	}
	logFilePath = filepath.Join("logs", fmt.Sprintf("scans-%s.csv", time.Now().Format("2006-01-02")))
	ensureLogHeader()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/display/", handleDisplayPage)
	mux.HandleFunc("/station/", handleDisplayPage) // backward-compatible alias
	mux.HandleFunc("/events/", handleSSE)
	mux.HandleFunc("/simulator", handleSimulatorPage)
	mux.HandleFunc("/supervisor", handleSupervisorPage)
	mux.HandleFunc("/api/scan", handleAPIScan)
	mux.HandleFunc("/api/reset", handleAPIReset)
	mux.HandleFunc("/api/config", handleAPIConfig)
	mux.HandleFunc("/api/snapshots", handleAPISnapshots)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Workbench Line Display (pipeline) listening on http://localhost%s", addr)
	log.Printf("Monitors: %d | Parts: %d | Log: %s", len(displayList), len(cfg.Parts), logFilePath)
	log.Printf("Open http://localhost%s/simulator to test", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = 8080
	}
	if cfg.CodePattern == "" {
		cfg.CodePattern = `P(\d{5})`
	}
	return nil
}

// ---------- HTTP handlers ----------

func serveNoCache(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	http.ServeFile(w, r, path)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	serveNoCache(w, r, filepath.Join("static", "index.html"))
}

func handleDisplayPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if _, ok := displays[id]; !ok {
		http.Error(w, "unknown display", http.StatusNotFound)
		return
	}
	serveNoCache(w, r, filepath.Join("static", "display.html"))
}

func handleSimulatorPage(w http.ResponseWriter, r *http.Request) {
	serveNoCache(w, r, filepath.Join("static", "simulator.html"))
}

func handleSupervisorPage(w http.ResponseWriter, r *http.Request) {
	serveNoCache(w, r, filepath.Join("static", "supervisor.html"))
}

func handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	type partOut struct {
		Code  string `json:"code"`
		Name  string `json:"name"`
		Steps int    `json:"steps"`
	}
	parts := make([]partOut, 0, len(cfg.Parts))
	for code, p := range cfg.Parts {
		parts = append(parts, partOut{Code: code, Name: p.Name, Steps: len(p.Steps)})
	}
	writeJSON(w, struct {
		Displays []DisplayConfig `json:"displays"`
		Parts    []partOut       `json:"parts"`
	}{Displays: cfg.Displays, Parts: parts})
}

func handleAPISnapshots(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, struct {
		Positions []PosSnapshot `json:"positions"`
	}{Positions: line.Snapshot()})
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/events/")
	d, ok := displays[id]
	if !ok {
		http.Error(w, "unknown display", http.StatusNotFound)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := d.Subscribe()
	defer d.Unsubscribe(ch)

	keepAlive := time.NewTicker(20 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-keepAlive.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// entryError flashes an error on the entry monitor (position 1), where the
// operator doing the scanning can see it. It does not advance the line.
func entryError(message, partCode string) {
	if len(displayList) == 0 {
		return
	}
	displayList[0].Flash(Event{Type: "error", Message: message, PartCode: partCode, Timestamp: time.Now()})
}

func handleAPIScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Raw       string `json:"raw"`
		StationID string `json:"station_id"` // accepted but ignored (one scanner feeds the line)
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	raw := strings.TrimSpace(req.Raw)
	if raw == "" {
		entryError("empty scan", "")
		logScan(raw, "", "ERROR", "empty scan")
		writeJSON(w, map[string]any{"ok": false, "error": "empty scan"})
		return
	}
	code := extractCode(raw)
	if code == "" {
		entryError("no part code found in scan", "")
		logScan(raw, "", "ERROR", "no part code found in scan")
		writeJSON(w, map[string]any{"ok": false, "error": "no part code found"})
		return
	}
	spec, found := cfg.Parts[code]
	if !found {
		entryError(fmt.Sprintf("unknown part code: %s", code), code)
		logScan(raw, code, "ERROR", "unknown part code")
		writeJSON(w, map[string]any{"ok": false, "part_code": code, "error": "unknown part"})
		return
	}
	if len(spec.Steps) == 0 {
		entryError(fmt.Sprintf("part %s has no steps configured", code), code)
		logScan(raw, code, "ERROR", "no steps configured")
		writeJSON(w, map[string]any{"ok": false, "part_code": code, "error": "no steps configured"})
		return
	}

	line.Advance(code, spec.Name, spec.Steps)
	logScan(raw, code, "ADVANCED", "")
	writeJSON(w, map[string]any{"ok": true, "part_code": code, "outcome": "ADVANCED", "total_steps": len(spec.Steps)})
}

func handleAPIReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	line.Reset()
	logScan("", "", "RESET", "")
	writeJSON(w, map[string]any{"ok": true})
}

// extractCode applies codeRegex to the raw scan. With a capture group, group 1
// is returned (so a structured ISO 15434 / H-KMC label is parsed down to the
// part number in the "P" field); otherwise the whole match.
func extractCode(s string) string {
	m := codeRegex.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	if len(m) > 1 && m[1] != "" {
		return m[1]
	}
	return m[0]
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// ---------- logging ----------

func ensureLogHeader() {
	logMu.Lock()
	defer logMu.Unlock()
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		f, err := os.Create(logFilePath)
		if err != nil {
			log.Printf("log create: %v", err)
			return
		}
		defer f.Close()
		f.WriteString("timestamp,raw_scan,part_code,outcome,error\n")
	}
}

func logScan(raw, partCode, outcome, errMsg string) {
	logMu.Lock()
	defer logMu.Unlock()

	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("log open: %v", err)
		return
	}
	defer f.Close()
	rawEsc := strings.ReplaceAll(raw, "\"", "\"\"")
	fmt.Fprintf(f, "%s,\"%s\",%s,%s,%s\n", time.Now().Format(time.RFC3339), rawEsc, partCode, outcome, errMsg)
}
