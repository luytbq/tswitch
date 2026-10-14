package tui

import (
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/luytbq/tswitch/internal/config"
	"github.com/luytbq/tswitch/internal/keys"
	"github.com/luytbq/tswitch/internal/tmux"
)

// clipboard holds a cut window or pane awaiting paste.
type clipboard struct {
	kind    string // "window" or "pane"
	srcSess string
	srcWin  int    // window index (both kinds)
	srcPane int    // pane kind only
	label   string // e.g. `window "editor" from work`
}

// Mode represents the current navigation level.
type Mode int

const (
	ModeSessionGrid Mode = iota
	ModeWindowGrid
	ModePaneGrid
)

// Model is the top-level Bubbletea model.
type Model struct {
	// Dependencies (injected via constructor).
	tmux      tmux.Service
	config    *config.Config
	appConfig *config.AppConfig
	styles    Styles

	// UI components.
	sessionGrid  *Grid
	windowGrid   *Grid
	paneGrid     *Grid
	tree         *Tree
	previewPanel *PreviewPanel

	// State.
	currentMode      Mode
	sessions         []tmux.Session
	windows          []tmux.Window
	panes            []tmux.Pane
	currentSess      string              // session name when in window view
	currentWin       int                 // window index when in pane view
	windowsBySession map[string][]string // session -> window names (for search)
	allWindows       map[string][]tmux.Window
	allPanes         map[string]map[int][]tmux.Pane
	clientSess       string // session this tmux client is attached to
	viewMode         string // config.ViewGrid | config.ViewList
	helpShown        bool
	markingMode      bool        // waiting for a mark-key press
	markingTarget    focusTarget // item the pending mark will point at
	filterMode       bool        // search input is active
	filterQuery      string      // current fuzzy-search term
	dialog           *Dialog
	pendingAction    dialogAction
	pendingTarget    focusTarget // item the open dialog acts on
	clipboard        *clipboard

	// Live preview.
	previewInterval time.Duration // 0 disables the periodic refetch
	previewGen      int           // bumped on every focus change; see captureResultMsg

	// Viewport.
	width  int
	height int

	// Feedback (status bar message with auto-expiry).
	statusMsg     string
	statusMsgTime time.Time
	isStatusError bool
}

// NewModel creates a Model wired to a real tmux client and default app config.
func NewModel(appCfg *config.AppConfig) (*Model, error) {
	return NewModelWith(tmux.NewClient(), appCfg)
}

// NewModelWith creates a Model using the given tmux.Service (useful for tests).
func NewModelWith(svc tmux.Service, appCfg *config.AppConfig) (*Model, error) {
	cfg, err := config.LoadState()
	if err != nil {
		cfg = config.Default()
	}
	if appCfg == nil {
		appCfg = config.DefaultAppConfig()
	}

	styles := NewStyles()
	m := &Model{
		tmux:            svc,
		config:          cfg,
		appConfig:       appCfg,
		styles:          styles,
		width:           80,
		height:          24,
		currentMode:     ModeSessionGrid,
		viewMode:        cfg.Settings.ViewMode,
		clientSess:      svc.CurrentSession(),
		previewInterval: appCfg.PreviewRefresh(),
	}

	gridW, gridH, previewW, previewH := m.layoutSizes()
	m.sessionGrid = NewGrid(gridW, gridH, styles)
	m.windowGrid = NewGrid(gridW, gridH, styles)
	m.paneGrid = NewGrid(gridW, gridH, styles)
	m.tree = NewTree(gridW, gridH, styles)
	if w := appCfg.UI.CardMinWidth; w > 0 {
		m.sessionGrid.SetMinCardWidth(w)
		m.windowGrid.SetMinCardWidth(w)
		m.paneGrid.SetMinCardWidth(w)
	}
	m.previewPanel = NewPreviewPanel(previewW, previewH, styles)
	if m.config.Settings.PreviewMode == config.PreviewModeMetadata {
		m.previewPanel.mode = PreviewMetadata
	}

	if err := m.loadSessions(); err != nil {
		m.setStatusError(err.Error())
	}
	m.focusInitialSession()

	return m, nil
}

