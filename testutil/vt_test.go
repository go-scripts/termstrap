package testutil

import (
	"testing"
)

func TestVirtualScreen_TextAndChar(t *testing.T) {
	content := "Hello World\nLine Two"
	screen := NewScreen(20, 5, content)

	screen.AssertChar(t, 0, 0, 'H')
	screen.AssertChar(t, 10, 0, 'd')
	screen.AssertChar(t, 0, 1, 'L')
	screen.AssertChar(t, 7, 1, 'o')

	screen.AssertText(t, 0, 0, "Hello World")
	screen.AssertText(t, 6, 0, "World")
	screen.AssertText(t, 0, 1, "Line Two")
}

func TestVirtualScreen_BgColor(t *testing.T) {
	// TokyoNight primary: 121;162;247 (hex 79a2f7)
	content := "\x1b[48;2;121;162;247mColored\x1b[0m Plain"
	screen := NewScreen(20, 3, content)

	screen.AssertBgColor(t, 0, 0, "121;162;247")
	screen.AssertBgColor(t, 0, 0, "#79a2f7")
	screen.AssertBgColor(t, 0, 0, "79a2f7")
	screen.AssertBgColor(t, 8, 0, "default")
}

func TestVirtualScreen_BorderBox(t *testing.T) {
	roundedBox := "╭────╮\n│Hi  │\n╰────╯"
	screen1 := NewScreen(10, 5, roundedBox)
	screen1.AssertBorderBox(t, 0, 0, 6, 3)

	normalBox := "┌────┐\n│Hi  │\n└────┘"
	screen2 := NewScreen(10, 5, normalBox)
	screen2.AssertBorderBox(t, 0, 0, 6, 3)
}
