package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// Common emoji/chars with descriptions — embedded directly, no external file needed.
var emojiData = []struct {
	char string
	desc string
}{
	{"😀", "grinning face"}, {"😂", "face with tears of joy"}, {"🥲", "smiling face with tear"},
	{"😊", "smiling face"}, {"😎", "sunglasses"}, {"🤔", "thinking face"},
	{"😴", "sleeping face"}, {"🤯", "exploding head"}, {"🥳", "partying face"},
	{"😤", "steam from nose"}, {"👍", "thumbs up"}, {"👎", "thumbs down"},
	{"👋", "waving hand"}, {"🤝", "handshake"}, {"✌️", "victory hand"},
	{"🖖", "vulcan salute"}, {"🫡", "saluting face"}, {"💪", "flexed biceps"},
	{"❤️", "red heart"}, {"🔥", "fire"}, {"⭐", "star"}, {"✨", "sparkles"},
	{"💡", "light bulb"}, {"⚡", "high voltage"}, {"🎵", "musical note"},
	{"🎮", "video game"}, {"💻", "laptop"}, {"🖥️", "desktop computer"},
	{"⌨️", "keyboard"}, {"🔧", "wrench"}, {"🔨", "hammer"}, {"⚙️", "gear"},
	{"📁", "file folder"}, {"📂", "open folder"}, {"📝", "memo"},
	{"📌", "pushpin"}, {"✅", "check mark"}, {"❌", "cross mark"},
	{"⚠️", "warning"}, {"🚀", "rocket"}, {"🏠", "house"},
	{"📡", "satellite"}, {"🔒", "locked"}, {"🔓", "unlocked"},
	{"🌙", "crescent moon"}, {"☀️", "sun"}, {"🌊", "water wave"},
	{"🌲", "evergreen tree"},
	{"→", "right arrow"}, {"←", "left arrow"}, {"↑", "up arrow"}, {"↓", "down arrow"},
	{"↔", "left-right arrow"}, {"⇒", "double right arrow"}, {"⇐", "double left arrow"},
	{"•", "bullet"}, {"…", "ellipsis"}, {"—", "em dash"}, {"–", "en dash"},
	{"©", "copyright"}, {"®", "registered"}, {"™", "trademark"},
	{"°", "degree"}, {"±", "plus-minus"}, {"×", "multiplication"}, {"÷", "division"},
	{"≠", "not equal"}, {"≈", "approximately"}, {"≤", "less or equal"}, {"≥", "greater or equal"},
	{"∞", "infinity"}, {"∑", "summation"}, {"√", "square root"}, {"∫", "integral"},
	{"λ", "lambda"}, {"π", "pi"}, {"θ", "theta"}, {"α", "alpha"},
	{"β", "beta"}, {"γ", "gamma"}, {"δ", "delta"}, {"ε", "epsilon"},
}

func emojiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "emoji",
		Short: "Emoji and special character picker",
		Long:  "Search and type emoji via wtype (Wayland) or copy to clipboard with wl-copy.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var opts []huh.Option[string]
			for _, e := range emojiData {
				opts = append(opts, huh.NewOption(
					fmt.Sprintf("%s  %s", e.char, e.desc), e.char))
			}

			var selected string
			err := huh.NewSelect[string]().
				Title("sumi :: emoji / char picker").
				Options(opts...).
				Value(&selected).
				Run()
			if err != nil {
				return silenceQuit(err)
			}

			// Copy to clipboard
			copy := exec.Command("wl-copy")
			copy.Stdin = strings.NewReader(selected)
			copy.Run() //nolint:errcheck

			// Type into focused window
			exec.Command("wtype", selected).Run() //nolint:errcheck

			exec.Command("notify-send", "-t", "1500", "[ char ]", "copied: "+selected).Run() //nolint:errcheck
			return nil
		},
	}
}
