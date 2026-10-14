package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/luytbq/tswitch/internal/config"
	"github.com/luytbq/tswitch/internal/keys"
	"github.com/luytbq/tswitch/internal/tmux"
)

// ---------------------------------------------------------------------------
// Session / window management
// ---------------------------------------------------------------------------

type dialogAction int

const (
	dialogNone          dialogAction = iota
	dialogNewSession                 // input → tmux new-session
	dialogRenameSession              // input → tmux rename-session
	dialogKillSession                // confirm → tmux kill-session
	dialogNewWindow                  // input → tmux new-window
	dialogRenameWindow               // input → tmux rename-window
	dialogKillWindow                 // confirm → tmux kill-window
)

func (m *Model) handleNew() (tea.Model, tea.Cmd) {
	target := m.focusedTarget()
	if !target.ok {
		return m, nil
	}
	m.pendingTarget = target
	if target.kind == targetSession {
		m.dialog = NewInputDialog("New Session", "Session name:", "", m.styles)
		m.pendingAction = dialogNewSession
		return m, nil
	}
	// Below the session level, a new window belongs to the session in focus.
	m.dialog = NewInputDialog("New Window", "Window name:", "", m.styles)
	m.pendingAction = dialogNewWindow
	return m, nil
}

func (m *Model) handleRename() (tea.Model, tea.Cmd) {
	target := m.focusedTarget()
	if !target.ok {
		return m, nil
	}
	m.pendingTarget = target
	switch target.kind {
	case targetSession:
		m.dialog = NewInputDialog("Rename Session", "New name:", target.session, m.styles)
		m.pendingAction = dialogRenameSession
	case targetWindow:
		w, ok := m.windowAt(target.session, target.window)
		if !ok {
			return m, nil
		}
		m.dialog = NewInputDialog("Rename Window", "New name:", w.Name, m.styles)
		m.pendingAction = dialogRenameWindow
	case targetPane:
		// No-op: pane renaming is not supported.
	}
	return m, nil
}

func (m *Model) handleKill() (tea.Model, tea.Cmd) {
	target := m.focusedTarget()
	if !target.ok {
		return m, nil
	}
	m.pendingTarget = target
	switch target.kind {
	case targetSession:
		m.dialog = NewConfirmDialog("Kill Session",
			fmt.Sprintf("Kill session %q?", target.session), m.styles)
		m.pendingAction = dialogKillSession
	case targetWindow:
		w, ok := m.windowAt(target.session, target.window)
		if !ok {
			return m, nil
		}
		m.dialog = NewConfirmDialog("Kill Window",
			fmt.Sprintf("Kill window %q?", w.Name), m.styles)
		m.pendingAction = dialogKillWindow
	case targetPane:
		// No-op: pane killing is not supported.
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Dialog key handling
// ---------------------------------------------------------------------------

func (m *Model) handleDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	d := m.dialog
	switch d.Kind {
	case DialogInput:
		switch msg.String() {
		case "esc":
			m.dialog = nil
		case "enter":
			return m.submitDialog()
		case "backspace":
			if len(d.Input) > 0 {
				runes := []rune(d.Input)
				d.Input = string(runes[:len(runes)-1])
			}
		default:
			if msg.Type == tea.KeyRunes {
				d.Input += msg.String()
			}
		}

	case DialogConfirm:
		switch msg.String() {
		case "esc":
			m.dialog = nil
		case "y":
			d.SelectedIdx = 0
			return m.submitDialog()
		case "n":
			m.dialog = nil
		case "enter":
			return m.submitDialog()
		case "left", "h", "right", "l":
			if len(d.Options) == 2 {
				d.SelectedIdx = 1 - d.SelectedIdx
			}
		}
	}
	return m, nil
}

func (m *Model) submitDialog() (tea.Model, tea.Cmd) {
	d := m.dialog
	action := m.pendingAction
	m.dialog = nil
	m.pendingAction = dialogNone

	switch action {
	case dialogNewSession:
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("session name cannot be empty")
			return m, nil
		}
		if err := m.tmux.NewSession(name); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus("Created: " + name)
		_ = m.loadSessions()
		m.applyFilter()
		m.sessionGrid.FocusFirstWhere(func(item GridItem) bool { return item.Title() == name })
		return m, m.syncPreview()

	case dialogRenameSession:
		oldName := m.pendingTarget.session
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("session name cannot be empty")
			return m, nil
		}
		if err := m.tmux.RenameSession(oldName, name); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		// Saved state is keyed by session name, so it has to follow the rename.
		m.config.RenameSessionState(oldName, name)
		m.saveState()
		m.setStatus("Renamed to: " + name)
		return m.refreshSessions()

	case dialogKillSession:
		if d.SelectedIdx != 0 {
			return m, nil // "No" selected
		}
		name := m.pendingTarget.session
		if err := m.tmux.KillSession(name); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus("Killed: " + name)
		return m.refreshSessions()

	case dialogNewWindow:
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("window name cannot be empty")
			return m, nil
		}
		if err := m.tmux.NewWindow(m.pendingTarget.session, name); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus("Created: " + name)
		return m.refreshAfterWindowChange(name)

	case dialogRenameWindow:
		target := m.pendingTarget
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("window name cannot be empty")
			return m, nil
		}
		if err := m.tmux.RenameWindow(target.session, target.window, name); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus("Renamed to: " + name)
		return m.refreshAfterWindowChange("")

	case dialogKillWindow:
		if d.SelectedIdx != 0 {
			return m, nil // "No" selected
		}
		target := m.pendingTarget
		w, ok := m.windowAt(target.session, target.window)
		if !ok {
			return m, nil
		}
		if err := m.tmux.KillWindow(target.session, target.window); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus("Killed: " + w.Name)
		return m.refreshAfterWindowChange("")
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Filter / Search
// ---------------------------------------------------------------------------

