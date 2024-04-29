package internal

import (
	_ "embed"
	"fmt"
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//go:embed quotes.txt
var quotesFile string

func GetQuote() string {
	quoteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#707070")).
		Italic(true)

	quotes := strings.Split(quotesFile, "\n")
	chosen := quotes[rand.Intn(len(quotes))]
	parts := strings.SplitN(chosen, " - ", 2)

	if len(parts) == 2 {
		return quoteStyle.Render(fmt.Sprintf("“%s” – %s", parts[0], parts[1]))
	}
	return quoteStyle.Render(chosen)
}
