package main

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/power"
	"sumi/internal/theme"
	"sumi/internal/waybar"
)

func powerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power",
		Short: "Power management (profiles, battery, menu)",
		Long: "Cycle power profiles, view battery info and charge limits, or open the power menu (lock/suspend/reboot/shutdown).",
		Example: `  sumi power profile              # cycle profiles
  sumi power profile performance  # set specific profile
  sumi power battery              # show battery info
  sumi power menu                 # lock/suspend/reboot/shutdown`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Power management").
						Options(
							huh.NewOption("Cycle power profile", "cycle"),
							huh.NewOption("Battery info", "battery"),
							huh.NewOption("Power menu (lock/suspend/reboot/shutdown)", "menu"),
						).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			switch choice {
			case "cycle":
				p, err := power.CycleProfile()
				if err != nil {
					return err
				}
				fmt.Println(theme.Ok("power profile: " + p))
			case "battery":
				info := power.GetBatteryInfo()
				fmt.Println(theme.Step(fmt.Sprintf("battery: %d%% (%s)", info.Capacity, info.Status)))
				fmt.Println(theme.Step(fmt.Sprintf("power: %.1fW · health: %d%% · limit: %d%%",
					info.PowerWatts, info.HealthPct, info.ChargeLimit)))
			case "menu":
				return power.RunPowerMenu()
			}
			return nil
		},
	}

	var waybarFlag bool

	profileCmd := &cobra.Command{
		Use:       "profile [name]",
		Short:     "Get, set, or cycle power profile",
		ValidArgs: []string{"power-saver", "balanced", "performance"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if waybarFlag {
				p := power.CurrentProfile()
				waybar.PrintJSON(waybar.Module{
					Text:    "pwr:" + power.ProfileIcon(p),
					Tooltip: "Power profile: " + p,
					Class:   p,
				})
				return nil
			}
			if len(args) > 0 {
				if err := power.SetProfile(args[0]); err != nil {
					return err
				}
				fmt.Println(theme.Ok("power profile: " + args[0]))
				return nil
			}
			newProfile, err := power.CycleProfile()
			if err != nil {
				return err
			}
			exec.Command("notify-send", "-t", "2000", "[ power ]", "profile: "+newProfile).Run() //nolint:errcheck
			fmt.Println(theme.Ok("power profile: " + newProfile))
			return nil
		},
	}
	profileCmd.Flags().BoolVar(&waybarFlag, "waybar", false, "Output waybar JSON")

	var batteryWaybar bool
	var batteryLimit int

	batteryCmd := &cobra.Command{
		Use:   "battery",
		Short: "Show battery info or set charge limit",
		RunE: func(cmd *cobra.Command, args []string) error {
			if batteryLimit > 0 {
				if err := power.SetChargeLimit(batteryLimit); err != nil {
					return err
				}
				exec.Command("notify-send", "-t", "2000", "[ battery ]",
					fmt.Sprintf("charge limit: %d%%", batteryLimit)).Run() //nolint:errcheck
				return nil
			}
			info := power.GetBatteryInfo()
			if batteryWaybar {
				text := fmt.Sprintf("bat:%d%%%s %.1fW", info.Capacity, info.StatusSuffix(), info.PowerWatts)
				tooltip := fmt.Sprintf("Status: %s\nPower: %.1fW\nHealth: %d%%\nCharge limit: %d%%",
					info.Status, info.PowerWatts, info.HealthPct, info.ChargeLimit)
				waybar.PrintJSON(waybar.Module{
					Text:       text,
					Tooltip:    tooltip,
					Class:      info.WaybarClass(),
					Percentage: info.Capacity,
				})
				return nil
			}
			fmt.Println(theme.Step(fmt.Sprintf("battery: %d%% (%s)", info.Capacity, info.Status)))
			fmt.Println(theme.Step(fmt.Sprintf("power:   %.1fW", info.PowerWatts)))
			fmt.Println(theme.Step(fmt.Sprintf("health:  %d%%", info.HealthPct)))
			fmt.Println(theme.Step(fmt.Sprintf("limit:   %d%%", info.ChargeLimit)))
			return nil
		},
	}
	batteryCmd.Flags().BoolVar(&batteryWaybar, "waybar", false, "Output waybar JSON")
	batteryCmd.Flags().IntVar(&batteryLimit, "limit", 0, "Set charge limit (60-100)")

	menuCmd := &cobra.Command{
		Use:   "menu",
		Short: "Interactive power menu",
		RunE:  func(cmd *cobra.Command, args []string) error { return power.RunPowerMenu() },
	}

	cmd.AddCommand(profileCmd, batteryCmd, menuCmd)
	return cmd
}