func (m *Model) enterFilterMode() tea.Cmd {
	m.filterMode = true
	m.filterQuery = ""
	m.applyFilter()
	return m.syncPreview()
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filterMode = false
		m.filterQuery = ""
		m.applyFilter()
		return m, m.syncPreview()
	case "enter":
		m.filterMode = false
	case "backspace":
		if len(m.filterQuery) > 0 {
			runes := []rune(m.filterQuery)
			m.filterQuery = string(runes[:len(runes)-1])
			m.applyFilter()
			return m, m.syncPreview()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.filterQuery += msg.String()
			m.applyFilter()
			return m, m.syncPreview()
		}
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

// handleBack navigates backwards: pane->window->session->quit.
func (m *Model) handleBack() (tea.Model, tea.Cmd) {
	switch {
	case m.markingMode:
		m.markingMode = false
		m.setStatus("")
	case m.currentMode == ModePaneGrid:
		m.resetFilter()
		m.currentMode = ModeWindowGrid
		m.applyLayout()
		return m, m.syncPreview()
	case m.currentMode == ModeWindowGrid:
		m.resetFilter()
		m.currentMode = ModeSessionGrid
		m.applyLayout()
		return m, m.syncPreview()
	default:
		return m, tea.Quit
	}
	return m, nil
}

// handleConfirm drills into children. At the leaf level (panes), it switches.
// Single-child skipping: single-window sessions skip to panes (or switch if
// also single-pane); single-pane windows switch immediately.
func (m *Model) handleConfirm() (tea.Model, tea.Cmd) {
	switch m.currentMode {
	case ModeSessionGrid:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		// Single-window session: skip window grid, drill into panes directly.
		if card.session.WindowCount == 1 {
			if err := m.loadWindows(card.session.Name); err != nil {
				m.setStatusError(err.Error())
				return m, nil
			}
			// The single window — check if it also has a single pane.
			if len(m.windows) == 1 {
				w := m.windows[0]
				if w.PaneCount <= 1 {
					// Single window, single pane: switch immediately.
					if err := m.switchTo(focusTarget{kind: targetPane, session: card.session.Name, window: w.Index}); err != nil {
						m.setStatusError(err.Error())
						return m, nil
					}
					return m, tea.Quit
				}
				// Single window, multiple panes: skip to pane grid.
				if err := m.loadPanes(card.session.Name, w.Index); err != nil {
					m.setStatusError(err.Error())
					return m, nil
				}
				m.resetFilter()
				m.currentMode = ModePaneGrid
				m.applyLayout()
				return m, m.syncPreview()
			}
		}
		if err := m.loadWindows(card.session.Name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.resetFilter()
			m.currentMode = ModeWindowGrid
			m.applyLayout()
			return m, m.syncPreview()
		}

	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		// Single-pane window: switch immediately.
		if card.window.PaneCount <= 1 {
			if err := m.switchTo(focusTarget{kind: targetPane, session: m.currentSess, window: card.window.Index}); err != nil {
				m.setStatusError(err.Error())
			} else {
				return m, tea.Quit
			}
			return m, nil
		}
		if err := m.loadPanes(m.currentSess, card.window.Index); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.resetFilter()
			m.currentMode = ModePaneGrid
			m.applyLayout()
			return m, m.syncPreview()
		}

	case ModePaneGrid:
		// Leaf level: Enter acts same as Space (switch & quit).
		return m.handleQuickSwap()
	}
	return m, nil
}

