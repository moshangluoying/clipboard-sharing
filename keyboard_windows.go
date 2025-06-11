//go:build windows
// +build windows

package main

import (
	"context"
	"log"
	"syscall"
	"time"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetKeyState   = user32.NewProc("GetKeyState")
	procKeyboardEvent = user32.NewProc("keybd_event")
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
	// 获取键盘状态
	numLockState, _, _ := procGetKeyState.Call(uintptr(VK_NUMLOCK))
	capsLockState, _, _ := procGetKeyState.Call(uintptr(VK_CAPITAL))

	// GetKeyState返回负值表示键被按下，最低位为1表示toggled state
	return &KeyboardState{
		NumLock:  (numLockState & 1) != 0,  // 使用 != 0 代替 == 1 更准确
		CapsLock: (capsLockState & 1) != 0,
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
		ticker := time.NewTicker(20 * time.Millisecond) // 更高的检查频率
		defer ticker.Stop()

		var err error
		lastState, err = k.GetState() // 初始化lastState
		if err != nil {
			log.Printf("Initial keyboard state error: %v", err)
			return
		}

		log.Printf("Started keyboard state monitoring. Initial state: NumLock=%v, CapsLock=%v", 
			lastState.NumLock, lastState.CapsLock)

		for {
			select {			case <-ctx.Done():
				return
			case <-ticker.C:
				state, err := k.GetState()
				if err != nil {
					log.Printf("Error getting keyboard state: %v", err)
					continue
				}
				
				// 检测状态变化
				if lastState == nil || 
				   lastState.NumLock != state.NumLock || 
				   lastState.CapsLock != state.CapsLock {
					log.Printf("Keyboard state changed: NumLock: %v->%v, CapsLock: %v->%v",
						lastState.NumLock, state.NumLock,
						lastState.CapsLock, state.CapsLock)
					
					ch <- state
					lastState = state
				}
			}
		}
	}()
	return ch
}
