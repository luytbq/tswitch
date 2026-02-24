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
	ActionConfirm   // enter - drill into / switch
	ActionQuickSwap // space - quick switch
	ActionBack      // esc - go back or quit

	// Marks
	ActionStartMark // m - enter marking mode

	// Management (future)
	ActionNew    // n
	ActionRename // r
	ActionKill   // x
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
	"esc": true, "enter": true, "space": true, "tab": true,
	"up": true, "down": true, "left": true, "right": true,
	"j": true, "k": true, "h": true, "l": true,
	"?": true, "q": true, "m": true, "/": true,
	"f": true,
	"n": true, "r": true, "x": true, "t": true,
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

	"enter": ActionConfirm,
	"space": ActionQuickSwap,
	"esc":   ActionBack,

	"m": ActionStartMark,

	"f": ActionBrowseDirs,
	"n": ActionNew,
	"r": ActionRename,
	"x": ActionKill,
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
	ActionBack:           "back",
	ActionStartMark:      "start_mark",
	ActionNew:            "new",
	ActionRename:         "rename",
	ActionKill:           "kill",
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

// ApplyOverrides replaces default key bindings with user-specified overrides.
// Each entry maps an action name (e.g. "quit") to a key string (e.g. "Q").
// The old key for that action is removed; the new key is registered.
func ApplyOverrides(overrides map[string]string) {
	for actionName, newKey := range overrides {
		action, ok := nameToAction[actionName]
		if !ok {
			continue // unknown action name, skip
		}

		// Remove old key(s) bound to this action.
		for key, a := range defaultKeymap {
			if a == action {
				delete(defaultKeymap, key)
				delete(reservedKeys, key)
			}
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