// switchTo moves the client onto the target. Selecting a pane only changes
// which pane is active inside its own window, so the client still has to be
// switched for the user to end up there.
func (m *Model) switchTo(target focusTarget) error {
	switch target.kind {
	case targetPane:
		if err := m.tmux.SelectPane(target.session, target.window, target.pane); err != nil {
			return err
		}
		return m.tmux.SwitchClient(target.session, target.window)
	case targetWindow:
		return m.tmux.SwitchClient(target.session, target.window)
	default:
		return m.tmux.SwitchToSession(target.session)
	}
}

// handleQuickSwap switches to the focused item at any level.
func (m *Model) handleQuickSwap() (tea.Model, tea.Cmd) {
	target := m.focusedTarget()
	if !target.ok {
		return m, nil
	}

	if s, ok := m.sessionByName(target.session); ok && s.WindowCount == 0 {
		return m, nil
	}

	if err := m.switchTo(target); err != nil {
		m.setStatusError(err.Error())
		return m, nil
	}
	return m, tea.Quit
}

// ---------------------------------------------------------------------------
// Marks
// ---------------------------------------------------------------------------

// enterMarkingMode puts the model into mark-assignment mode.
func (m *Model) enterMarkingMode() {
	m.markingMode = true
	m.markingTarget = m.focusedTarget()
}

// handleMarkAssignment processes a key press while in marking mode.
func (m *Model) handleMarkAssignment(keyStr string) (tea.Model, tea.Cmd) {
	defer func() { m.markingMode = false }()

	// Cancel on ESC.
	if keyStr == "esc" {
		m.setStatus("")
		return m, nil
	}

	if keys.IsReserved(keyStr) {
		m.setStatusError(fmt.Sprintf("'%s' is reserved", keyStr))
		return m, nil
	}

	target := m.markingTarget
	if !target.ok {
		return m, nil
	}

	switch target.kind {
	case targetSession:
		// Remove any existing mark pointing to the same target before setting new one.
		m.config.RemoveMarksForTarget(target.session, -1)
		m.config.SetMark(keyStr, target.session, -1, 0)
		if err := config.SaveState(m.config); err != nil {
			m.setStatusError(fmt.Sprintf("Failed to save: %v", err))
		} else {
			m.setStatus(fmt.Sprintf("Marked %s -> [%s]", target.session, keyStr))
		}

	case targetWindow:
		m.config.RemoveMarksForTarget(target.session, target.window)
		m.config.SetMark(keyStr, target.session, target.window, 0)
		if err := config.SaveState(m.config); err != nil {
			m.setStatusError(fmt.Sprintf("Failed to save: %v", err))
		} else {
			m.setStatus(fmt.Sprintf("Marked %s:%d -> [%s]", target.session, target.window, keyStr))
		}

	case targetPane:
		m.config.SetMark(keyStr, target.session, target.window, target.pane)
		if err := config.SaveState(m.config); err != nil {
			m.setStatusError(fmt.Sprintf("Failed to save: %v", err))
		} else {
			m.setStatus(fmt.Sprintf("Marked %s:%d.%d -> [%s]", target.session, target.window, target.pane, keyStr))
		}
	}

	return m, nil
}

