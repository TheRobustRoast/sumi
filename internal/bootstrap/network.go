package bootstrap

import (
	"sumi/internal/runner"
)

// CheckNetwork tests internet connectivity by pinging archlinux.org.
func CheckNetwork(send func(string)) bool {
	send("Checking internet connectivity...")
	err := runner.RunCmd(send, "ping", "-c", "1", "-W", "5", "archlinux.org")
	return err == nil
}

// BringUpEthernet attempts to bring up all ethernet interfaces and request DHCP.
func BringUpEthernet(send func(string)) error {
	send("Bringing up ethernet interfaces...")
	return runner.RunCmd(send, "bash", "-c", `
		for iface in $(ip -o link show | awk -F": " '{print $2}' | grep -E "^en"); do
			ip link set "$iface" up 2>/dev/null || true
			if ! ip -4 addr show "$iface" 2>/dev/null | grep -q "inet "; then
				dhcpcd "$iface" --timeout 10 2>/dev/null &
			fi
		done
		sleep 5
	`)
}

// RetryDHCP retries DHCP on all ethernet interfaces with a longer timeout.
func RetryDHCP(send func(string)) error {
	send("Retrying DHCP on all ethernet interfaces...")
	return runner.RunCmd(send, "bash", "-c", `
		for iface in $(ip -o link show | awk -F": " '{print $2}' | grep -E "^en"); do
			ip link set "$iface" up 2>/dev/null || true
			dhcpcd "$iface" --timeout 15 2>/dev/null || true
		done
	`)
}

// DetectTimezone attempts to detect the timezone from the IP address.
func DetectTimezone(send func(string)) string {
	var tz string
	_ = runner.RunCmd(func(line string) {
		if tz == "" {
			tz = line
		}
	}, "curl", "-s", "--max-time", "5", "https://ipapi.co/timezone")
	if tz == "" {
		return "UTC"
	}
	return tz
}
