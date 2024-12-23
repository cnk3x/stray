package main

import (
	"context"

	"github.com/getlantern/systray"
)

func AddMenuItem(ctx context.Context, label string, options ...MenuAddOption) (menu *systray.MenuItem) {
	var mOpts MenuOptions

	for _, opt := range options {
		opt(&mOpts)
	}

	if mOpts.Description == "" {
		mOpts.Description = label
	}

	if mOpts.Parent != nil {
		if mOpts.Checkable {
			menu = mOpts.Parent.AddSubMenuItemCheckbox(label, mOpts.Description, mOpts.Checked)
		} else {
			menu = mOpts.Parent.AddSubMenuItem(label, mOpts.Description)
		}
	} else {
		if mOpts.Checkable {
			menu = systray.AddMenuItemCheckbox(label, mOpts.Description, mOpts.Checked)
		} else {
			menu = systray.AddMenuItem(label, mOpts.Description)
		}
	}

	if mOpts.Action != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-menu.ClickedCh:
					go mOpts.Action(menu)
				}
			}
		}()
	}

	return menu
}

type MenuOptions struct {
	Description string
	Action      func(menu *systray.MenuItem)
	Parent      *systray.MenuItem
	Checkable   bool
	Checked     bool
}

type MenuAddOption func(options *MenuOptions)

func Action(action func(menu *systray.MenuItem)) MenuAddOption {
	return func(options *MenuOptions) { options.Action = action }
}

func Description(description string) MenuAddOption {
	return func(options *MenuOptions) { options.Description = description }
}

func Parent(parent *systray.MenuItem) MenuAddOption {
	return func(options *MenuOptions) { options.Parent = parent }
}

func Checkable(checked bool) MenuAddOption {
	return func(options *MenuOptions) { options.Checkable, options.Checked = true, checked }
}