// handleJumpToMark switches to the target of a mark.
func (m *Model) handleJumpToMark(keyStr string) (tea.Model, tea.Cmd) {
	mark := m.config.GetMark(keyStr)
	if mark == nil {
		m.setStatusError(fmt.Sprintf("no mark '%s'", keyStr))
		return m, nil
	}

	target := focusTarget{kind: targetSession, session: mark.SessionName, window: -1, pane: -1}
	switch {
	case mark.PaneIndex > 0:
		target.kind = targetPane
		target.window = mark.WindowIndex
		target.pane = mark.PaneIndex
	case mark.WindowIndex >= 0:
		target.kind = targetWindow
		target.window = mark.WindowIndex
	}
	if err := m.switchTo(target); err != nil {
		m.setStatusError(err.Error())
		return m, nil
	}
	return m, tea.Quit
}

// ---------------------------------------------------------------------------
// Preview sync
// ---------------------------------------------------------------------------

// syncPreview updates the preview panel for the currently focused item.
// In capture mode it clears stale content and returns a tea.Cmd that fetches
// the pane capture asynchronously; in metadata mode it updates synchronously
// and returns nil.
func (m *Model) syncPreview() tea.Cmd {
	if m.previewPanel.IsCapture() {
		m.previewPanel.SetCaptureContent("") // clear stale content
		return m.fetchCapture()
	}

	target := m.focusedTarget()
	if !target.ok {
		return nil
	}
	switch target.kind {
	case targetWindow:
		if w, ok := m.windowAt(target.session, target.window); ok {
			m.previewPanel.SetWindowMetadata(w)
		}
	case targetPane:
		if p, ok := m.paneAt(target.session, target.window, target.pane); ok {
			m.previewPanel.SetPaneMetadata(p)
		}
	default:
		if sess, ok := m.sessionByName(target.session); ok {
			m.previewPanel.SetSessionMetadata(sess)
		}
	}
	return nil
}

// fetchCapture returns a Cmd that runs tmux capture-pane for the focused item
// in a goroutine and delivers the result as a captureResultMsg.
func (m *Model) fetchCapture() tea.Cmd {
	target := m.focusedTarget()
	if !target.ok {
		return nil
	}

	sessName := target.session
	winIdx := target.window // -1 on a session target: capture its active window
	paneIdx := target.pane
	if paneIdx < 0 {
		paneIdx = 0
	}

	return func() tea.Msg {
		content, err := m.tmux.CapturePane(sessName, winIdx, paneIdx)
		if err != nil {
			return captureResultMsg{"(capture error: " + err.Error() + ")"}
		}
		// tmux pads every line with spaces to the full pane width and may
		// include \r; strip both so the preview box doesn't overflow.
		content = strings.ReplaceAll(content, "\r", "")
		rawLines := strings.Split(content, "\n")
		for i, l := range rawLines {
			rawLines[i] = strings.TrimRight(l, " ")
		}
		return captureResultMsg{strings.Join(rawLines, "\n")}
	}
}

// moveFocus moves the active grid's focus and syncs the preview.
func (m *Model) moveFocus(dx, dy int) tea.Cmd {
	m.activeGrid().MoveFocus(dx, dy)
	return m.syncPreview()
}

// ---------------------------------------------------------------------------
// Cut / Paste (move window / pane)
// ---------------------------------------------------------------------------

// handleCut captures the focused window or pane onto m.clipboard. A second
// press while the clipboard is already set clears it (toggle).
func (m *Model) handleCut() (tea.Model, tea.Cmd) {
	if m.clipboard != nil {
		m.clipboard = nil
		m.setStatus("cut cleared")
		return m, nil
	}

	target := m.focusedTarget()
	if !target.ok {
		return m, nil
	}

	switch target.kind {
	case targetWindow:
		w, ok := m.windowAt(target.session, target.window)
		if !ok {
			return m, nil
		}
		m.clipboard = &clipboard{
			kind:    "window",
			srcSess: target.session,
			srcWin:  target.window,
			label:   fmt.Sprintf("window %q from %s", w.Name, target.session),
		}
		m.setStatus("cut: " + m.clipboard.label)

	case targetPane:
		m.clipboard = &clipboard{
			kind:    "pane",
			srcSess: target.session,
			srcWin:  target.window,
			srcPane: target.pane,
			label:   fmt.Sprintf("pane %d from %s:%d", target.pane, target.session, target.window),
		}
		m.setStatus("cut: " + m.clipboard.label)

	case targetSession:
		m.setStatusError("nothing to cut here")
	}
	return m, nil
}

