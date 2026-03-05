package osd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func requireOSD(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s not found — install with: pacman -S %s", name, name)
	}
	return nil
}

// ShowVolume adjusts volume and shows OSD via dunst.
func ShowVolume(action string) error {
	if err := requireOSD("wpctl"); err != nil {
		return err
	}
	switch action {
	case "up":
		exec.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SINK@", "5%+").Run() //nolint:errcheck
	case "down":
		exec.Command("wpctl", "set-volume", "@DEFAULT_AUDIO_SINK@", "5%-").Run() //nolint:errcheck
	case "mute":
		exec.Command("wpctl", "set-mute", "@DEFAULT_AUDIO_SINK@", "toggle").Run() //nolint:errcheck
	}

	vol, muted := getSinkVolume()
	if muted {
		return notify("vol: MUTED", makeBar(vol), 0)
	}
	return notify(fmt.Sprintf("vol: %d%%", vol), makeBar(vol), vol)
}

// ShowBrightness adjusts brightness and shows OSD via dunst.
func ShowBrightness(action string) error {
	if err := requireOSD("brightnessctl"); err != nil {
		return err
	}
	switch action {
	case "up":
		exec.Command("brightnessctl", "set", "5%+").Run() //nolint:errcheck
	case "down":
		exec.Command("brightnessctl", "set", "5%-").Run() //nolint:errcheck
	}

	bri := getBrightness()
	return notify(fmt.Sprintf("lcd: %d%%", bri), makeBar(bri), bri)
}

// ShowMic toggles mic mute and shows OSD via dunst.
func ShowMic() error {
	if err := requireOSD("wpctl"); err != nil {
		return err
	}
	exec.Command("wpctl", "set-mute", "@DEFAULT_AUDIO_SOURCE@", "toggle").Run() //nolint:errcheck

	vol, muted := getSourceVolume()
	if muted {
		return notify("mic: MUTED", makeBar(0), 0)
	}
	return notify(fmt.Sprintf("mic: %d%%", vol), makeBar(vol), vol)
}

func getSinkVolume() (int, bool) {
	out, err := exec.Command("wpctl", "get-volume", "@DEFAULT_AUDIO_SINK@").Output()
	if err != nil {
		return 0, false
	}
	s := strings.TrimSpace(string(out))
	muted := strings.Contains(s, "MUTED")
	parts := strings.Fields(s)
	if len(parts) >= 2 {
		f, _ := strconv.ParseFloat(parts[1], 64)
		return int(f * 100), muted
	}
	return 0, muted
}

func getSourceVolume() (int, bool) {
	out, err := exec.Command("wpctl", "get-volume", "@DEFAULT_AUDIO_SOURCE@").Output()
	if err != nil {
		return 0, false
	}
	s := strings.TrimSpace(string(out))
	muted := strings.Contains(s, "MUTED")
	parts := strings.Fields(s)
	if len(parts) >= 2 {
		f, _ := strconv.ParseFloat(parts[1], 64)
		return int(f * 100), muted
	}
	return 0, muted
}

func getBrightness() int {
	out, err := exec.Command("brightnessctl", "-m").Output()
	if err != nil {
		return 0
	}
	// Format: device,class,cur,max,pct%
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) >= 5 {
		pct := strings.TrimRight(parts[4], "%")
		v, _ := strconv.Atoi(pct)
		return v
	}
	return 0
}

func makeBar(pct int) string {
	const max = 20
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := pct * max / 100
	empty := max - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + fmt.Sprintf("] %d%%", pct)
}

func notify(title, body string, value int) error {
	args := []string{
		"-a", "sumi-osd",
		"-h", "string:x-dunst-stack-tag:osd",
		"-h", fmt.Sprintf("int:value:%d", value),
		"-t", "1500",
		title, body,
	}
	return exec.Command("notify-send", args...).Run()
}
