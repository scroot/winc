/*
 * Copyright (C) 2019 The Winc Authors. All Rights Reserved.
 */

package winc

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/scroot/winc/w32"
)

var (
	nextMenuItemID  uint16 = 3
	actionsByID            = make(map[uint16]*MenuItem)
	shortcut2Action        = make(map[Shortcut]*MenuItem)
	menuItems              = make(map[w32.HMENU][]*MenuItem)
)

var NoShortcut = Shortcut{}

// Menu for main window and context menus on controls.
// Most methods used for both main window menu and context menu.
type Menu struct {
	hMenu w32.HMENU
	hwnd  w32.HWND // hwnd might be nil if it is context menu.
}

type MenuItem struct {
	hMenu    w32.HMENU
	hSubMenu w32.HMENU // Non zero if this item is in itself a submenu.

	text     string
	toolTip  string
	image    *Bitmap
	shortcut Shortcut
	enabled  bool

	checkable bool
	checked   bool

	id uint16

	onClick EventManager
}

func NewContextMenu() *MenuItem {
	hMenu := w32.CreatePopupMenu()
	if hMenu == 0 {
		panic("failed CreateMenu")
	}

	item := &MenuItem{
		hMenu:    hMenu,
		hSubMenu: hMenu,
	}
	return item
}

func (m *Menu) Dispose() {
	if m.hMenu != 0 {
		w32.DestroyMenu(m.hMenu)
		m.hMenu = 0
	}
}

func (m *Menu) IsDisposed() bool {
	return m.hMenu == 0
}

func initMenuItemInfoFromAction(mii *w32.MENUITEMINFO, a *MenuItem) {
	mii.CbSize = uint32(unsafe.Sizeof(*mii))
	mii.FMask = w32.MIIM_FTYPE | w32.MIIM_ID | w32.MIIM_STATE | w32.MIIM_STRING
	if a.image != nil {
		mii.FMask |= w32.MIIM_BITMAP
		mii.HbmpItem = a.image.handle
	}
	if a.IsSeparator() {
		mii.FType = w32.MFT_SEPARATOR
	} else {
		mii.FType = w32.MFT_STRING
		var text string
		if s := a.shortcut; s.Key != 0 {
			text = fmt.Sprintf("%s\t%s", a.text, s.String())
			shortcut2Action[a.shortcut] = a
		} else {
			text = a.text
		}
		mii.DwTypeData = syscall.StringToUTF16Ptr(text)
		mii.Cch = uint32(len([]rune(a.text)))
	}
	mii.WID = uint32(a.id)

	if a.Enabled() {
		mii.FState &^= w32.MFS_DISABLED
	} else {
		mii.FState |= w32.MFS_DISABLED
	}

	if a.Checkable() {
		mii.FMask |= w32.MIIM_CHECKMARKS
	}
	if a.Checked() {
		mii.FState |= w32.MFS_CHECKED
	}

	if a.hSubMenu != 0 {
		mii.FMask |= w32.MIIM_SUBMENU
		mii.HSubMenu = a.hSubMenu
	}
}

// Show menu on the main window.
func (m *Menu) Show() {
	if !w32.DrawMenuBar(m.hwnd) {
		panic("DrawMenuBar failed")
	}
}

// AddSubMenu returns item that is used as submenu to perform AddItem(s).
func (m *Menu) AddSubMenu(text string) *MenuItem {
	hSubMenu := w32.CreateMenu()
	if hSubMenu == 0 {
		panic("failed CreateMenu")
	}
	return addMenuItem(m.hMenu, hSubMenu, text, Shortcut{}, nil, false)
}

func (mi *MenuItem) OnClick() *EventManager {
	return &mi.onClick
}

func (mi *MenuItem) AddSeparator() {
	addMenuItem(mi.hSubMenu, 0, "-", Shortcut{}, nil, false)
}

// AddItem adds plain menu item.
func (mi *MenuItem) AddItem(text string, shortcut Shortcut) *MenuItem {
	return addMenuItem(mi.hSubMenu, 0, text, shortcut, nil, false)
}

// AddItemCheckable adds plain menu item that can have a checkmark.
func (mi *MenuItem) AddItemCheckable(text string, shortcut Shortcut) *MenuItem {
	return addMenuItem(mi.hSubMenu, 0, text, shortcut, nil, true)
}

// AddItemWithBitmap adds menu item with shortcut and bitmap.
func (mi *MenuItem) AddItemWithBitmap(text string, shortcut Shortcut, image *Bitmap) *MenuItem {
	return addMenuItem(mi.hSubMenu, 0, text, shortcut, image, false)
}

// AddItem to the menu, set text to "-" for separators.
func addMenuItem(hMenu, hSubMenu w32.HMENU, text string, shortcut Shortcut, image *Bitmap, checkable bool) *MenuItem {
	item := &MenuItem{
		hMenu:     hMenu,
		hSubMenu:  hSubMenu,
		text:      text,
		shortcut:  shortcut,
		image:     image,
		enabled:   true,
		id:        nextMenuItemID,
		checkable: checkable,
		//visible:  true,
	}
	nextMenuItemID++
	actionsByID[item.id] = item
	menuItems[hMenu] = append(menuItems[hMenu], item)

	var mii w32.MENUITEMINFO
	initMenuItemInfoFromAction(&mii, item)

	index := -1
	if !w32.InsertMenuItem(hMenu, uint32(index), true, &mii) {
		panic("InsertMenuItem failed")
	}
	return item
}

func indexInObserver(a *MenuItem) int {
	var idx int
	for _, mi := range menuItems[a.hMenu] {
		if mi == a {
			return idx
		}
		idx++
	}
	return -1
}

func findMenuItemByID(id int) *MenuItem {
	return actionsByID[uint16(id)]
}

func (mi *MenuItem) update() {
	var mii w32.MENUITEMINFO
	initMenuItemInfoFromAction(&mii, mi)

	if !w32.SetMenuItemInfo(mi.hMenu, uint32(indexInObserver(mi)), true, &mii) {
		panic("SetMenuItemInfo failed")
	}
	//mi.menu.MenuItemChange(mi)
}

func (mi *MenuItem) IsSeparator() bool { return mi.text == "-" }
func (mi *MenuItem) SetSeparator()     { mi.text = "-" }

func (mi *MenuItem) Enabled() bool     { return mi.enabled }
func (mi *MenuItem) SetEnabled(b bool) { mi.enabled = b; mi.update() }

func (mi *MenuItem) Checkable() bool     { return mi.checkable }
func (mi *MenuItem) SetCheckable(b bool) { mi.checkable = b; mi.update() }

func (mi *MenuItem) Checked() bool     { return mi.checked }
func (mi *MenuItem) SetChecked(b bool) { mi.checked = b; mi.update() }

func (mi *MenuItem) Text() string     { return mi.text }
func (mi *MenuItem) SetText(s string) { mi.text = s; mi.update() }

func (mi *MenuItem) Image() *Bitmap     { return mi.image }
func (mi *MenuItem) SetImage(b *Bitmap) { mi.image = b; mi.update() }

func (mi *MenuItem) ToolTip() string     { return mi.toolTip }
func (mi *MenuItem) SetToolTip(s string) { mi.toolTip = s; mi.update() }
