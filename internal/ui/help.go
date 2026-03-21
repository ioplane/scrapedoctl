package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")) // Light blue

	commandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")) // Light green

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")) // Light gray
)

// SetCustomHelp sets a custom, stylized help function for the given command.
func SetCustomHelp(cmd *cobra.Command) {
	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(Banner))
		fmt.Printf("Scrape.do CLI & MCP Server\n\n")

		// Usage
		fmt.Println(headerStyle.Render("USAGE"))
		fmt.Printf("  %s\n\n", c.UseLine())

		// Commands
		if len(c.Commands()) > 0 {
			fmt.Println(headerStyle.Render("COMMANDS"))
			for _, sub := range c.Commands() {
				if sub.Hidden {
					continue
				}
				fmt.Printf("  %s %s\n",
					commandStyle.Width(15).Render(sub.Name()),
					descStyle.Render(sub.Short))
			}
			fmt.Println()
		}

		// Flags
		if c.HasAvailableFlags() {
			fmt.Println(headerStyle.Render("FLAGS"))
			fmt.Print(c.Flags().FlagUsages())
			fmt.Println()
		}
	})
}
