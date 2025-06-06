//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"syscall"
)

var (
	user32                = syscall.NewLazyDLL("user32.dll")
	procGetKeyboardState = user32.NewProc("GetKeyboardState")
	procKeyboardEvent    = user32.NewProc("keybd_event")
)

const (
	VK_NUMLOCK  = 0x90
	VK_CAPITAL  = 0x14
	KEYEVENTF_EXTENDEDKEY = 0x1
	KEYEVENTF_KEYUP       = 0x2
)

type KeyboardState struct {
	NumLock  bool
	CapsLock bool
}

type KeyboardManager struct{}

func NewKeyboardManager() *KeyboardManager {
	return &KeyboardManager{}
}

func (k *KeyboardManager) GetState() (*KeyboardState, error) {
	var keyState [256]byte
	ret, _, _ := procGetKeyboardState.Call(uintptr(unsafe.Pointer(&keyState[0])))
	if ret == 0 {
		return nil, fmt.Errorf("get keyboard state failed")
	}

	return &KeyboardState{
		NumLock:  keyState[VK_NUMLOCK]&1 == 1,
		CapsLock: keyState[VK_CAPITAL]&1 == 1,
	}, nil
}

func (k *KeyboardManager) SetState(state *KeyboardState) error {
	currentState, err := k.GetState()
	if err != nil {
		return err
	}

	// Toggle NumLock if needed
	if currentState.NumLock != state.NumLock {
		procKeyboardEvent.Call(
			uintptr(VK_NUMLOCK),
			uintptr(0),
			uintptr(KEYEVENTF_EXTENDEDKEY),
			0,
		)
		procKeyboardEvent.Call(
			uintptr(VK_NUMLOCK),
			uintptr(0),
			uintptr(KEYEVENTF_EXTENDEDKEY|KEYEVENTF_KEYUP),
			0,
		)
	}

	// Toggle CapsLock if needed
	if currentState.CapsLock != state.CapsLock {
		procKeyboardEvent.Call(
			uintptr(VK_CAPITAL),
			uintptr(0),
			uintptr(KEYEVENTF_EXTENDEDKEY),
			0,
		)
		procKeyboardEvent.Call(
			uintptr(VK_CAPITAL),
			uintptr(0),
			uintptr(KEYEVENTF_EXTENDEDKEY|KEYEVENTF_KEYUP),
			0,
		)
	}

	return nil
}

func (k *KeyboardManager) Watch(ctx context.Context) <-chan *KeyboardState {
	ch := make(chan *KeyboardState)
	go func() {
		defer close(ch)
		var lastState *KeyboardState
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				state, err := k.GetState()
				if err != nil {
					continue
				}
				if lastState == nil || 
				   lastState.NumLock != state.NumLock || 
				   lastState.CapsLock != state.CapsLock {
					ch <- state
					lastState = state
				}
			}
		}
	}()
	return ch
}
