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
	"n": true, "r": true, "x": true, "t": true,
}

// defaultKeymap maps key strings to actions.
var defaultKeymap = map[string]Action{
	"up": ActionMoveUp, "k": ActionMoveUp,
	"down": ActionMoveDown, "j": ActionMoveDown,
	"left": ActionMoveLeft, "h": ActionMoveLeft,
	"right": ActionMoveRight, "l": ActionMoveRight,

	"enter": ActionConfirm,
	"space": ActionQuickSwap,
	"esc":   ActionBack,

	"m": ActionStartMark,

	"n": ActionNew,
	"r": ActionRename,
	"x": ActionKill,
	"t": ActionTag,

	"tab": ActionTogglePreview,
	"?":   ActionToggleHelp,
	"/":   ActionFilter,
	"q":   ActionQuit,
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
