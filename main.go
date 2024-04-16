package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"evo-cli/cmd"

	_ "embed"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

//go:embed quotes.txt
var quotesFile string

var bananaArt = `
             ████              
            ██                 
         █████                 
       ██   ███                
     ██     ███                
    ██      █ █                
   ██       █ █                
  ██       ██ █                
  █        █  ██               
 ██        ██  █               
 ██         █  ███             
 ██         ██   ██            
  ██         █    ███          
  ██          ██    ████       
   ██           ██     ███     
    ██            █       ██   
      ██           ██      ██  
       ███          ██       ██
         ████        █████████ 
             ██████████        
                               `

func displayArt() string {
	bananaStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FDEB60"))

	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFE105")).
		MarginLeft(2)

	return bananaStyle.Render(bananaArt) + "\n\n" +
		bannerStyle.Render("E V O L I Z") + "  –  " + displayQuote() + "\n"
}

func displayQuote() string {
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

func main() {
	// Setup config
	viper.SetConfigName("evo-cli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("config file not found in $HOME/.config/evo-cli.yml")
		} else {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	// Display art
	if len(os.Args) == 1 {
		fmt.Println(displayArt())
	}

	// Execute CLI
	cmd.Execute()
}
