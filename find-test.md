To modify the existing function so that it accepts a directory path and searches within that specific directory and its subdirectories, you need to adjust the `filepath.Walk` call to start from the provided directory instead of the current directory (`"."`). Here is the adjusted code with an explanation of the changes:

```go
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// findMatchingFile searches recursively in the specified directory and all subdirectories
// for files containing specific patterns based on the provided word.
// Returns the path of the first file with a matching pattern or an empty string if no match is found.
func findMatchingFile(dir, word string) (string, error) {
	// Define the regex pattern based on the word's prefix.
	var pattern string
	if hasTestPrefix(word) {
		pattern = `public function ` + regexp.QuoteMeta(word)
	} else {
		pattern = `class ` + regexp.QuoteMeta(word)
	}

	// Compile the regex to ensure it's valid.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err // return with error if regex compilation fails
	}

	// Declare a variable to store the path of the file that contains the match.
	var matchedFilePath string

	// Use filepath.Walk to iterate over all files in the provided directory tree.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // propagate errors encountered by Walk
		}

		// Skip directories since we're only interested in files.
		if info.IsDir() {
			return nil
		}

		// Read the file content.
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err // continue walking if file can't be read
		}

		// Check if content matches the pattern.
		if re.Match(content) {
			matchedFilePath = path
			return filepath.SkipDir // stop walking the directory tree
		}

		return nil
	})

	if err != nil {
		return "", err // return with error encountered during walking
	}

	return matchedFilePath, nil
}

// hasTestPrefix checks if the given word starts with "test".
func hasTestPrefix(word string) bool {
	return strings.HasPrefix(word, "test")
}

func main() {
	// Example usage:
	dirPath := "/path/to/search/directory" // Set the directory path
	filePath, err := findMatchingFile(dirPath, "testFunction")
	if err != nil {
		println("Error:", err.Error())
	} else if filePath != "" {
		println("Match found in file:", filePath)
	} else {
		println("No match found.")
	}
}
```

### Changes:
- **Function Signature**: The `findMatchingFile` function now takes an additional `dir` parameter which specifies the directory to search in.
- **Starting Directory**: The `filepath.Walk` function is called with `dir` instead of `"."`, so it starts the file tree walk from the directory provided by the user.
- **Usage Example**: The `main` function includes an example of how to call `findMatchingFile` with a specified directory path.

This approach ensures that the search is confined to the specified directory and its subdirectories, allowing for more targeted searches across different locations in the filesystem.