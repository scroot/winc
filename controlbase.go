/*
 * Copyright (C) 2019 The Winc Authors. All Rights Reserved.
 * Copyright (C) 2010-2013 Allen Dang. All Rights Reserved.
 */

package winc

import (
	"fmt"
	"syscall"

	"github.com/scroot/winc/w32"
)

type ControlBase struct {
	hwnd        w32.HWND
	font        *Font
	parent      Controller
	contextMenu *MenuItem

	isForm bool

	// General events
	onCreate EventManager
	onClose  EventManager

	// Focus events
	onKillFocus EventManager
	onSetFocus  EventManager

	// Drag and drop events
	onDropFiles EventManager

	// Mouse events
	onLBDown    EventManager
	onLBUp      EventManager
	onLBDbl     EventManager
	onMBDown    EventManager
	onMBUp      EventManager
	onRBDown    EventManager
	onRBUp      EventManager
	onRBDbl     EventManager
	onMouseMove EventManager

	// use MouseControl to capture onMouseHover and onMouseLeave events.
	onMouseHover EventManager
	onMouseLeave EventManager

	// Keyboard events
	onKeyUp EventManager

	// Paint events
	onPaint EventManager
	onSize  EventManager
}

// initControl is called by controls: edit, button, treeview, listview, and so on.
func (cba *ControlBase) InitControl(className string, parent Controller, exstyle, style uint) {
	cba.hwnd = CreateWindow(className, parent, exstyle, style)
	if cba.hwnd == 0 {
		panic("cannot create window for " + className)
	}
	cba.parent = parent
}

// InitWindow is called by custom window based controls such as split, panel, etc.
func (cba *ControlBase) InitWindow(className string, parent Controller, exstyle, style uint) {
	RegClassOnlyOnce(className)
	cba.hwnd = CreateWindow(className, parent, exstyle, style)
	if cba.hwnd == 0 {
		panic("cannot create window for " + className)
	}
	cba.parent = parent
}

// SetTheme for TreeView and ListView controls.
func (cba *ControlBase) SetTheme(appName string) error {
	if hr := w32.SetWindowTheme(cba.hwnd, syscall.StringToUTF16Ptr(appName), nil); w32.FAILED(hr) {
		return fmt.Errorf("SetWindowTheme %d", hr)
	}
	return nil
}

func (cba *ControlBase) Handle() w32.HWND {
	return cba.hwnd
}

func (cba *ControlBase) SetHandle(hwnd w32.HWND) {
	cba.hwnd = hwnd
}

func (cba *ControlBase) SetAndClearStyleBits(set, clear uint32) error {
	style := uint32(w32.GetWindowLong(cba.hwnd, w32.GWL_STYLE))
	if style == 0 {
		return fmt.Errorf("GetWindowLong")
	}

	if newStyle := style&^clear | set; newStyle != style {
		if w32.SetWindowLong(cba.hwnd, w32.GWL_STYLE, newStyle) == 0 {
			return fmt.Errorf("SetWindowLong")
		}
	}
	return nil
}

func (cba *ControlBase) SetIsForm(isform bool) {
	cba.isForm = isform
}

func (cba *ControlBase) SetText(caption string) {
	w32.SetWindowText(cba.hwnd, caption)
}

func (cba *ControlBase) Text() string {
	return w32.GetWindowText(cba.hwnd)
}

func (cba *ControlBase) Close() {
	UnRegMsgHandler(cba.hwnd)
	w32.DestroyWindow(cba.hwnd)
}

func (cba *ControlBase) SetSize(width, height int) {
	x, y := cba.Pos()
	w32.MoveWindow(cba.hwnd, x, y, width, height, true)
}

func (cba *ControlBase) Size() (width, height int) {
	rect := w32.GetWindowRect(cba.hwnd)
	width = int(rect.Right - rect.Left)
	height = int(rect.Bottom - rect.Top)
	return
}

func (cba *ControlBase) Width() int {
	rect := w32.GetWindowRect(cba.hwnd)
	return int(rect.Right - rect.Left)
}

func (cba *ControlBase) Height() int {
	rect := w32.GetWindowRect(cba.hwnd)
	return int(rect.Bottom - rect.Top)
}

func (cba *ControlBase) SetPos(x, y int) {
	w, h := cba.Size()
	if w == 0 {
		w = 100
	}
	if h == 0 {
		h = 25
	}
	w32.MoveWindow(cba.hwnd, x, y, w, h, true)
}

func (cba *ControlBase) Pos() (x, y int) {
	rect := w32.GetWindowRect(cba.hwnd)
	x = int(rect.Left)
	y = int(rect.Top)
	if !cba.isForm && cba.parent != nil {
		x, y, _ = w32.ScreenToClient(cba.parent.Handle(), x, y)
	}
	return
}

func (cba *ControlBase) Visible() bool {
	return w32.IsWindowVisible(cba.hwnd)
}

func (cba *ControlBase) ToggleVisible() bool {
	visible := w32.IsWindowVisible(cba.hwnd)
	if visible {
		cba.Hide()
	} else {
		cba.Show()
	}
	return !visible
}