// handlePaste commits the clipboard onto the focused destination, validating
// that the current mode matches the clipboard kind.
func (m *Model) handlePaste() (tea.Model, tea.Cmd) {
	cb := m.clipboard
	if cb == nil {
		m.setStatusError("clipboard is empty")
		return m, nil
	}

	target := m.focusedTarget()
	if !target.ok {
		return m, nil
	}

	switch cb.kind {
	case "window":
		if target.kind != targetSession {
			m.setStatusError("paste a window onto a session")
			return m, nil
		}
		if target.session == cb.srcSess {
			m.setStatusError("already in this session")
			return m, nil
		}
		if err := m.tmux.MoveWindow(cb.srcSess, cb.srcWin, target.session); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus(fmt.Sprintf("moved %s to %s", cb.label, target.session))

	case "pane":
		if target.kind != targetWindow {
			m.setStatusError("paste a pane onto a window")
			return m, nil
		}
		if target.session == cb.srcSess && target.window == cb.srcWin {
			m.setStatusError("already in this window")
			return m, nil
		}
		if err := m.tmux.JoinPane(cb.srcSess, cb.srcWin, cb.srcPane, target.session, target.window); err != nil {
			m.setStatusError(err.Error())
			return m, nil
		}
		m.setStatus(fmt.Sprintf("moved %s to %s:%d", cb.label, target.session, target.window))
	}

	m.clipboard = nil
	_ = m.loadSessions()
	if !m.isListView() && m.currentMode != ModeSessionGrid {
		_ = m.loadWindows(m.currentSess)
	}
	m.applyFilter()
	return m, m.syncPreview()
}

// handleReorder swaps the focused item with its neighbor and persists the new order.
func (m *Model) handleReorder(dx, dy int) (tea.Model, tea.Cmd) {
	if m.currentMode == ModePaneGrid {
		return m, nil // reordering panes is not supported
	}

	grid := m.activeGrid()

	// Extract and persist the new order.
	switch m.currentMode {
	case ModeSessionGrid:
		if m.config.Settings.SortBy != config.SortManual {
			m.setStatusError("reorder needs sort: manual (press s)")
			return m, nil
		}
		if !grid.MoveItem(dx, dy) {
			return m, nil
		}
		items := grid.Items()
		order := make([]string, len(items))
		for i, item := range items {
			order[i] = item.(SessionCard).session.Name
		}
		m.config.SetSessionOrder(order)
		// Update m.sessions to match new order.
		m.sessions = make([]tmux.Session, len(items))
		for i, item := range items {
			m.sessions[i] = item.(SessionCard).session
		}

	case ModeWindowGrid:
		// Capture the two windows before the visual swap so we can call swap-window.
		oldFocusPos := grid.FocusIndex()
		srcCard := grid.Items()[oldFocusPos].(WindowCard)

		if !grid.MoveItem(dx, dy) {
			return m, nil
		}

		newFocusPos := grid.FocusIndex()
		dstCard := grid.Items()[oldFocusPos].(WindowCard)

		// Swap the windows in TMUX to keep display order in sync with actual order.
		if err := m.tmux.SwapWindow(m.currentSess, srcCard.window.Index, dstCard.window.Index); err != nil {
			m.setStatusError(err.Error())
			// Undo the visual swap so the grid stays consistent with TMUX.
			grid.MoveItem(-dx, -dy)
			return m, nil
		}

		// After swap-window, each window now occupies the other's index.
		updatedSrc := srcCard.window
		updatedSrc.Index = dstCard.window.Index
		updatedDst := dstCard.window
		updatedDst.Index = srcCard.window.Index
		grid.ReplaceItem(newFocusPos, WindowCard{window: updatedSrc})
		grid.ReplaceItem(oldFocusPos, WindowCard{window: updatedDst})

		// Update m.windows to match new order and indices.
		items := grid.Items()
		m.windows = make([]tmux.Window, len(items))
		for i, item := range items {
			m.windows[i] = item.(WindowCard).window
		}

		// TMUX is now the source of truth; clear any saved visual override.
		m.config.ClearWindowOrder(m.currentSess)
	}

	if err := config.SaveState(m.config); err != nil {
		m.setStatusError(err.Error())
	}
	m.rebuildTree()
	return m, m.syncPreview()
}

