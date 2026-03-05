package keybinds

import (
	"os"
	"strings"
)

// Keybind represents a single Hyprland keybinding.
type Keybind struct {
	Mod         string
	Key         string
	Dispatcher  string
	Args        string
	Description string
	Section     string
}

// ParseKeybindsConf reads a Hyprland keybinds.conf and extracts bind definitions.
func ParseKeybindsConf(path string) ([]Keybind, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var binds []Keybind
	var currentSection string
	var lastComment string

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		// Track section headers: # ── Section ──
		if strings.HasPrefix(line, "# ── ") {
			rest := strings.TrimPrefix(line, "# ── ")
			if end := strings.Index(rest, " ──"); end > 0 {
				currentSection = rest[:end]
			}
			lastComment = ""
			continue
		}

		// Track inline comments for description
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# ╔") && !strings.HasPrefix(line, "# ╚") && !strings.HasPrefix(line, "# ║") {
			lastComment = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			continue
		}

		// Parse bind/binde/bindm lines
		bindType := ""
		switch {
		case strings.HasPrefix(line, "binde = "):
			bindType = "binde"
			line = strings.TrimPrefix(line, "binde = ")
		case strings.HasPrefix(line, "bindm = "):
			bindType = "bindm"
			line = strings.TrimPrefix(line, "bindm = ")
		case strings.HasPrefix(line, "bind = "):
			bindType = "bind"
			line = strings.TrimPrefix(line, "bind = ")
		default:
			lastComment = ""
			continue
		}

		// Extract inline comment as description
		desc := lastComment
		if idx := strings.Index(line, " # "); idx >= 0 {
			desc = strings.TrimSpace(line[idx+3:])
			line = strings.TrimSpace(line[:idx])
		}

		parts := strings.SplitN(line, ", ", 4)
		if len(parts) < 3 {
			lastComment = ""
			continue
		}

		b := Keybind{
			Mod:         formatMod(parts[0]),
			Key:         parts[1],
			Dispatcher:  parts[2],
			Section:     currentSection,
			Description: desc,
		}
		if len(parts) >= 4 {
			b.Args = parts[3]
		}

		// Generate description from dispatcher if none
		if b.Description == "" {
			b.Description = descFromDispatcher(b, bindType)
		}

		binds = append(binds, b)
		lastComment = ""
	}

	return binds, nil
}

func formatMod(mod string) string {
	mod = strings.TrimSpace(mod)
	mod = strings.ReplaceAll(mod, "$mod", "SUPER")
	return mod
}

func descFromDispatcher(b Keybind, bindType string) string {
	switch b.Dispatcher {
	case "exec":
		// Extract the command name
		cmd := b.Args
		if idx := strings.LastIndex(cmd, "/"); idx >= 0 {
			cmd = cmd[idx+1:]
		}
		if idx := strings.Index(cmd, " "); idx >= 0 {
			cmd = cmd[:idx]
		}
		return cmd
	case "killactive":
		return "kill active window"
	case "togglefloating":
		return "toggle floating"
	case "fullscreen":
		return "fullscreen"
	case "movefocus":
		return "focus " + b.Args
	case "movewindow":
		return "move window " + b.Args
	case "workspace":
		return "workspace " + b.Args
	case "movetoworkspace":
		return "move to workspace " + b.Args
	case "togglespecialworkspace":
		return "toggle " + b.Args + " scratchpad"
	case "resizeactive":
		return "resize"
	case "togglegroup":
		return "toggle group"
	default:
		return b.Dispatcher
	}
}
