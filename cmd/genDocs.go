package cmd

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var genDocsCmd = &cobra.Command{
	Use:   "gen-docs",
	Short: "Generate markdown documentation for the CLI commands",
	Long:  `Generate markdown documentation for the CLI commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		evoCli := cmd.Root()
		docsDirectory := "docs"

		_ = os.Mkdir(docsDirectory, 0o755)
		_ = os.RemoveAll(docsDirectory + "/*")
		log.Printf("Generating documentation in %s/", docsDirectory)

		err := doc.GenMarkdownTree(evoCli, docsDirectory)
		if err != nil {
			log.Fatal(err)
		}

		files, err := os.ReadDir(docsDirectory)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if filepath.Ext(file.Name()) == ".md" {
				filePath := filepath.Join(docsDirectory, file.Name())
				err := removeLastTwoLines(filePath)
				if err != nil {
					log.Printf("Failed to modify file %s: %v", filePath, err)
				}
			}
		}

		log.Printf("Documentation generated in %s/", docsDirectory)
	},
}

func removeLastTwoLines(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(lines) > 2 {
		lines = lines[:len(lines)-2]
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
}

func init() {
	rootCmd.AddCommand(genDocsCmd)
}
