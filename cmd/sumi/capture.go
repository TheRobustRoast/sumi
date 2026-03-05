package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/capture"
	"sumi/internal/theme"
	"sumi/internal/waybar"
)

func captureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "capture",
		Aliases: []string{"ss"},
		Short:   "Screenshot and screen recording",
		Long: "Take screenshots (area, fullscreen, window) or record the screen.\nResults go to clipboard by default; use -save variants for files.",
		Example: `  sumi capture area          # select area → clipboard
  sumi capture screen-save   # fullscreen → file
  sumi capture record        # start/stop recording
  sumi capture record --gif  # record as GIF`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Interactive mode picker when no subcommand given
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Capture mode").
						Options(
							huh.NewOption("Area → clipboard", "area"),
							huh.NewOption("Area → file", "area-save"),
							huh.NewOption("Fullscreen → clipboard", "screen"),
							huh.NewOption("Fullscreen → file", "screen-save"),
							huh.NewOption("Window → clipboard", "window"),
							huh.NewOption("Window → file", "window-save"),
							huh.NewOption("Start recording (area)", "record"),
							huh.NewOption("Start recording (fullscreen)", "record-full"),
							huh.NewOption("Stop recording", "stop"),
						).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			opts := capture.CaptureOpts{Notify: true}
			switch choice {
			case "area":
				opts.Clipboard = true
				_, err := capture.CaptureArea(opts)
				return err
			case "area-save":
				path, err := capture.CaptureArea(opts)
				if err == nil && path != "" {
					fmt.Println(theme.Ok("saved: " + path))
				}
				return err
			case "screen":
				opts.Clipboard = true
				_, err := capture.CaptureScreen(opts)
				return err
			case "screen-save":
				path, err := capture.CaptureScreen(opts)
				if err == nil {
					fmt.Println(theme.Ok("saved: " + path))
				}
				return err
			case "window":
				opts.Clipboard = true
				_, err := capture.CaptureWindow(opts)
				return err
			case "window-save":
				path, err := capture.CaptureWindow(opts)
				if err == nil {
					fmt.Println(theme.Ok("saved: " + path))
				}
				return err
			case "record":
				return capture.StartRecording(capture.RecordOpts{Area: true})
			case "record-full":
				return capture.StartRecording(capture.RecordOpts{Area: false})
			case "stop":
				return capture.StopRecording()
			}
			return nil
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "screen",
			Short: "Full screen screenshot → clipboard",
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := capture.CaptureScreen(capture.CaptureOpts{Clipboard: true, Notify: true})
				return err
			},
		},
		&cobra.Command{
			Use:   "screen-save",
			Short: "Full screen screenshot → file",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := capture.CaptureScreen(capture.CaptureOpts{Notify: true})
				if err != nil {
					return err
				}
				fmt.Println(theme.Ok("saved: " + path))
				return nil
			},
		},
		&cobra.Command{
			Use:   "area",
			Short: "Area selection screenshot → clipboard",
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := capture.CaptureArea(capture.CaptureOpts{Clipboard: true, Notify: true})
				return err
			},
		},
		&cobra.Command{
			Use:   "area-save",
			Short: "Area selection screenshot → file",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := capture.CaptureArea(capture.CaptureOpts{Notify: true})
				if err != nil {
					return err
				}
				if path != "" {
					fmt.Println(theme.Ok("saved: " + path))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "window",
			Short: "Active window screenshot → clipboard",
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := capture.CaptureWindow(capture.CaptureOpts{Clipboard: true, Notify: true})
				return err
			},
		},
		&cobra.Command{
			Use:   "window-save",
			Short: "Active window screenshot → file",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := capture.CaptureWindow(capture.CaptureOpts{Notify: true})
				if err != nil {
					return err
				}
				fmt.Println(theme.Ok("saved: " + path))
				return nil
			},
		},
		recordCmd(),
	)

	return cmd
}

func recordCmd() *cobra.Command {
	var areaFlag, gifFlag, stopFlag, statusFlag bool

	cmd := &cobra.Command{
		Use:   "record",
		Short: "Start/stop screen recording",
		Long:  "Toggle screen recording with wf-recorder. Records selected area by default.\nUse --area=false for fullscreen, --gif for animated GIF output.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if statusFlag {
				text, class := capture.RecordingStatus()
				waybar.PrintJSON(waybar.Module{Text: text, Class: class})
				return nil
			}
			if stopFlag || capture.IsRecording() {
				return capture.StopRecording()
			}
			return capture.StartRecording(capture.RecordOpts{
				Area: areaFlag,
				GIF:  gifFlag,
			})
		},
	}
	cmd.Flags().BoolVar(&areaFlag, "area", true, "Select area with slurp (use --area=false for fullscreen)")
	cmd.Flags().BoolVar(&gifFlag, "gif", false, "Record as GIF")
	cmd.Flags().BoolVar(&stopFlag, "stop", false, "Stop active recording")
	cmd.Flags().BoolVar(&statusFlag, "status", false, "Output recording status as waybar JSON")
	return cmd
}
