package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	"github.com/SamHL/zs/internal/config"
)

// Format represents an output format
type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatTable Format = "table"
	FormatPlain Format = "plain"
)

// Formatter handles output formatting
type Formatter struct {
	format Format
	writer io.Writer
	color  bool
}

// NewFormatter creates a new formatter with default settings
func NewFormatter() *Formatter {
	cfg := config.Get()
	return &Formatter{
		format: Format(cfg.Output.Format),
		writer: os.Stdout,
		color:  cfg.Output.Color,
	}
}

// WithFormat sets the output format
func (f *Formatter) WithFormat(format string) *Formatter {
	f.format = Format(strings.ToLower(format))
	return f
}

// WithWriter sets the output writer
func (f *Formatter) WithWriter(w io.Writer) *Formatter {
	f.writer = w
	return f
}

// WithColor sets whether to use colors
func (f *Formatter) WithColor(color bool) *Formatter {
	f.color = color
	return f
}

// Print outputs data in the configured format
func (f *Formatter) Print(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.printJSON(data)
	case FormatYAML:
		return f.printYAML(data)
	case FormatTable:
		return f.printTable(data)
	case FormatPlain:
		return f.printPlain(data)
	default:
		return f.printTable(data)
	}
}

// printJSON outputs data as formatted JSON
func (f *Formatter) printJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printYAML outputs data as YAML
func (f *Formatter) printYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

// printPlain outputs data as plain text
func (f *Formatter) printPlain(data interface{}) error {
	switch v := data.(type) {
	case string:
		fmt.Fprintln(f.writer, v)
	case []string:
		for _, s := range v {
			fmt.Fprintln(f.writer, s)
		}
	case fmt.Stringer:
		fmt.Fprintln(f.writer, v.String())
	default:
		// Fall back to JSON for complex types
		return f.printJSON(data)
	}
	return nil
}

// printTable outputs data as a table
func (f *Formatter) printTable(data interface{}) error {
	// For table output, we need to handle different types
	switch v := data.(type) {
	case TableData:
		return f.renderTable(v.Headers, v.Rows)
	case *TableData:
		return f.renderTable(v.Headers, v.Rows)
	default:
		// Fall back to JSON for types we can't table-ify
		return f.printJSON(data)
	}
}

// renderTable renders a table with headers and rows
func (f *Formatter) renderTable(headers []string, rows [][]string) error {
	table := tablewriter.NewWriter(f.writer)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	for _, row := range rows {
		table.Append(row)
	}

	table.Render()
	return nil
}

// TableData represents data for table output
type TableData struct {
	Headers []string
	Rows    [][]string
}

// NewTableData creates a new TableData
func NewTableData(headers ...string) *TableData {
	return &TableData{
		Headers: headers,
		Rows:    [][]string{},
	}
}

// AddRow adds a row to the table
func (t *TableData) AddRow(values ...string) {
	t.Rows = append(t.Rows, values)
}

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println("✓", msg)
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, "✗", msg)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println("⚠", msg)
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println("ℹ", msg)
}

// Confirm prompts the user for confirmation
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// GetFormat returns the current output format from config or flag
func GetFormat(flagValue string) Format {
	if flagValue != "" {
		return Format(strings.ToLower(flagValue))
	}
	return Format(config.Get().Output.Format)
}
