package main

import (
	"encoding/json"
	"fmt"
	"io"
)

const codexCapabilitiesSchemaVersion = "codex-cli-capabilities/v1"

type codexCapabilitiesOptions struct {
	format   string
	showHelp bool
}

type codexCapabilitiesResult struct {
	SchemaVersion    string `json:"schema_version"`
	CLI              string `json:"cli"`
	Version          string `json:"version"`
	Prepare          bool   `json:"prepare"`
	ValidateComments bool   `json:"validate_comments"`
	Report           bool   `json:"report"`
	Context          bool   `json:"context"`
	Scan             bool   `json:"scan"`
	Session          bool   `json:"session"`
	ReadOnlyFallback bool   `json:"read_only_fallback"`
}

func runCodexCapabilities(args []string, writer io.Writer) error {
	options, err := parseCodexCapabilitiesFlags(args)
	if err != nil {
		return err
	}
	if options.showHelp {
		printCodexCapabilitiesUsage(writer)
		return nil
	}
	result := codexCapabilitiesResult{
		SchemaVersion:    codexCapabilitiesSchemaVersion,
		CLI:              "ocr",
		Version:          Version,
		Prepare:          true,
		ValidateComments: true,
		Report:           true,
		Context:          true,
		Scan:             true,
		Session:          true,
		ReadOnlyFallback: true,
	}
	switch options.format {
	case "json":
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("encode codex capabilities: %w", err)
		}
		_, err = writer.Write(append(encoded, '\n'))
		return err
	case "text":
		fmt.Fprintf(writer, "CLI: %s\n", result.CLI)
		fmt.Fprintf(writer, "Version: %s\n", result.Version)
		fmt.Fprintf(writer, "prepare: %t\n", result.Prepare)
		fmt.Fprintf(writer, "validate-comments: %t\n", result.ValidateComments)
		fmt.Fprintf(writer, "report: %t\n", result.Report)
		fmt.Fprintf(writer, "context: %t\n", result.Context)
		fmt.Fprintf(writer, "scan: %t\n", result.Scan)
		fmt.Fprintf(writer, "session: %t\n", result.Session)
		fmt.Fprintf(writer, "read-only fallback: %t\n", result.ReadOnlyFallback)
		return nil
	default:
		return fmt.Errorf("invalid --format value %q: must be json or text", options.format)
	}
}

func parseCodexCapabilitiesFlags(args []string) (codexCapabilitiesOptions, error) {
	flags := newOcrFlagSet("ocr codex capabilities")
	options := codexCapabilitiesOptions{}
	flags.StringVarP(&options.format, "format", "f", "json", "output format: json or text")
	if err := flags.Parse(args); err != nil {
		return options, fmt.Errorf("parse flags: %w", err)
	}
	options.showHelp = flags.showHelp
	if options.showHelp {
		return options, nil
	}
	switch options.format {
	case "json", "text":
	default:
		return options, fmt.Errorf("invalid --format value %q: must be json or text", options.format)
	}
	return options, nil
}

func printCodexCapabilitiesUsage(writer io.Writer) {
	fmt.Fprintln(writer, `Usage:
  ocr codex capabilities [--format json|text]

Description:
  Report which Codex-owned review capabilities this binary supports.
  Codex CLI can probe this command before deciding whether to run the
  primary Codex workflow or a read-only fallback.`)
}
