package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSetCustomHelp(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "testcmd",
		Short: "Test command description",
	}
	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "Subcommand description",
	}
	cmd.AddCommand(subCmd)
	cmd.Flags().String("foo", "", "Foo flag")

	SetCustomHelp(cmd)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.HelpFunc()(cmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Scrape.do CLI & MCP Server") {
		t.Errorf("Expected banner and title, got: %s", output)
	}
	if !strings.Contains(output, "USAGE") {
		t.Error("Expected USAGE section")
	}
	if !strings.Contains(output, "testcmd") {
		t.Error("Expected command usage")
	}
	if !strings.Contains(output, "COMMANDS") {
		t.Error("Expected COMMANDS section")
	}
	if !strings.Contains(output, "sub") {
		t.Error("Expected subcommand name")
	}
	if !strings.Contains(output, "FLAGS") {
		t.Error("Expected FLAGS section")
	}
	if !strings.Contains(output, "--foo") {
		t.Error("Expected flag name")
	}
}
