//go:build linux
// +build linux

package main

import (
	"context"
	"golang.design/x/clipboard"
)

type ClipboardManager struct {}

func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{}
}

func (c *ClipboardManager) Init() error {
	return clipboard.Init()
}

func (c *ClipboardManager) Watch(ctx context.Context) <-chan []byte {
	return clipboard.Watch(ctx, clipboard.FmtText)
}

func (c *ClipboardManager) Read() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func (c *ClipboardManager) Write(text []byte) error {
	clipboard.Write(clipboard.FmtText, text)
	return nil
}