func (cba *ControlBase) ContextMenu() *MenuItem {
	return cba.contextMenu
}

func (cba *ControlBase) SetContextMenu(menu *MenuItem) {
	cba.contextMenu = menu
}

func (cba *ControlBase) Bounds() *Rect {
	rect := w32.GetWindowRect(cba.hwnd)
	if cba.isForm {
		return &Rect{*rect}
	}

	return ScreenToClientRect(cba.hwnd, rect)
}

func (cba *ControlBase) ClientRect() *Rect {
	rect := w32.GetClientRect(cba.hwnd)
	return ScreenToClientRect(cba.hwnd, rect)
}
func (cba *ControlBase) ClientWidth() int {
	rect := w32.GetClientRect(cba.hwnd)
	return int(rect.Right - rect.Left)
}

func (cba *ControlBase) ClientHeight() int {
	rect := w32.GetClientRect(cba.hwnd)
	return int(rect.Bottom - rect.Top)
}

func (cba *ControlBase) Show() {
	w32.ShowWindow(cba.hwnd, w32.SW_SHOWDEFAULT)
}

func (cba *ControlBase) Hide() {
	w32.ShowWindow(cba.hwnd, w32.SW_HIDE)
}

func (cba *ControlBase) Enabled() bool {
	return w32.IsWindowEnabled(cba.hwnd)
}

func (cba *ControlBase) SetEnabled(b bool) {
	w32.EnableWindow(cba.hwnd, b)
}

func (cba *ControlBase) SetFocus() {
	w32.SetFocus(cba.hwnd)
}

func (cba *ControlBase) Invalidate(erase bool) {
	// pRect := w32.GetClientRect(cba.hwnd)
	// if cba.isForm {
	// 	w32.InvalidateRect(cba.hwnd, pRect, erase)
	// } else {
	// 	rc := ScreenToClientRect(cba.hwnd, pRect)
	// 	w32.InvalidateRect(cba.hwnd, rc.GetW32Rect(), erase)
	// }
	w32.InvalidateRect(cba.hwnd, nil, erase)
}

func (cba *ControlBase) Parent() Controller {
	return cba.parent
}

func (cba *ControlBase) SetParent(parent Controller) {
	cba.parent = parent
}

func (cba *ControlBase) Font() *Font {
	return cba.font
}

func (cba *ControlBase) SetFont(font *Font) {
	w32.SendMessage(cba.hwnd, w32.WM_SETFONT, uintptr(font.hfont), 1)
	cba.font = font
}

func (cba *ControlBase) EnableDragAcceptFiles(b bool) {
	w32.DragAcceptFiles(cba.hwnd, b)
}

func (cba *ControlBase) InvokeRequired() bool {
	if cba.hwnd == 0 {
		return false
	}

	windowThreadId, _ := w32.GetWindowThreadProcessId(cba.hwnd)
	currentThreadId := w32.GetCurrentThread()

	return windowThreadId != currentThreadId
}

func (cba *ControlBase) PreTranslateMessage(msg *w32.MSG) bool {
	if msg.Message == w32.WM_GETDLGCODE {
		println("pretranslate, WM_GETDLGCODE")
	}
	return false
}

//Events
func (cba *ControlBase) OnCreate() *EventManager {
	return &cba.onCreate
}

func (cba *ControlBase) OnClose() *EventManager {
	return &cba.onClose
}

func (cba *ControlBase) OnKillFocus() *EventManager {
	return &cba.onKillFocus
}

func (cba *ControlBase) OnSetFocus() *EventManager {
	return &cba.onSetFocus
}

func (cba *ControlBase) OnDropFiles() *EventManager {
	return &cba.onDropFiles
}

func (cba *ControlBase) OnLBDown() *EventManager {
	return &cba.onLBDown
}

func (cba *ControlBase) OnLBUp() *EventManager {
	return &cba.onLBUp
}

func (cba *ControlBase) OnLBDbl() *EventManager {
	return &cba.onLBDbl
}

func (cba *ControlBase) OnMBDown() *EventManager {
	return &cba.onMBDown
}

func (cba *ControlBase) OnMBUp() *EventManager {
	return &cba.onMBUp
}

func (cba *ControlBase) OnRBDown() *EventManager {
	return &cba.onRBDown
}

func (cba *ControlBase) OnRBUp() *EventManager {
	return &cba.onRBUp
}

func (cba *ControlBase) OnRBDbl() *EventManager {
	return &cba.onRBDbl
}

func (cba *ControlBase) OnMouseMove() *EventManager {
	return &cba.onMouseMove
}

func (cba *ControlBase) OnMouseHover() *EventManager {
	return &cba.onMouseHover
}

func (cba *ControlBase) OnMouseLeave() *EventManager {
	return &cba.onMouseLeave
}

func (cba *ControlBase) OnPaint() *EventManager {
	return &cba.onPaint
}

func (cba *ControlBase) OnSize() *EventManager {
	return &cba.onSize
}

func (cba *ControlBase) OnKeyUp() *EventManager {
	return &cba.onKeyUp
}