// ---------------------------------------------------------------------------
// View mode / sort
// ---------------------------------------------------------------------------

// saveState persists runtime state, reporting failures in the status bar.
func (m *Model) saveState() {
	if err := config.SaveState(m.config); err != nil {
		m.setStatusError(err.Error())
	}
}

// refreshAfterWindowChange reloads after a window was created, renamed or
// killed, optionally putting the cursor on a window by name.
func (m *Model) refreshAfterWindowChange(focusName string) (tea.Model, tea.Cmd) {
	_ = m.loadSessions()
	inWindowGrid := !m.isListView() && m.currentMode == ModeWindowGrid
	if inWindowGrid {
		_ = m.loadWindows(m.currentSess)
	}
	m.applyFilter()
	if focusName != "" && inWindowGrid {
		m.windowGrid.FocusFirstWhere(func(item GridItem) bool {
			return strings.HasSuffix(item.Title(), ": "+focusName)
		})
	}
	return m, m.syncPreview()
}

// handleToggleView swaps between the card grid and the tree list, keeping the
// cursor on the same session.
func (m *Model) handleToggleView() (tea.Model, tea.Cmd) {
	target := m.focusedTarget()

	if m.isListView() {
		m.viewMode = config.ViewGrid
	} else {
		m.viewMode = config.ViewList
	}
	// The tree spans all three levels on its own, so leaving the grid means
	// leaving its drill-down state behind.
	m.currentMode = ModeSessionGrid
	m.config.Settings.ViewMode = m.viewMode
	m.saveState()

	m.applyLayout()
	m.applyFilter()

	if target.ok {
		if m.isListView() {
			m.tree.FocusRowWhere(func(r TreeRow) bool {
				return r.Kind == NodeSession && r.Session == target.session
			})
		} else {
			m.sessionGrid.FocusFirstWhere(func(item GridItem) bool { return item.Title() == target.session })
		}
	}
	return m, m.syncPreview()
}

// handleCycleSort rotates the session sort mode and re-sorts in place.
func (m *Model) handleCycleSort() (tea.Model, tea.Cmd) {
	switch m.config.Settings.SortBy {
	case config.SortRecent:
		m.config.Settings.SortBy = config.SortName
	case config.SortName:
		m.config.Settings.SortBy = config.SortManual
	default:
		m.config.Settings.SortBy = config.SortRecent
	}
	m.saveState()

	target := m.focusedTarget()
	m.sessions = m.orderSessions(m.sessions)
	m.resetSessionItems()
	m.applyFilter()
	if target.ok {
		m.tree.FocusRowWhere(func(r TreeRow) bool {
			return r.Kind == NodeSession && r.Session == target.session
		})
		m.sessionGrid.FocusFirstWhere(func(item GridItem) bool { return item.Title() == target.session })
	}
	m.setStatus("sort: " + m.config.Settings.SortBy)
	return m, m.syncPreview()
}

// ---------------------------------------------------------------------------
// Tree navigation
// ---------------------------------------------------------------------------

