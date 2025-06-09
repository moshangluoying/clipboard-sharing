//go:build linux
// +build linux

package main

func enableProcessPrivileges() error {
	// Linux doesn't need special privilege handling for this case
	return nil
}
