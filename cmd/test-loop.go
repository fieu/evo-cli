package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/creack/pty"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func findMatchingFile(baseDir, word string) (string, error) {
	// Define specific subdirectories to search within.
	testDirs := []string{"testing", "tests"}
	for _, dir := range testDirs {
		// Construct the full directory path to search in.
		fullPath := filepath.Join(baseDir, dir)

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
		if info.Mode().Perm()&0400 == 0 {
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
		return "", fmt.Errorf("multiple instances of 'function %s' exist, please rename your function", word)
	}

	return "", nil
}

var testloopCmd = &cobra.Command{
	Use:     "test-loop [test-class-or-function name]",
	Aliases: []string{"tl"},
	Short:   "Run a test class/function in a loop, re-executing it constantly",
	Long:    `Run a test class/function in a loop, re-executing it constantly`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		makefileDirectory := viper.GetString("makefile_path")
		dirPath := "."
		if cmd.Flags().Lookup("dir").Value.String() != "" {
			dirPath = cmd.Flags().Lookup("dir").Value.String()
		}
		filePath, err := findMatchingFile(dirPath, args[0])
		if err != nil {
			fmt.Println("Error:", err.Error())
			return
		}
		if filePath == "" {
			fmt.Println("No match found for:", args[0])
			return
		}

		fmt.Println("Running test", color.HiGreenString(args[0]), "in a loop", "("+color.GreenString(filepath.Base(filePath))+")")

		for {
			makeCommand := exec.Command("make", "-C", makefileDirectory, "test-file", "FILTER="+filePath)
			makeCommand.Dir = dirPath
			ptmx, err := pty.Start(makeCommand)
			if err != nil {
				fmt.Println("Error starting command:", err)
				return
			}

			go func() {
				io.Copy(os.Stdout, ptmx)
				io.Copy(os.Stderr, ptmx)
			}()

			err = makeCommand.Wait()
			ptmx.Close()
			if err != nil {
				fmt.Println("Error during command execution:", err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(testloopCmd)

	testloopCmd.Flags().StringP("dir", "d", "", "Directory to search for test files")
}
