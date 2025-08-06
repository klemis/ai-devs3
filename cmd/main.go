package main

import (
	"fmt"
	"os"

	"ai-devs3/internal/config"
	s01e01 "ai-devs3/internal/tasks/s01/e01"
	s01e02 "ai-devs3/internal/tasks/s01/e02"
	s01e03 "ai-devs3/internal/tasks/s01/e03"
	s01e05 "ai-devs3/internal/tasks/s01/e05"
	s02e01 "ai-devs3/internal/tasks/s02/e01"
	s02e02 "ai-devs3/internal/tasks/s02/e02"
	s02e03 "ai-devs3/internal/tasks/s02/e03"
	s02e04 "ai-devs3/internal/tasks/s02/e04"
	s02e05 "ai-devs3/internal/tasks/s02/e05"
	s03e01 "ai-devs3/internal/tasks/s03/e01"
	s03e02 "ai-devs3/internal/tasks/s03/e02"
	s03e03 "ai-devs3/internal/tasks/s03/e03"
	s03e04 "ai-devs3/internal/tasks/s03/e04"
	s03e05 "ai-devs3/internal/tasks/s03/e05"
	"ai-devs3/internal/tasks/utils/ocr"
	"ai-devs3/internal/tasks/utils/video"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-devs3",
	Short: "AI-DEVS3 task runner",
	Long: `AI-DEVS3 CLI tool for running various AI development tasks and challenges.

This tool provides a unified interface for executing different seasons and episodes
of AI-DEVS challenges, including:
- Robot authentication and communication
- Text processing and analysis
- Image and audio processing
- Data categorization and analysis
- API integration and automation

Each task is organized by season and episode (e.g., s01e01, s01e02, etc.)`,
	Example: `  # Run specific tasks
  ai-devs3 s01e01  # Robot authentication
  ai-devs3 s01e02  # RoboISO verification
  ai-devs3 s01e03  # JSON data processing
  ai-devs3 s01e05  # Text censoring
  ai-devs3 s02e01  # Audio transcription and analysis
  ai-devs3 s02e02  # Map analysis
  ai-devs3 s02e03  # Robot image generation
  ai-devs3 s02e04  # File categorization
  ai-devs3 s02e05  # Arxiv document analysis
  ai-devs3 s03e01  # Security reports processing
  ai-devs3 s03e04  # Barbara search task

  # Utility commands
  ai-devs3 ocr [image_url]                      # OCR text extraction

  # Get help for a specific task
  ai-devs3 s01e01 --help`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please check your environment variables:\n")
		fmt.Fprintf(os.Stderr, "  AI_DEVS_API_KEY: required\n")
		fmt.Fprintf(os.Stderr, "  OPENAI_API_KEY: required\n")
		fmt.Fprintf(os.Stderr, "  OLLAMA_BASE_URL: optional (default: http://localhost:11434)\n")
		fmt.Fprintf(os.Stderr, "  OLLAMA_MODEL: optional (default: llama3.2)\n")
		os.Exit(1)
	}

	// Add Season 1 tasks
	rootCmd.AddCommand(s01e01.NewCommand(cfg))
	rootCmd.AddCommand(s01e02.NewCommand(cfg))
	rootCmd.AddCommand(s01e03.NewCommand(cfg))
	rootCmd.AddCommand(s01e05.NewCommand(cfg))

	// Add Season 2 tasks
	rootCmd.AddCommand(s02e01.NewCommand(cfg))
	rootCmd.AddCommand(s02e02.NewCommand(cfg))
	rootCmd.AddCommand(s02e03.NewCommand(cfg))
	rootCmd.AddCommand(s02e04.NewCommand(cfg))
	rootCmd.AddCommand(s02e05.NewCommand(cfg))

	// Add Season 3 tasks
	rootCmd.AddCommand(s03e01.NewCommand(cfg))
	rootCmd.AddCommand(s03e02.NewCommand(cfg))
	rootCmd.AddCommand(s03e03.NewCommand(cfg))
	rootCmd.AddCommand(s03e04.NewCommand(cfg))
	rootCmd.AddCommand(s03e05.NewCommand(cfg))

	// Add utility commands
	rootCmd.AddCommand(ocr.NewCommand(cfg))
	rootCmd.AddCommand(video.NewCommand(cfg))

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ai-devs3 version 1.0.0")
		},
	})

	// Add list command to show available tasks
	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available tasks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available tasks:")
			fmt.Println()
			fmt.Println("Season 1:")
			fmt.Println("  s01e01  - Robot Authentication")
			fmt.Println("  s01e02  - RoboISO Verification")
			fmt.Println("  s01e03  - JSON Data Processing")
			fmt.Println("  s01e05  - Text Censoring")
			fmt.Println()
			fmt.Println("Season 2:")
			fmt.Println("  s02e01  - Audio Transcription and Analysis")
			fmt.Println("  s02e02  - Map Analysis")
			fmt.Println("  s02e03  - Robot Image Generation")
			fmt.Println("  s02e04  - File Categorization")
			fmt.Println("  s02e05  - Arxiv Document Analysis")
			fmt.Println()
			fmt.Println("Season 3:")
			fmt.Println("  s03e01  - Security Reports Processing")
			fmt.Println("  s03e02  - Weapon Reports Vector Search")
			fmt.Println("  s03e03  - Database Query Task")
			fmt.Println("  s03e04  - Barbara Search Task (loop)")
			fmt.Println("  s03e05  - Connections Task (Neo4j Graph)")
			fmt.Println()
			fmt.Println("Utilities:")
			fmt.Println("  ocr      - OCR Text Extraction")
			fmt.Println()
			fmt.Println("Use 'ai-devs3 <task> --help' for more information about a specific task.")
		},
	})
}
