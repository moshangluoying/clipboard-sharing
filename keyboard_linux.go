//go:build linux
// +build linux

package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
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
	// Using xset to get LED states
	cmd := exec.Command("xset", "q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get keyboard LED states: %v", err)
	}

	outputStr := string(output)
	return &KeyboardState{
		NumLock:  strings.Contains(outputStr, "Num Lock:    on"),
		CapsLock: strings.Contains(outputStr, "Caps Lock:   on"),
	}, nil
}

func (k *KeyboardManager) SetState(state *KeyboardState) error {
	currentState, err := k.GetState()
	if err != nil {
		return err
	}

	// Toggle NumLock if needed
	if currentState.NumLock != state.NumLock {
		cmd := exec.Command("xdotool", "key", "Num_Lock")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to toggle NumLock: %v", err)
		}
	}

	// Toggle CapsLock if needed
	if currentState.CapsLock != state.CapsLock {
		cmd := exec.Command("xdotool", "key", "Caps_Lock")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to toggle CapsLock: %v", err)
		}
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