// captureResultMsg carries the output of an async pane capture. The capture
// runs off the update loop, so a result can arrive after the cursor has already
// moved on; gen identifies the focus the request was issued for and anything
// older than the current focus is dropped rather than painted over it.
type captureResultMsg struct {
	gen     int
	content string
	// periodic marks a refresh nobody asked for, which keeps the last good
	// content instead of replacing the panel with an error line.
	periodic bool
}

// previewTickMsg drives the live preview. Exactly one tick is ever in flight:
// each one is rescheduled when it arrives.
type previewTickMsg struct{}

// fzfResultMsg carries the result of the fzf directory browser.
type fzfResultMsg struct {
	path string // selected directory path (empty if cancelled)
	err  error
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.syncPreview(), m.previewTick())
}

// Update implements tea.Model. It dispatches to focused handlers.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
	case captureResultMsg:
		if msg.gen != m.previewGen {
			return m, nil // issued for an item the cursor has since left
		}
		if msg.periodic && msg.content == "" {
			return m, nil
		}
		m.previewPanel.SetCaptureContent(msg.content)

	case previewTickMsg:
		return m, tea.Batch(m.refreshCapture(), m.previewTick())
	case fzfResultMsg:
		return m.handleFzfResult(msg)
	}
	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.dialog != nil {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.dialog.Render(m.width, m.height))
	}
	if m.helpShown {
		return m.renderHelp()
	}
	if m.isListView() {
		return m.renderListView()
	}
	switch m.currentMode {
	case ModeSessionGrid:
		return m.renderSessionView()
	case ModeWindowGrid:
		return m.renderWindowView()
	case ModePaneGrid:
		return m.renderPaneView()
	}
	return ""
}

