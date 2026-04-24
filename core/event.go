package core

// EventType identifies the kind of event.
type EventType int

const (
	EventMotion       EventType = iota
	EventClick
	EventLongClick
	EventScroll
	EventKeyDown
	EventKeyUp
	EventFocusChanged
	EventMenuCommand
	EventShortcut
)

// Event is the base interface for all events.
type Event interface {
	Type() EventType
	IsConsumed() bool
	Consume()
}

// baseEvent provides common event fields.
type baseEvent struct {
	eventType EventType
	consumed  bool
}

func (e *baseEvent) Type() EventType  { return e.eventType }
func (e *baseEvent) IsConsumed() bool { return e.consumed }
func (e *baseEvent) Consume()         { e.consumed = true }

// PointerSource distinguishes input device.
type PointerSource int

const (
	PointerMouse PointerSource = iota
	PointerTouch
	PointerPen
)

// MotionAction describes what the pointer did.
type MotionAction int

const (
	ActionDown       MotionAction = iota
	ActionMove
	ActionUp
	ActionCancel
	ActionHoverEnter
	ActionHoverMove
	ActionHoverExit
)

// MouseButton identifies a mouse button.
type MouseButton int

const (
	MouseButtonLeft   MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// KeyModifier represents modifier keys held during an event.
type KeyModifier int

const (
	ModNone  KeyModifier = 0
	ModCtrl  KeyModifier = 1
	ModShift KeyModifier = 2
	ModAlt   KeyModifier = 4
)

// Pointer represents a single pointer (finger/pen) in a multi-touch event.
type Pointer struct {
	ID       int
	X, Y     float64
	Pressure float32
}

// MotionEvent represents pointer (mouse/touch) events.
type MotionEvent struct {
	baseEvent
	Action   MotionAction
	Source   PointerSource
	X, Y     float64 // relative to target node
	RawX     float64
	RawY     float64
	Pointers []Pointer
	Button   MouseButton
	Modifier KeyModifier
	Pressure float32
}

// NewMotionEvent creates a MotionEvent with the given action and coordinates.
func NewMotionEvent(action MotionAction, x, y float64) *MotionEvent {
	return &MotionEvent{
		baseEvent: baseEvent{eventType: EventMotion},
		Action:    action,
		X:         x,
		Y:         y,
		Source:    PointerMouse,
		Pressure:  1.0,
	}
}

// Key represents platform-specific key codes.
type Key int

// KeyAction describes the kind of keyboard action.
type KeyAction int

const (
	ActionKeyDown KeyAction = iota // 键按下
	ActionKeyUp                    // 键抬起
)

// KeyEvent represents keyboard events.
type KeyEvent struct {
	baseEvent
	Action   KeyAction
	KeyCode  Key
	Char     rune
	Modifier KeyModifier
}

// NewKeyEvent creates a KeyEvent with the given action and key code.
func NewKeyEvent(action KeyAction, keyCode int) *KeyEvent {
	et := EventKeyDown
	if action == ActionKeyUp {
		et = EventKeyUp
	}
	return &KeyEvent{
		baseEvent: baseEvent{eventType: et},
		Action:    action,
		KeyCode:   Key(keyCode),
	}
}

// ScrollEvent represents mouse wheel / trackpad scroll events.
type ScrollEvent struct {
	baseEvent
	X, Y     float64 // pointer position relative to target node
	DeltaX   float64 // horizontal scroll amount (positive = right)
	DeltaY   float64 // vertical scroll amount (positive = down)
	Modifier KeyModifier
}

// NewScrollEvent creates a ScrollEvent with the given position and scroll deltas.
func NewScrollEvent(x, y, deltaX, deltaY float64) *ScrollEvent {
	return &ScrollEvent{
		baseEvent: baseEvent{eventType: EventScroll},
		X:         x,
		Y:         y,
		DeltaX:    deltaX,
		DeltaY:    deltaY,
	}
}

// FocusEvent represents focus gain/loss on a node.
type FocusEvent struct {
	baseEvent
	Focused bool
}
