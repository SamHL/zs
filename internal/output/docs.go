package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandDoc represents documentation for a command
type CommandDoc struct {
	Name        string       `json:"name" yaml:"name"`
	FullCommand string       `json:"full_command" yaml:"full_command"`
	Description string       `json:"description" yaml:"description"`
	Usage       string       `json:"usage" yaml:"usage"`
	Examples    []string     `json:"examples,omitempty" yaml:"examples,omitempty"`
	Flags       []FlagDoc    `json:"flags,omitempty" yaml:"flags,omitempty"`
	Subcommands []CommandDoc `json:"subcommands,omitempty" yaml:"subcommands,omitempty"`
}

// FlagDoc represents documentation for a flag
type FlagDoc struct {
	Name        string `json:"name" yaml:"name"`
	Shorthand   string `json:"shorthand,omitempty" yaml:"shorthand,omitempty"`
	Type        string `json:"type" yaml:"type"`
	Default     string `json:"default,omitempty" yaml:"default,omitempty"`
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`
}

// CLIDocs represents the full CLI documentation
type CLIDocs struct {
	Name        string       `json:"name" yaml:"name"`
	Version     string       `json:"version" yaml:"version"`
	Description string       `json:"description" yaml:"description"`
	Commands    []CommandDoc `json:"commands" yaml:"commands"`
	GlobalFlags []FlagDoc    `json:"global_flags" yaml:"global_flags"`
	APIInfo     APIInfo      `json:"api_info" yaml:"api_info"`
}

// APIInfo contains API-related information
type APIInfo struct {
	BaseURLs   map[string]string `json:"base_urls" yaml:"base_urls"`
	RateLimit  string            `json:"rate_limit" yaml:"rate_limit"`
	AuthMethod string            `json:"auth_method" yaml:"auth_method"`
	Scopes     []string          `json:"scopes" yaml:"scopes"`
}

// GenerateDocs generates full CLI documentation
func GenerateDocs(rootCmd *cobra.Command, version string) *CLIDocs {
	docs := &CLIDocs{
		Name:        "zs",
		Version:     version,
		Description: "CLI tool for managing Zoho Sprints - create, update, and track sprints, items, and projects from the command line.",
		Commands:    []CommandDoc{},
		GlobalFlags: []FlagDoc{},
		APIInfo: APIInfo{
			BaseURLs: map[string]string{
				"US": "https://sprints.zoho.com/zsapi",
				"EU": "https://sprints.zoho.eu/zsapi",
				"IN": "https://sprints.zoho.in/zsapi",
				"AU": "https://sprints.zoho.com.au/zsapi",
				"JP": "https://sprints.zoho.jp/zsapi",
				"CA": "https://sprints.zohocloud.ca/zsapi",
			},
			RateLimit:  "100 requests per 2 minutes",
			AuthMethod: "OAuth 2.0 with Zoho-oauthtoken header",
			Scopes: []string{
				"ZohoSprints.teams.READ",
				"ZohoSprints.projects.ALL",
				"ZohoSprints.sprints.ALL",
				"ZohoSprints.items.ALL",
				"ZohoSprints.epics.ALL",
			},
		},
	}

	// Extract global flags
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		docs.GlobalFlags = append(docs.GlobalFlags, FlagDoc{
			Name:        flag.Name,
			Shorthand:   flag.Shorthand,
			Type:        flag.Value.Type(),
			Default:     flag.DefValue,
			Description: flag.Usage,
		})
	})

	// Extract subcommands
	for _, cmd := range rootCmd.Commands() {
		if !cmd.Hidden {
			docs.Commands = append(docs.Commands, generateCommandDoc(cmd, rootCmd.Name()))
		}
	}

	return docs
}

// generateCommandDoc generates documentation for a single command
func generateCommandDoc(cmd *cobra.Command, parentPath string) CommandDoc {
	fullCmd := parentPath + " " + cmd.Name()

	doc := CommandDoc{
		Name:        cmd.Name(),
		FullCommand: fullCmd,
		Description: cmd.Short,
		Usage:       cmd.UseLine(),
		Examples:    []string{},
		Flags:       []FlagDoc{},
		Subcommands: []CommandDoc{},
	}

	// Add long description if different
	if cmd.Long != "" && cmd.Long != cmd.Short {
		doc.Description = cmd.Long
	}

	// Extract examples
	if cmd.Example != "" {
		examples := strings.Split(cmd.Example, "\n")
		for _, ex := range examples {
			ex = strings.TrimSpace(ex)
			if ex != "" {
				doc.Examples = append(doc.Examples, ex)
			}
		}
	}

	// Extract local flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden {
			doc.Flags = append(doc.Flags, FlagDoc{
				Name:        flag.Name,
				Shorthand:   flag.Shorthand,
				Type:        flag.Value.Type(),
				Default:     flag.DefValue,
				Description: flag.Usage,
			})
		}
	})

	// Extract subcommands
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden {
			doc.Subcommands = append(doc.Subcommands, generateCommandDoc(subCmd, fullCmd))
		}
	}

	return doc
}

// GenerateMarkdown generates Markdown documentation
func GenerateMarkdown(docs *CLIDocs) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s CLI Documentation\n\n", docs.Name))
	sb.WriteString(fmt.Sprintf("Version: %s\n\n", docs.Version))
	sb.WriteString(fmt.Sprintf("%s\n\n", docs.Description))

	// API Info
	sb.WriteString("## API Information\n\n")
	sb.WriteString(fmt.Sprintf("- **Authentication**: %s\n", docs.APIInfo.AuthMethod))
	sb.WriteString(fmt.Sprintf("- **Rate Limit**: %s\n", docs.APIInfo.RateLimit))
	sb.WriteString("- **Required Scopes**:\n")
	for _, scope := range docs.APIInfo.Scopes {
		sb.WriteString(fmt.Sprintf("  - `%s`\n", scope))
	}
	sb.WriteString("\n### Data Center URLs\n\n")
	sb.WriteString("| Region | Base URL |\n|--------|----------|\n")
	for region, url := range docs.APIInfo.BaseURLs {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", region, url))
	}
	sb.WriteString("\n")

	// Global Flags
	if len(docs.GlobalFlags) > 0 {
		sb.WriteString("## Global Flags\n\n")
		sb.WriteString("These flags are available for all commands:\n\n")
		sb.WriteString("| Flag | Type | Default | Description |\n|------|------|---------|-------------|\n")
		for _, flag := range docs.GlobalFlags {
			name := "--" + flag.Name
			if flag.Shorthand != "" {
				name = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", name, flag.Type, flag.Default, flag.Description))
		}
		sb.WriteString("\n")
	}

	// Commands
	sb.WriteString("## Commands\n\n")
	for _, cmd := range docs.Commands {
		writeCommandMarkdown(&sb, cmd, 3)
	}

	return sb.String()
}

// writeCommandMarkdown writes a command's documentation in Markdown
func writeCommandMarkdown(sb *strings.Builder, cmd CommandDoc, headingLevel int) {
	heading := strings.Repeat("#", headingLevel)
	sb.WriteString(fmt.Sprintf("%s %s\n\n", heading, cmd.Name))
	sb.WriteString(fmt.Sprintf("%s\n\n", cmd.Description))
	sb.WriteString(fmt.Sprintf("**Usage**: `%s`\n\n", cmd.Usage))

	// Flags
	if len(cmd.Flags) > 0 {
		sb.WriteString("**Flags**:\n\n")
		sb.WriteString("| Flag | Type | Default | Description |\n|------|------|---------|-------------|\n")
		for _, flag := range cmd.Flags {
			name := "--" + flag.Name
			if flag.Shorthand != "" {
				name = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", name, flag.Type, flag.Default, flag.Description))
		}
		sb.WriteString("\n")
	}

	// Examples
	if len(cmd.Examples) > 0 {
		sb.WriteString("**Examples**:\n\n```bash\n")
		for _, ex := range cmd.Examples {
			sb.WriteString(ex + "\n")
		}
		sb.WriteString("```\n\n")
	}

	// Subcommands
	for _, subCmd := range cmd.Subcommands {
		writeCommandMarkdown(sb, subCmd, headingLevel+1)
	}
}

// GenerateJSON generates JSON documentation
func GenerateJSON(docs *CLIDocs) (string, error) {
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GenerateCommandDocs generates documentation for a specific command
func GenerateCommandDocs(cmd *cobra.Command, format string) (string, error) {
	doc := generateCommandDoc(cmd, "zs")

	switch format {
	case "json":
		data, err := json.MarshalIndent(doc, "", "  ")
		return string(data), err
	case "markdown", "md":
		var sb strings.Builder
		writeCommandMarkdown(&sb, doc, 2)
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