// ---------------------------------------------------------------------------
// Key dispatch
// ---------------------------------------------------------------------------

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// Dialog intercepts all keys.
	if m.dialog != nil {
		return m.handleDialogKey(msg)
	}

	// Filter mode intercepts all keys.
	if m.filterMode {
		return m.handleFilterKey(msg)
	}

	// Marking mode intercepts all keys.
	if m.markingMode {
		return m.handleMarkAssignment(keyStr)
	}

	action := keys.Resolve(keyStr)

	// The help screen covers the whole terminal, so anything that would move
	// focus or mutate a session happens out of sight. Only the keys that leave
	// it are live.
	if m.helpShown {
		switch action {
		case keys.ActionToggleHelp, keys.ActionBack:
			m.helpShown = false
		case keys.ActionQuit:
			return m, tea.Quit
		}
		return m, nil
	}

	// The tree redefines navigation, folding and reordering; everything it
	// leaves alone falls through to the handlers shared with the grid.
	if m.isListView() {
		if model, cmd, handled := m.handleTreeAction(action); handled {
			return model, cmd
		}
	}

	switch action {
	case keys.ActionQuit:
		return m, tea.Quit

	case keys.ActionBack:
		return m.handleBack()

	case keys.ActionToggleHelp:
		m.helpShown = !m.helpShown

	case keys.ActionToggleView:
		return m.handleToggleView()

	case keys.ActionCycleSort:
		return m.handleCycleSort()

	case keys.ActionTogglePreview:
		m.previewPanel.ToggleMode()
		if m.previewPanel.IsCapture() {
			m.config.Settings.PreviewMode = config.PreviewModeCapture
		} else {
			m.config.Settings.PreviewMode = config.PreviewModeMetadata
		}
		_ = config.SaveState(m.config)
		return m, m.syncPreview()

	case keys.ActionStartMark:
		m.enterMarkingMode()

	case keys.ActionFilter:
		return m, m.enterFilterMode()

	case keys.ActionBrowseDirs:
		return m.handleBrowseDirs()

	case keys.ActionNew:
		return m.handleNew()

	case keys.ActionRename:
		return m.handleRename()

	case keys.ActionKill:
		return m.handleKill()

	case keys.ActionCut:
		return m.handleCut()

	case keys.ActionPaste:
		return m.handlePaste()

	case keys.ActionMoveUp:
		return m, m.moveFocus(0, -1)

	case keys.ActionMoveDown:
		return m, m.moveFocus(0, 1)
	case keys.ActionMoveLeft:
		return m, m.moveFocus(-1, 0)
	case keys.ActionMoveRight:
		return m, m.moveFocus(1, 0)

	case keys.ActionReorderUp:
		return m.handleReorder(0, -1)
	case keys.ActionReorderDown:
		return m.handleReorder(0, 1)
	case keys.ActionReorderLeft:
		return m.handleReorder(-1, 0)
	case keys.ActionReorderRight:
		return m.handleReorder(1, 0)

	case keys.ActionConfirm:
		return m.handleConfirm()

	case keys.ActionDirectSwitch:
		return m.handleQuickSwap()

	case keys.ActionQuickSwap:
		return m.handleQuickSwap()

	case keys.ActionNone:
		// Might be a mark-jump key.
		if m.config.HasMark(keyStr) {
			return m.handleJumpToMark(keyStr)
		}
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Data loading
// ---------------------------------------------------------------------------

func (m *Model) loadSessions() error {
	sessions, err := m.tmux.ListSessions()
	if err != nil {
		return err
	}

	// One tmux call each for every window and pane on the server: the tree
	// shows children of sessions it never drilled into, so fetching per
	// session would mean one call per visible row.
	if aw, err := m.tmux.ListAllWindows(); err == nil {
		m.allWindows = aw
	}
	if ap, err := m.tmux.ListAllPanes(); err == nil {
		m.allPanes = ap
	}

	m.windowsBySession = make(map[string][]string, len(m.allWindows))
	for name, windows := range m.allWindows {
		for _, w := range windows {
			m.windowsBySession[name] = append(m.windowsBySession[name], w.Name)
		}
	}
	for i := range sessions {
		total := 0
		for _, w := range m.allWindows[sessions[i].Name] {
			total += w.PaneCount
		}
		sessions[i].PaneCount = total
	}

	m.sessions = m.orderSessions(sessions)
	m.pruneExpanded()
	m.resetSessionItems()
	return nil
}

// resetSessionItems pushes the session list into both views, holding each one
// on the item it was already focused on.
func (m *Model) resetSessionItems() {
	focus := m.sessionGrid.FocusIndex()
	m.sessionGrid.SetItems(toGridItems(m.sessions, func(s tmux.Session) GridItem { return SessionCard{s} }))
	if focus >= len(m.sessions) {
		focus = len(m.sessions) - 1
	}
	m.sessionGrid.SetFocusIndex(focus)
	m.rebuildTree()
}

// focusInitialSession places the cursor for a freshly opened TUI. The tree is
// addressed by session name rather than by row number, because an expanded
// session puts its own windows in the rows right below it.
func (m *Model) focusInitialSession() {
	idx := m.initialFocusIndex()
	if idx >= len(m.sessions) {
		return
	}
	m.sessionGrid.SetFocusIndex(idx)
	name := m.sessions[idx].Name
	m.tree.FocusRowWhere(func(r TreeRow) bool {
		return r.Kind == NodeSession && r.Session == name
	})
}

// orderSessions arranges sessions per the active sort mode, then lifts the
// session this client is attached to to the top when pinning is on.
func (m *Model) orderSessions(sessions []tmux.Session) []tmux.Session {
	switch m.config.Settings.SortBy {
	case config.SortName:
		sort.SliceStable(sessions, func(i, j int) bool {
			return naturalLess(sessions[i].Name, sessions[j].Name)
		})
	case config.SortManual:
		sessions = m.applySavedSessionOrder(sessions)
	default: // config.SortRecent
		sort.SliceStable(sessions, func(i, j int) bool {
			return sessions[i].LastActive.After(sessions[j].LastActive)
		})
	}

	if m.config.PinCurrentFirst() && m.clientSess != "" {
		for i, s := range sessions {
			if s.Name != m.clientSess {
				continue
			}
			pinned := append([]tmux.Session{s}, sessions[:i]...)
			sessions = append(pinned, sessions[i+1:]...)
			break
		}
	}
	return sessions
}

// initialFocusIndex picks the row the cursor starts on. The current session is
// the one row the user will never pick, so when it sits at the top the cursor
// starts on the row below it: open, Enter, and you are back where you were.
func (m *Model) initialFocusIndex() int {
	if m.clientSess == "" {
		for i, s := range m.sessions {
			if s.Attached {
				return i
			}
		}
		return 0
	}
	if len(m.sessions) > 1 && m.sessions[0].Name == m.clientSess {
		return 1
	}
	for i, s := range m.sessions {
		if s.Name == m.clientSess {
			return i
		}
	}
	return 0
}

// pruneExpanded drops expansion state for sessions and windows tmux no longer
// has, so state.yaml does not collect dead entries.
func (m *Model) pruneExpanded() {
	live := make(map[string][]int, len(m.allWindows))
	for name, windows := range m.allWindows {
		indices := make([]int, 0, len(windows))
		for _, w := range windows {
			indices = append(indices, w.Index)
		}
		live[name] = indices
	}
	m.config.PruneExpanded(live)
}

// naturalLess compares names case-insensitively with digit runs compared as
// numbers, so session "2" sorts before "19" rather than after it.
func naturalLess(a, b string) bool {
	ar, br := []rune(strings.ToLower(a)), []rune(strings.ToLower(b))
	i, j := 0, 0
	for i < len(ar) && j < len(br) {
		if isDigit(ar[i]) && isDigit(br[j]) {
			si, sj := i, j
			for i < len(ar) && isDigit(ar[i]) {
				i++
			}
			for j < len(br) && isDigit(br[j]) {
				j++
			}
			na, _ := strconv.Atoi(string(ar[si:i]))
			nb, _ := strconv.Atoi(string(br[sj:j]))
			if na != nb {
				return na < nb
			}
			continue
		}
		if ar[i] != br[j] {
			return ar[i] < br[j]
		}
		i++
		j++
	}
	return len(ar)-i < len(br)-j
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func (m *Model) loadWindows(sessionName string) error {
	windows, err := m.tmux.ListWindows(sessionName)
	if err != nil {
		return err
	}
	windows = m.applySavedWindowOrder(sessionName, windows)
	m.windows = windows
	m.currentSess = sessionName

	m.windowGrid.SetItems(toGridItems(windows, func(w tmux.Window) GridItem { return WindowCard{w} }))
	return nil
}

func (m *Model) loadPanes(sessionName string, windowIndex int) error {
	panes, err := m.tmux.ListPanes(sessionName, windowIndex)
	if err != nil {
		return err
	}
	m.panes = panes
	m.currentWin = windowIndex

	m.paneGrid.SetItems(toGridItems(panes, func(p tmux.Pane) GridItem { return PaneCard{p} }))
	return nil
}

// toGridItems wraps a typed slice into []GridItem using the provided wrapper.
func toGridItems[T any](items []T, wrap func(T) GridItem) []GridItem {
	out := make([]GridItem, len(items))
	for i, item := range items {
		out[i] = wrap(item)
	}
	return out
}

// applyFilter re-filters the current mode's items from the full list and
// updates the grid. Called whenever filterQuery changes.
func (m *Model) applyFilter() {
	if m.isListView() {
		m.rebuildTree()
		return
	}
	switch m.currentMode {
	case ModeSessionGrid:
		filtered := FilterSessions(m.sessions, m.filterQuery, m.windowsBySession)
		m.sessionGrid.SetItems(toGridItems(filtered, func(s tmux.Session) GridItem { return SessionCard{s} }))
	case ModeWindowGrid:
		filtered := FilterWindows(m.windows, m.filterQuery)
		m.windowGrid.SetItems(toGridItems(filtered, func(w tmux.Window) GridItem { return WindowCard{w} }))
	case ModePaneGrid:
		filtered := FilterPanes(m.panes, m.filterQuery)
		m.paneGrid.SetItems(toGridItems(filtered, func(p tmux.Pane) GridItem { return PaneCard{p} }))
	}
}

// applySavedSessionOrder reorders sessions according to the saved order.
// Sessions not in the saved order are appended at the end.
func (m *Model) applySavedSessionOrder(sessions []tmux.Session) []tmux.Session {
	if len(m.config.SessionOrder) == 0 {
		return sessions
	}

	byName := make(map[string]tmux.Session, len(sessions))
	for _, s := range sessions {
		byName[s.Name] = s
	}

	result := make([]tmux.Session, 0, len(sessions))
	seen := make(map[string]bool)

	for _, name := range m.config.SessionOrder {
		if s, ok := byName[name]; ok {
			result = append(result, s)
			seen[name] = true
		}
	}
	for _, s := range sessions {
		if !seen[s.Name] {
			result = append(result, s)
		}
	}
	return result
}

// applySavedWindowOrder reorders windows according to the saved order.
// Windows not in the saved order are appended at the end.
func (m *Model) applySavedWindowOrder(sessionName string, windows []tmux.Window) []tmux.Window {
	order, ok := m.config.WindowOrder[sessionName]
	if !ok || len(order) == 0 {
		return windows
	}

	byIndex := make(map[int]tmux.Window, len(windows))
	for _, w := range windows {
		byIndex[w.Index] = w
	}

	result := make([]tmux.Window, 0, len(windows))
	seen := make(map[int]bool)

	for _, idx := range order {
		if w, ok := byIndex[idx]; ok {
			result = append(result, w)
			seen[idx] = true
		}
	}
	for _, w := range windows {
		if !seen[w.Index] {
			result = append(result, w)
		}
	}
	return result
}

// refreshSessions reloads session data, re-applies the current filter, and
// syncs the preview. Call after any operation that changes session state.
func (m *Model) refreshSessions() (tea.Model, tea.Cmd) {
	_ = m.loadSessions()
	m.applyFilter()
	return m, m.syncPreview()
}

// refreshWindows reloads window data for the current session, re-applies the
// current filter, and syncs the preview.
func (m *Model) refreshWindows() (tea.Model, tea.Cmd) {
	_ = m.loadWindows(m.currentSess)
	m.applyFilter()
	return m, m.syncPreview()
}

// resetFilter clears filter state. The grid already contains all items
// (no filter was applied when the mode was entered), so no SetItems call needed.
func (m *Model) resetFilter() {
	m.filterMode = false
	m.filterQuery = ""
}

// ---------------------------------------------------------------------------
// Resize
// ---------------------------------------------------------------------------

func (m *Model) resize(w, h int) {
	m.width = w
	m.height = h
	m.applyLayout()
}

// layoutSizes computes the width/height for the grid and preview panel.
//
// Session view: preview ~40%, grid ~60%.
// Window view:  50-50 split so there's room to show window content.
func (m *Model) layoutSizes() (gridW, gridH, previewContentW, previewH int) {
	const gap = 1           // space between grid and preview
	const previewBorder = 4 // border(1 each side=2) + Padding(1) horizontal(1 each side=2)

	previewPct := 40
	if m.currentMode == ModeWindowGrid || m.currentMode == ModePaneGrid {
		previewPct = 50
	}

	minW := minCardContentW
	if m.appConfig.UI.CardMinWidth > minW {
		minW = m.appConfig.UI.CardMinWidth
	}
	previewRendered := m.width * previewPct / 100
	if previewRendered < previewBorder+minW {
		previewRendered = previewBorder + minW
	}
	previewContentW = previewRendered - previewBorder

	gridW = m.width - previewRendered - gap
	if gridW < 1 {
		gridW = 1
	}

	bodyH := m.height - 3 // header(1) + separator(1) + statusbar(1)
	if bodyH < 1 {
		bodyH = 1
	}
	gridH = bodyH
	previewH = bodyH
	return
}

// applyLayout recalculates and pushes updated sizes to all components.
// Call after any mode switch in addition to terminal resize.
func (m *Model) applyLayout() {
	gridW, gridH, previewW, previewH := m.layoutSizes()
	m.sessionGrid.SetSize(gridW, gridH)
	m.windowGrid.SetSize(gridW, gridH)
	m.paneGrid.SetSize(gridW, gridH)
	m.tree.SetSize(gridW, gridH)

	// The grid may not use its full allocated width (integer division
	// remainder). Give the leftover to the preview so there's no gap.
	const gap = 1
	const previewBorder = 4
	extraW := gridW - m.mainUsedWidth()
	previewW += extraW
	if previewW < previewBorder {
		previewW = previewBorder
	}

	m.previewPanel.SetSize(previewW, previewH)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusMsgTime = time.Now()
	m.isStatusError = false
}

func (m *Model) setStatusError(msg string) {
	m.statusMsg = msg
	m.statusMsgTime = time.Now()
	m.isStatusError = true
}

// statusMessage returns the current status message, or empty if it has expired.
const statusMsgTTL = 5 * time.Second

func (m *Model) statusMessage() string {
	if m.statusMsg == "" {
		return ""
	}
	if time.Since(m.statusMsgTime) > statusMsgTTL {
		m.statusMsg = ""
		return ""
	}
	return m.statusMsg
}

// mainUsedWidth is the width the left pane actually occupies. Cards leave a
// remainder when the grid width does not divide evenly; the tree does not.
func (m *Model) mainUsedWidth() int {
	if m.isListView() {
		gridW, _, _, _ := m.layoutSizes()
		return gridW
	}
	return m.activeGrid().UsedWidth()
}

func (m *Model) activeGrid() *Grid {
	switch m.currentMode {
	case ModeWindowGrid:
		return m.windowGrid
	case ModePaneGrid:
		return m.paneGrid
	default:
		return m.sessionGrid
	}
}

// ---------------------------------------------------------------------------
// Focus target
// ---------------------------------------------------------------------------

// targetKind says which tmux object the user is acting on.
type targetKind int

const (
	targetSession targetKind = iota
	targetWindow
	targetPane
)

// focusTarget identifies the item under the cursor independently of the view
// mode. The grid derives it from the current drill-down level, the tree from
// the depth of the focused row, so handlers below need no view-mode branch.
type focusTarget struct {
	kind    targetKind
	session string
	window  int // -1 for session targets
	pane    int // -1 for session and window targets
	ok      bool
}

func (m *Model) focusedTarget() focusTarget {
	if m.isListView() {
		row, ok := m.tree.Focused()
		if !ok {
			return focusTarget{}
		}
		switch row.Kind {
		case NodeWindow:
			return focusTarget{kind: targetWindow, session: row.Session, window: row.Window, pane: -1, ok: true}
		case NodePane:
			return focusTarget{kind: targetPane, session: row.Session, window: row.Window, pane: row.Pane, ok: true}
		default:
			return focusTarget{kind: targetSession, session: row.Session, window: -1, pane: -1, ok: true}
		}
	}

	switch m.currentMode {
	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return focusTarget{}
		}
		return focusTarget{kind: targetWindow, session: m.currentSess, window: card.window.Index, pane: -1, ok: true}
	case ModePaneGrid:
		card, ok := m.paneGrid.GetFocused().(PaneCard)
		if !ok {
			return focusTarget{}
		}
		return focusTarget{kind: targetPane, session: m.currentSess, window: m.currentWin, pane: card.pane.Index, ok: true}
	default:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return focusTarget{}
		}
		return focusTarget{kind: targetSession, session: card.session.Name, window: -1, pane: -1, ok: true}
	}
}

func (m *Model) isListView() bool { return m.viewMode == config.ViewList }

func (m *Model) sessionByName(name string) (tmux.Session, bool) {
	for _, s := range m.sessions {
		if s.Name == name {
			return s, true
		}
	}
	return tmux.Session{}, false
}

func (m *Model) windowAt(session string, index int) (tmux.Window, bool) {
	for _, w := range m.windowsOf(session) {
		if w.Index == index {
			return w, true
		}
	}
	return tmux.Window{}, false
}

func (m *Model) paneAt(session string, windowIndex, paneIndex int) (tmux.Pane, bool) {
	if session == m.currentSess && windowIndex == m.currentWin {
		for _, p := range m.panes {
			if p.Index == paneIndex {
				return p, true
			}
		}
	}
	for _, p := range m.panesOf(session, windowIndex) {
		if p.Index == paneIndex {
			return p, true
		}
	}
	return tmux.Pane{}, false
}
