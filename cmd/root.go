package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bannerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFE105")).
	PaddingTop(1).
	PaddingBottom(2).
	PaddingLeft(2).
	Width(22)

var commandHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFE105")).
	Bold(false)

var commandStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#5EA704")).
	MarginLeft(1).
	Bold(false)

var descriptionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#E3E3E3")).
	Bold(false)

var rootCmd = &cobra.Command{
	Use:   "evo",
	Short: bannerStyle.Render("E V O L I Z"),
	Long:  bannerStyle.Render("E V O L I Z"),
}

type MakefileTarget struct {
	Name        string
	Description string
}

func customHelpFunc(cmd *cobra.Command, args []string) {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render
	if cmd.Long != "" {
		fmt.Println(descriptionStyle.Render(cmd.Long))
	} else {
		fmt.Println(descriptionStyle.Render(cmd.Short))
	}

	if len(cmd.Commands()) > 0 {
		fmt.Println(commandHeaderStyle.Render("Commands:"))
		t := table.New()
		for _, c := range cmd.Commands() {
			t.Row(s(c.Use), s(c.Short))
		}
		t.Border(lipgloss.HiddenBorder())
		fmt.Println(t.Render())
	}

	fmt.Println()
	fmt.Println(commandStyle.Render("Use \"" + cmd.Use + " [command] --help\" for more information about a command."))
}

func Execute() {
	rootCmd.SetHelpFunc(customHelpFunc)
	var makefileDirectory = viper.GetString("makefile_path")
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
		description := strings.Join(parts[1:], " ") // Rejoin the rest as the description.

		targets = append(targets, MakefileTarget{Name: name, Description: description})
	}

	for _, target := range targets {
		target.Name = strings.TrimSpace(target.Name)
		target.Description = strings.TrimSpace(target.Description)
		makeTargetCmd := &cobra.Command{
			Use:   commandStyle.Render(target.Name),
			Short: descriptionStyle.Render(target.Description),
			Run: func(cmd *cobra.Command, args []string) {
				exec.Command("make", target.Name).Run()
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
