package keys

// Action represents a user action triggered by a key press.
type Action int

const (
	ActionNone Action = iota

	// Navigation
	ActionMoveUp
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight

	// Selection
	ActionConfirm      // o - drill into child view
	ActionQuickSwap    // space - quick switch
	ActionDirectSwitch // enter - switch to active pane immediately
	ActionBack         // esc - go back or quit

	// Marks
	ActionStartMark // m - enter marking mode

	// Management (future)
	ActionNew    // n
	ActionRename // r
	ActionKill   // d - delete (moved from x)
	ActionCut    // x - cut window/pane to clipboard
	ActionPaste  // p - paste clipboard onto focused destination
	ActionTag    // t

	// Reorder
	ActionReorderUp
	ActionReorderDown
	ActionReorderLeft
	ActionReorderRight

	// Browse
	ActionBrowseDirs // f

	// UI
	ActionTogglePreview // tab
	ActionToggleHelp    // ?
	ActionFilter        // /
	ActionQuit          // q
)

// reservedKeys are keys that cannot be used as mark assignments.
// Kept as a map for O(1) lookup.
var reservedKeys = map[string]bool{
	"esc": true, "enter": true, " ": true, "tab": true,
	"up": true, "down": true, "left": true, "right": true,
	"j": true, "k": true, "h": true, "l": true,
	"?": true, "q": true, "m": true, "/": true,
	"f": true, "o": true,
	"n": true, "r": true, "d": true, "x": true, "p": true, "t": true,
	"H": true, "J": true, "K": true, "L": true,
}

// defaultKeymap maps key strings to actions.
var defaultKeymap = map[string]Action{
	"up": ActionMoveUp, "k": ActionMoveUp,
	"down": ActionMoveDown, "j": ActionMoveDown,
	"left": ActionMoveLeft, "h": ActionMoveLeft,
	"right": ActionMoveRight, "l": ActionMoveRight,

	"K": ActionReorderUp, "J": ActionReorderDown,
	"H": ActionReorderLeft, "L": ActionReorderRight,

	"o":     ActionConfirm,
	"enter": ActionDirectSwitch,
	" ":     ActionQuickSwap,
	"esc":   ActionBack,

	"m": ActionStartMark,

	"f": ActionBrowseDirs,
	"n": ActionNew,
	"r": ActionRename,
	"d": ActionKill,
	"x": ActionCut,
	"p": ActionPaste,
	"t": ActionTag,

	"tab": ActionTogglePreview,
	"?":   ActionToggleHelp,
	"/":   ActionFilter,
	"q":   ActionQuit,
}

// actionToName maps actions to their config-file names.
var actionToName = map[Action]string{
	ActionMoveUp:         "move_up",
	ActionMoveDown:       "move_down",
	ActionMoveLeft:       "move_left",
	ActionMoveRight:      "move_right",
	ActionConfirm:        "confirm",
	ActionQuickSwap:      "quick_swap",
	ActionDirectSwitch:   "direct_switch",
	ActionBack:           "back",
	ActionStartMark:      "start_mark",
	ActionNew:            "new",
	ActionRename:         "rename",
	ActionKill:           "kill",
	ActionCut:            "cut",
	ActionPaste:          "paste",
	ActionTag:            "tag",
	ActionReorderUp:      "reorder_up",
	ActionReorderDown:    "reorder_down",
	ActionReorderLeft:    "reorder_left",
	ActionReorderRight:   "reorder_right",
	ActionBrowseDirs:    "browse_dirs",
	ActionTogglePreview:  "toggle_preview",
	ActionToggleHelp:     "toggle_help",
	ActionFilter:         "filter",
	ActionQuit:           "quit",
}

// nameToAction is the reverse of actionToName.
var nameToAction map[string]Action

func init() {
	nameToAction = make(map[string]Action, len(actionToName))
	for a, n := range actionToName {
		nameToAction[n] = a
	}
}

// ApplyOverrides adds user-specified key bindings on top of the defaults.
// Each entry maps an action name (e.g. "quit") to a key string (e.g. "Q").
// If the new key is already bound to a different action, that conflicting
// binding is removed. Existing bindings for the same action are preserved
// (e.g. arrow keys remain alongside hjkl overrides).
func ApplyOverrides(overrides map[string]string) {
	for actionName, newKey := range overrides {
		action, ok := nameToAction[actionName]
		if !ok {
			continue // unknown action name, skip
		}

		// Remove conflicting binding: if newKey is bound to a *different* action, drop it.
		if existing, bound := defaultKeymap[newKey]; bound && existing != action {
			delete(defaultKeymap, newKey)
		}

		// Add new binding.
		defaultKeymap[newKey] = action
		reservedKeys[newKey] = true
	}
}

// IsReserved returns true if the key is reserved and cannot be used as a mark.
func IsReserved(key string) bool {
	return reservedKeys[key]
}

// Resolve maps a raw key string to an Action using the default keymap.
// Returns ActionNone if the key has no binding.
func Resolve(key string) Action {
	if a, ok := defaultKeymap[key]; ok {
		return a
	}
	return ActionNone
}
