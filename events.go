package nksdl

import (
	"fmt"
	"strings"

	"github.com/kbolino/go-nk"
	"github.com/veandco/go-sdl2/sdl"
)

var DefaultBindings = map[KeyInput]KeyAction{
	{Code: sdl.K_LSHIFT}:                                 {Key1: nk.KeyShift},
	{Code: sdl.K_RSHIFT}:                                 {Key1: nk.KeyShift},
	{Code: sdl.K_RETURN}:                                 {Key1: nk.KeyEnter},
	{Code: sdl.K_TAB}:                                    {Key1: nk.KeyTab},
	{Code: sdl.K_BACKSPACE}:                              {Key1: nk.KeyBackspace},
	{Code: sdl.K_HOME}:                                   {Key1: nk.KeyTextStart, Key2: nk.KeyScrollStart},
	{Code: sdl.K_END}:                                    {Key1: nk.KeyTextEnd, Key2: nk.KeyScrollEnd},
	{Code: sdl.K_PAGEUP}:                                 {Key1: nk.KeyScrollUp},
	{Code: sdl.K_PAGEDOWN}:                               {Key2: nk.KeyScrollDown},
	{Code: sdl.K_UP}:                                     {Key1: nk.KeyUp},
	{Code: sdl.K_DOWN}:                                   {Key1: nk.KeyDown},
	{Code: sdl.K_LEFT}:                                   {Key1: nk.KeyLeft},
	{Code: sdl.K_RIGHT}:                                  {Key1: nk.KeyRight},
	{Code: sdl.K_c, Mod: sdl.KMOD_CTRL}:                  {Key1: nk.KeyCopy},
	{Code: sdl.K_x, Mod: sdl.KMOD_CTRL}:                  {Key1: nk.KeyCut},
	{Code: sdl.K_v, Mod: sdl.KMOD_CTRL}:                  {Key1: nk.KeyPaste},
	{Code: sdl.K_a, Mod: sdl.KMOD_CTRL}:                  {Key1: nk.KeyTextLineStart},
	{Code: sdl.K_e, Mod: sdl.KMOD_CTRL}:                  {Key1: nk.KeyTextLineEnd},
	{Code: sdl.K_z, Mod: sdl.KMOD_CTRL}:                  {Key1: nk.KeyTextUndo},
	{Code: sdl.K_z, Mod: sdl.KMOD_CTRL | sdl.KMOD_SHIFT}: {Key1: nk.KeyTextRedo},
}

// EventType represents the type of a handled event. This is not an exhaustive
// list of event types, and only represents those that are recognized by
// EventHandler, with a placeholder (EventTypeUnhandled) for all other types.
type EventType int32

const (
	EventTypeUnhandled EventType = iota
	EventTypeQuit
	EventTypeInputMotion
	EventTypeInputButton
	EventTypeInputScroll
	EventTypeInputKey
	EventTypeInputUnicode
)

// KeyInput is the reduced form of sdl.Keysym containing only the keycode and
// modifiers, used to match input events.
type KeyInput struct {
	Code sdl.Keycode
	Mod  sdl.Keymod
}

// ToString returns the string representation of ki, with the given name for
// the "GUI" key (aka "Meta", "Win", "Super", "Cmd").
func (ki KeyInput) ToString(guiKeyName string) string {
	if ki.Code == sdl.K_UNKNOWN {
		return ""
	}
	var buf strings.Builder
	if !appendMod(&buf, ki.Mod, sdl.KMOD_CTRL, "Ctrl") {
		appendMod(&buf, ki.Mod, sdl.KMOD_LCTRL, "LCtrl")
		appendMod(&buf, ki.Mod, sdl.KMOD_RCTRL, "RCtrl")
	}
	if !appendMod(&buf, ki.Mod, sdl.KMOD_SHIFT, "Shift") {
		appendMod(&buf, ki.Mod, sdl.KMOD_LSHIFT, "LShift")
		appendMod(&buf, ki.Mod, sdl.KMOD_RSHIFT, "RShift")
	}
	if !appendMod(&buf, ki.Mod, sdl.KMOD_ALT, "Alt") {
		appendMod(&buf, ki.Mod, sdl.KMOD_LALT, "LAlt")
		appendMod(&buf, ki.Mod, sdl.KMOD_RALT, "RAlt")
	}
	if !appendMod(&buf, ki.Mod, sdl.KMOD_GUI, guiKeyName) {
		appendMod(&buf, ki.Mod, sdl.KMOD_LGUI, "L"+guiKeyName)
		appendMod(&buf, ki.Mod, sdl.KMOD_RGUI, "R"+guiKeyName)
	}
	if buf.Len() != 0 {
		buf.WriteRune('+')
	}
	buf.WriteString(sdl.GetKeyName(ki.Code))
	return buf.String()
}

// String calls ToString("GUI").
func (ki KeyInput) String() string {
	return ki.ToString("GUI")
}

func appendMod(dst *strings.Builder, mod, mask sdl.Keymod, name string) bool {
	if mod&mask != mask {
		return false
	}
	if dst.Len() != 0 {
		dst.WriteRune('+')
	}
	dst.WriteString(name)
	return true
}

// KeysymInput converts sym to KeyInput.
func KeysymInput(sym sdl.Keysym) KeyInput {
	mod := sym.Mod
	return KeyInput{
		Code: sym.Sym,
		Mod:  sdl.Keymod(mod),
	}
}