// handleTreeAction handles the actions whose meaning changes in the tree view.
// It reports whether it consumed the action; everything else falls through to
// the shared handlers, which work off the focused target either way.
func (m *Model) handleTreeAction(action keys.Action) (tea.Model, tea.Cmd, bool) {
	switch action {
	case keys.ActionMoveUp:
		m.tree.MoveFocus(-1)
		return m, m.syncPreview(), true

	case keys.ActionMoveDown:
		m.tree.MoveFocus(1)
		return m, m.syncPreview(), true

	case keys.ActionMoveLeft:
		row, ok := m.tree.Focused()
		if !ok {
			return m, nil, true
		}
		if row.Expanded {
			return m.setRowExpanded(row, false)
		}
		m.tree.FocusParent()
		return m, m.syncPreview(), true

	case keys.ActionMoveRight:
		row, ok := m.tree.Focused()
		if !ok {
			return m, nil, true
		}
		if row.Expandable && !row.Expanded {
			return m.setRowExpanded(row, true)
		}
		if row.Expanded {
			m.tree.MoveFocus(1) // step onto the first child
		}
		return m, m.syncPreview(), true

	case keys.ActionConfirm:
		row, ok := m.tree.Focused()
		if !ok {
			return m, nil, true
		}
		if !row.Expandable {
			// Nothing to unfold: treat it as "take me there".
			model, cmd := m.handleQuickSwap()
			return model, cmd, true
		}
		return m.setRowExpanded(row, !row.Expanded)

	case keys.ActionReorderLeft:
		m.config.CollapseAll()
		m.saveState()
		m.rebuildTree()
		return m, m.syncPreview(), true

	case keys.ActionReorderRight:
		for _, s := range m.sessions {
			if len(m.windowsOf(s.Name)) > 0 {
				m.config.SetSessionExpanded(s.Name, true)
			}
		}
		m.saveState()
		m.rebuildTree()
		return m, m.syncPreview(), true

	case keys.ActionReorderUp:
		return m.handleTreeReorder(-1)

	case keys.ActionReorderDown:
		return m.handleTreeReorder(1)

	case keys.ActionBack:
		if row, ok := m.tree.Focused(); ok && row.Expanded {
			return m.setRowExpanded(row, false)
		}
		return m, nil, false
	}
	return m, nil, false
}

// setRowExpanded opens or closes a node and persists the new state.
func (m *Model) setRowExpanded(row TreeRow, expanded bool) (tea.Model, tea.Cmd, bool) {
	switch row.Kind {
	case NodeSession:
		m.config.SetSessionExpanded(row.Session, expanded)
	case NodeWindow:
		m.config.SetWindowExpanded(row.Session, row.Window, expanded)
	default:
		return m, nil, true
	}
	m.saveState()
	m.rebuildTree()
	return m, m.syncPreview(), true
}

// handleTreeReorder moves the focused row among its siblings.
func (m *Model) handleTreeReorder(delta int) (tea.Model, tea.Cmd, bool) {
	row, ok := m.tree.Focused()
	if !ok {
		return m, nil, true
	}

	switch row.Kind {
	case NodeSession:
		if m.config.Settings.SortBy != config.SortManual {
			m.setStatusError("reorder needs sort: manual (press s)")
			return m, nil, true
		}
		src := -1
		for i, s := range m.sessions {
			if s.Name == row.Session {
				src = i
				break
			}
		}
		dst := src + delta
		if src < 0 || dst < 0 || dst >= len(m.sessions) {
			return m, nil, true
		}
		m.sessions[src], m.sessions[dst] = m.sessions[dst], m.sessions[src]
		order := make([]string, len(m.sessions))
		for i, s := range m.sessions {
			order[i] = s.Name
		}
		m.config.SetSessionOrder(order)
		m.saveState()
		m.resetSessionItems()
		m.applyFilter()
		m.tree.FocusRowWhere(func(r TreeRow) bool {
			return r.Kind == NodeSession && r.Session == row.Session
		})
		return m, m.syncPreview(), true

	case NodeWindow:
		windows := m.windowsOf(row.Session)
		src := -1
		for i, w := range windows {
			if w.Index == row.Window {
				src = i
				break
			}
		}
		dst := src + delta
		if src < 0 || dst < 0 || dst >= len(windows) {
			return m, nil, true
		}
		dstIndex := windows[dst].Index
		if err := m.tmux.SwapWindow(row.Session, row.Window, dstIndex); err != nil {
			m.setStatusError(err.Error())
			return m, nil, true
		}
		// tmux is the source of truth for window order after a swap.
		m.config.ClearWindowOrder(row.Session)
		m.saveState()
		_ = m.loadSessions()
		m.applyFilter()
		// The focused window now sits at the index it was swapped with.
		m.tree.FocusRowWhere(func(r TreeRow) bool {
			return r.Kind == NodeWindow && r.Session == row.Session && r.Window == dstIndex
		})
		return m, m.syncPreview(), true
	}
	return m, nil, true
}
