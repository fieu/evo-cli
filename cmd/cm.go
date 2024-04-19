package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var cmCmd = &cobra.Command{
	Use:   "cm [--no-verify] <message>",
	Short: "Commits changes to a git repository with an optional message",
	Long: `This command commits changes to the current git repository.
You can specify a commit message directly and use the '--no-verify' flag
to skip the pre-commit and pre-push hooks.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		red := color.New(color.FgHiRed).SprintfFunc()
		yellow := color.New(color.FgHiYellow).SprintfFunc()
		noVerify, _ := cmd.Flags().GetBool("no-verify")

		currentDir, _ := os.Getwd()
		_, err := findGitRoot(currentDir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		result, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
		if err != nil {
			fmt.Println("Error getting current Git branch")
			os.Exit(1)
		}
		currentBranch := strings.TrimSpace(string(result))

		commitMessage := strings.Join(args, " ")

		conventionalCommitList := []string{
			"build", "ci", "docs", "feat", "fix", "perf", "refactor", "style", "test", "revert", "chore", "wip",
		}

		isConventional := "n"
		for _, prefix := range conventionalCommitList {
			if strings.Contains(commitMessage, prefix+": ") {
				isConventional = "y"
				break
			}
		}
		if isConventional == "n" {
			fmt.Print("Is this a conventional commit? (y/n): ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			isConventional = scanner.Text()
			if strings.ToLower(isConventional) == "y" {
				fmt.Printf("Choose a type (%s): ", strings.Join(conventionalCommitList, ", "))
				scanner.Scan()
				chosenType := scanner.Text()
				if !contains(conventionalCommitList, chosenType) {
					fmt.Println("Invalid type. Exiting.")
					os.Exit(1)
				}
				commitMessage = chosenType + ": " + commitMessage
			}
		}

		if strings.Contains(currentBranch, "/") {
			parts := strings.SplitN(currentBranch, "/", 2)
			prefix := parts[1]
			commitMessage = prefix + " - " + commitMessage
		}

		if noVerify {
			fmt.Println(red("Commit will not be verified."))
		}

		fmt.Printf("Commit message is:\n> " + yellow(commitMessage) + "\nConfirm? (y/n): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		confirm := scanner.Text()
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Commit cancelled.")
			os.Exit(0)
		}

		commitArgs := []string{"git", "commit", "-m", commitMessage}
		if noVerify {
			commitArgs = append(commitArgs, "--no-verify")
		}
		cmCommand := exec.Command(commitArgs[0], commitArgs[1:]...)

		// throw command in a pty cause docker is ass
		ptmx, err := pty.Start(cmCommand)
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

func init() {
	rootCmd.AddCommand(cmCmd)
	cmCmd.Flags().Bool("no-verify", false, "Skip pre-commit and pre-push hooks")
}

func findGitRoot(currentDirectory string) (string, error) {
	for currentDirectory != "/" {
		if _, err := os.Stat(filepath.Join(currentDirectory, ".git")); err == nil {
			return currentDirectory, nil
		}
		currentDirectory = filepath.Dir(currentDirectory)
	}
	return "", fmt.Errorf("no Git repository found")
}

func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if a == needle {
			return true
		}
	}
	return false
}
