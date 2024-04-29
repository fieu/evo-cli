package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"evo-cli/cmd"

	_ "embed"

	"evo-cli/internal"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

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
		bannerStyle.Render("E V O L I Z") + "  –  " + internal.GetQuote() + "\n"
}

var red = color.New(color.FgHiRed).SprintfFunc()

func main() {
	// Setup global signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println(red("\nReceived CTRL+C, shutting down..."))
		os.Exit(0)
	}()
	// Setup config
	viper.SetConfigName("evo-cli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file not found in $HOME/.config/evo-cli.yml")
			os.Exit(1)
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
