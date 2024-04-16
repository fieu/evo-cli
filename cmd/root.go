package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use: "evo",
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
				makeCommand := exec.Command("make", append([]string{target.Name}, args...)...)
				makeCommand.Dir = makefileDirectory

				makeCommand.Dir = makefileDirectory

				// throw command in a pty cause docker is ass
				ptmx, err := pty.Start(makeCommand)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				defer ptmx.Close()

				go func() {
					_, err := io.Copy(os.Stdout, ptmx)
					if err != nil {
						fmt.Println("Error:", err)
					}
					_, err = io.Copy(os.Stderr, ptmx)
					if err != nil {
						fmt.Println("Error:", err)
					}
				}()

				_ = makeCommand.Wait()
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
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
