package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"github.com/fatih/color"
	hook "github.com/robotn/gohook"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

func findMatchingFile(baseDir, word string) (string, error) {
	// Define specific subdirectories to search within.
	testDirs := []string{"testing", "tests"}
	for _, dir := range testDirs {
		// Construct the full directory path to search in.
		fullPath := filepath.Join(baseDir, dir)

		// If directory doesn't exist, skip it.
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		// Perform the search in the specified directory.
		filePath, err := searchDirectory(fullPath, word)
		if err != nil {
			return "", err // If an error occurred, return it immediately.
		}
		if filePath != "" {
			return filePath, nil // Return the first found file path.
		}
	}
	return "", nil // Return empty if no file was found after searching all specified directories.
}

func searchDirectory(dir, word string) (string, error) {
	// Define the regex pattern based on the word's prefix.
	var pattern string
	if strings.HasPrefix(word, "test") {
		pattern = `public function ` + regexp.QuoteMeta(word)
	} else {
		pattern = `class ` + regexp.QuoteMeta(word)
	}

	// Compile the regex to ensure it's valid.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err // return with error if regex compilation fails
	}

	// Slice to store paths of files that contain matches.
	var matchedFiles []string

	// Use filepath.Walk to iterate over all files in the provided directory tree.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // propagate errors encountered by Walk
		}

		// Check if the file is a symbolic link
		if info.Mode()&os.ModeSymlink != 0 {
			return nil // skip processing symbolic links
		}

		// If have no permission to read the file, skip it.
		if info.Mode().Perm()&0o400 == 0 {
			return nil
		}

		// Only process files.
		if !info.IsDir() {
			// Read the file content.
			content, err := os.ReadFile(path)
			if err != nil {
				return err // continue walking if file can't be read
			}

			// Check if content matches the pattern.
			if re.Match(content) {
				matchedFiles = append(matchedFiles, path)
			}
		}

		return nil
	})
	if err != nil {
		return "", err // return with error encountered during walking
	}

	// Check the number of matched files and decide on the return value.
	if len(matchedFiles) == 1 {
		return matchedFiles[0], nil
	} else if len(matchedFiles) > 1 {
		return "", fmt.Errorf("multiple instances of 'public function %s' exist, please rename your test method", word)
	}

	return "", nil
}

var testloopCmd = &cobra.Command{
	Use:     "test-loop [test-class-or-function name]",
	Aliases: []string{"tl"},
	Short:   "Run a test class/function on demand, restart it with Ctrl+0",
	Long:    `Run a test class/function on demand, restart it with Ctrl+0`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := viper.GetString("dir")
		if dirPath == "" {
			dirPath = "."
		}
		absoluteDirPath, err := filepath.Abs(dirPath)
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		fmt.Println(color.HiYellowString("Searching for test"), color.HiBlueString(args[0]), color.HiYellowString("in directory"), color.HiBlueString(absoluteDirPath)+color.HiYellowString("..."))

		filePath, err := findMatchingFile(dirPath, args[0])
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		if filePath == "" {
			fmt.Println("No match found for:", args[0])
			return
		}
		testRelativePath, _ := filepath.Rel(dirPath, filePath)
		testFilter := strings.TrimSpace(args[0])

		fmt.Println("Running test:", color.HiGreenString(args[0]), "("+color.GreenString(testRelativePath)+")")
		runTest(dirPath, filePath, testFilter)

		fmt.Println("Press CTRL+0 to restart the test", color.HiGreenString(args[0]), "("+color.GreenString(testRelativePath)+")")

		// Register the hotkey
		hook.Register(hook.KeyHold, []string{"ctrl", "0"}, func(e hook.Event) {
			fmt.Println("Restarting test", color.HiGreenString(args[0]), "("+color.GreenString(testRelativePath)+")")
			runTest(dirPath, filePath, testFilter)
		})

		s := hook.Start()
		defer hook.End()
		<-hook.Process(s)
	},
}

func runTest(dirPath, filePath, testFilter string) {
	makefileDirectory := viper.GetString("makefile_path")
	makeCommand := exec.Command("make", "-C", makefileDirectory, "test-file", "FILTER="+testFilter)
	makeCommand.Dir = dirPath

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
}

func init() {
	rootCmd.AddCommand(testloopCmd)

	testloopCmd.Flags().StringP("dir", "d", "", "Directory to search for test files")
}
