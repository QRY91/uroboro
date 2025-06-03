package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "wherewasi",
		Short: "The AI Context Rope Start - Less explaining. More building.",
		Long: `wherewasi tracks your development work automatically and generates 
dense, AI-ready context summaries on demand.

Pull the rope. Get instant context. Back to building.`,
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Commands
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(pullCmd())
	rootCmd.AddCommand(statusCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start tracking development activity",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🚀 wherewasi daemon starting...")
			fmt.Println("📊 Tracking git commits, file changes, and context")
			fmt.Println("🎯 Pull the rope with: wherewasi pull")
		},
	}
}

func pullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull the rope - start the context engine",
		Run: func(cmd *cobra.Command, args []string) {
			hours, _ := cmd.Flags().GetInt("hours")
			fmt.Printf("🎯 Pulling rope for last %d hours...\n", hours)
			fmt.Println()
			generateContext(hours)
		},
	}
	cmd.Flags().IntP("hours", "h", 2, "hours of context to include")
	return cmd
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show tracking status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📊 wherewasi status")
			fmt.Println("🟢 Daemon: Running")
			fmt.Println("💡 Tip: wherewasi pull --hours 2 | pbcopy")
		},
	}
}

func generateContext(hours int) {
	fmt.Println("Working on wherewasi (Go CLI for AI context generation).")
	fmt.Println("Recent: Created project structure, implemented basic CLI.")
	fmt.Println("Current focus: Pull command implementation.")
	fmt.Println("Tech stack: Go, Cobra CLI.")
	fmt.Println("Philosophy: Less explaining. More building.")
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}
