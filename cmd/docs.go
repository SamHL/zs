package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/output"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate LLM-friendly documentation",
	Long: `Generate comprehensive documentation for the CLI in a format
suitable for LLM consumption.

This command outputs documentation about all available commands,
their flags, usage examples, and API information. The output can
be in Markdown or JSON format for easy parsing by LLMs.`,
	Example: `  # Generate full docs in Markdown
  zs docs

  # Generate docs in JSON format
  zs docs --format json

  # Generate docs for a specific command
  zs docs --command items

  # Generate JSON docs for a specific command
  zs docs --command sprints --format json`,
	RunE: runDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)

	docsCmd.Flags().StringP("format", "f", "markdown", "output format: markdown, json")
	docsCmd.Flags().StringP("command", "c", "", "generate docs for a specific command")
}

func runDocs(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	commandName, _ := cmd.Flags().GetString("command")

	format = strings.ToLower(format)

	// If a specific command is requested
	if commandName != "" {
		return generateCommandDocs(commandName, format)
	}

	// Generate full documentation
	return generateFullDocs(format)
}

func generateFullDocs(format string) error {
	docs := output.GenerateDocs(rootCmd, version)

	switch format {
	case "json":
		jsonStr, err := output.GenerateJSON(docs)
		if err != nil {
			return fmt.Errorf("failed to generate JSON: %w", err)
		}
		fmt.Println(jsonStr)

	case "markdown", "md":
		mdStr := output.GenerateMarkdown(docs)
		fmt.Println(mdStr)

	default:
		return fmt.Errorf("unsupported format: %s (use 'markdown' or 'json')", format)
	}

	return nil
}

func generateCommandDocs(commandName string, format string) error {
	// Find the command
	var targetCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == commandName {
			targetCmd = cmd
			break
		}
	}

	if targetCmd == nil {
		return fmt.Errorf("command not found: %s", commandName)
	}

	docStr, err := output.GenerateCommandDocs(targetCmd, format)
	if err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	fmt.Println(docStr)
	return nil
}
