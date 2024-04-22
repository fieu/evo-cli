package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var ProgramName = "evo"

var rootCmd = &cobra.Command{
	Use: ProgramName,
}

type MakefileTarget struct {
	Name        string
	Description string
}

func Execute() {
	makefileDirectory := viper.GetString("makefile_path")
	cmd := exec.Command("make", "-C", makefileDirectory, "list-targets-full")

	cmd.Dir = makefileDirectory
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	lines := strings.Split(string(output), "\n")
	var targets []MakefileTarget

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		description := strings.Join(parts[1:], " ")

		targets = append(targets, MakefileTarget{Name: name, Description: description})
	}

	for _, target := range targets {
		target.Name = strings.TrimSpace(target.Name)
		target.Description = strings.TrimSpace(target.Description)
		makeTargetCmd := &cobra.Command{
			Use:   target.Name,
			Short: target.Description,
			Run: func(cmd *cobra.Command, args []string) {
				makeArgs := append([]string{"-C", makefileDirectory}, append([]string{target.Name}, args...)...)
				makeCommand := exec.Command("make", makeArgs...)
				makeCommand.Dir = "."

				// throw command in a pty cause docker is ass
				ptmx, err := pty.Start(makeCommand)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				// Handle pty size.
				ch := make(chan os.Signal, 1)
				signal.Notify(ch, syscall.SIGWINCH)
				go func() {
					for range ch {
						if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
							log.Printf("error resizing pty: %s", err)
						}
					}
				}()
				ch <- syscall.SIGWINCH                        // Initial resize.
				defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

				// set stdin in raw mode. (for ctrl+d and shit)
				oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
				if err != nil {
					panic(err)
				}
				defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

				// Copy stdin to the pty and the pty to stdout.
				go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
				_, _ = io.Copy(os.Stdout, ptmx)
			},
		}
		rootCmd.AddCommand(makeTargetCmd)
	}
	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