// KeyAction represents the action(s) to be taken when key input is received.
// Key2 is optional; if only a single action is needed, leave Key2 unset or set
// it to KeyNone.
type KeyAction struct {
	Key1 nk.Key
	Key2 nk.Key
}

// EventHandler is used to handle input events from SDL and report them to an
// nk.Context.
type EventHandler struct {
	bindings map[KeyInput]KeyAction
}

// NewEventHandler creates a new EventHandler from the given bindings. The map
// will be copied, and any generic modifier key bindings will be expanded into
// left and right key bindings. For example, if there is a binding with Ctrl
// modifier set it will be expanded into separate bindings to the same action
// for both LCtrl and RCtrl instead. NewEventHandler panics if the resulting
// bindings conflict.
func NewEventHandler(bindings map[KeyInput]KeyAction) EventHandler {
	bindingsCopy := make(map[KeyInput]KeyAction, 2*len(bindings))
	expandModBindings(bindingsCopy, bindings)
	return EventHandler{bindingsCopy}
}

// HandleEvent handles the given event, reporting its actions to nkc, using
// the defined bindings. The return value indicates the type of the event and
// whether the event was used at all.
func (h EventHandler) HandleEvent(nkc *nk.Context, event sdl.Event) (EventType, bool) {
	switch e := event.(type) {
	case *sdl.QuitEvent:
		return EventTypeQuit, false
	case *sdl.MouseMotionEvent:
		x, y := e.X, e.Y
		nkc.InputMotion(x, y)
		return EventTypeInputMotion, true
	case *sdl.MouseButtonEvent:
		x, y := e.X, e.Y
		down := false
		if e.State == sdl.PRESSED {
			down = true
		}
		switch e.Button {
		case sdl.BUTTON_LEFT:
			if e.Clicks == 2 {
				nkc.InputButton(nk.ButtonDouble, x, y, down)
			}
			nkc.InputButton(nk.ButtonLeft, x, y, down)
		case sdl.BUTTON_RIGHT:
			nkc.InputButton(nk.ButtonRight, x, y, down)
		case sdl.BUTTON_MIDDLE:
			nkc.InputButton(nk.ButtonMiddle, x, y, down)
		}
		return EventTypeInputButton, true
	case *sdl.MouseWheelEvent:
		nkc.InputScroll(e.PreciseX, e.PreciseY)
		return EventTypeInputScroll, true
	case *sdl.KeyboardEvent:
		var down bool
		if e.State == sdl.PRESSED {
			down = true
		}
		input := KeysymInput(e.Keysym)
		action := h.bindings[input]
		if action.Key1 != nk.KeyNone {
			nkc.InputKey(action.Key1, down)
			if action.Key2 != nk.KeyNone {
				nkc.InputKey(action.Key2, down)
			}
			return EventTypeInputKey, true
		}
		return EventTypeInputKey, false
	case *sdl.TextInputEvent:
		for _, r := range e.GetText() {
			nkc.InputUnicode(r)
		}
		return EventTypeInputUnicode, true
	default:
		return EventTypeUnhandled, false
	}
}

type keyBinding struct {
	input  KeyInput
	action KeyAction
}

// expandModBindings will expand input bindings that use modifier aliases like
// KMOD_CTRL into bindings for specific modifier variants like KMOD_LCTRL and
// KMOD_RCTRL. The new bindings will be read from src and written to dst, which
// may be different maps or the same map. Bindings that don't need to be
// expanded will be written directly to dst.
func expandModBindings(dst, src map[KeyInput]KeyAction) {
	var bindings []keyBinding
	for input, action := range src {
		bindings = bindings[:0]
		bindings = append(bindings, keyBinding{input, action})
		bindings = expandModBinding(bindings, sdl.KMOD_CTRL, sdl.KMOD_LCTRL, sdl.KMOD_RCTRL)
		bindings = expandModBinding(bindings, sdl.KMOD_SHIFT, sdl.KMOD_LSHIFT, sdl.KMOD_RSHIFT)
		bindings = expandModBinding(bindings, sdl.KMOD_ALT, sdl.KMOD_LALT, sdl.KMOD_RALT)
		bindings = expandModBinding(bindings, sdl.KMOD_GUI, sdl.KMOD_LGUI, sdl.KMOD_RGUI)
		for _, binding := range bindings {
			if dstBinding, exists := dst[binding.input]; exists && dstBinding != src[binding.input] {
				panic(fmt.Errorf("conflicting binding: input %s is bound to two different actions", input))
			}
			dst[binding.input] = binding.action
		}
	}
}

// expandModBinding implements expandModBindings for a single modifier key,
// expanding any binding for both into separate bindings for left and right,
// replacing the original binding in, and appending the additional one to, the
// supplied slice.
func expandModBinding(bindings []keyBinding, both, left, right sdl.Keymod) []keyBinding {
	originalLen := len(bindings)
	for i := 0; i < originalLen; i++ {
		input := bindings[i].input
		action := bindings[i].action
		if input.Mod&both != both {
			continue
		}
		leftMod := input.Mod&^both | left
		leftInput := input
		leftInput.Mod = leftMod
		bindings[i] = keyBinding{leftInput, action}
		rightMod := input.Mod&^both | right
		rightInput := input
		rightInput.Mod = rightMod
		bindings = append(bindings, keyBinding{rightInput, action})
	}
	return bindings
}
