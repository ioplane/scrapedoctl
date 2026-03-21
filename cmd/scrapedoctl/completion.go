package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// Shell completion installation paths (XDG / system standard locations).
// bash-completion >=2.0 lazy-loads from these directories automatically.
var completionPaths = map[string]struct {
	system string // requires root
	user   string // per-user, no root needed
}{
	shellBash: {
		system: "/usr/share/bash-completion/completions/scrapedoctl",
		user:   filepath.Join(xdgDataHome(), "bash-completion", "completions", "scrapedoctl"),
	},
	shellZsh: {
		system: "/usr/local/share/zsh/site-functions/_scrapedoctl",
		user:   filepath.Join(xdgDataHome(), shellZsh, "site-functions", "_scrapedoctl"),
	},
	shellFish: {
		system: "/usr/share/fish/vendor_completions.d/scrapedoctl.fish",
		user:   filepath.Join(xdgConfigHome(), shellFish, "completions", "scrapedoctl.fish"),
	},
}

const (
	shellBash       = "bash"
	shellZsh        = "zsh"
	shellFish       = "fish"
	shellPowershell = "powershell"
)

// errUnsupportedShell is returned for unknown shell types.
var errUnsupportedShell = errors.New("unsupported shell")

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate or install shell completions",
		Long: `Generate shell completion scripts for bash, zsh, fish, or powershell.
Use "completion install" to install to the standard directory automatically.`,
		ValidArgs: []string{shellBash, shellZsh, shellFish, shellPowershell},
		Args:      cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			w := os.Stdout
			switch args[0] {
			case shellBash:
				return root.GenBashCompletionV2(w, true)
			case shellZsh:
				return root.GenZshCompletion(w)
			case shellFish:
				return root.GenFishCompletion(w, true)
			case shellPowershell:
				return root.GenPowerShellCompletionWithDesc(w)
			default:
				return fmt.Errorf(
					"%w: %s", errUnsupportedShell, args[0],
				)
			}
		},
	}
	cmd.AddCommand(newCompletionInstallCmd(root))

	return cmd
}

func newCompletionInstallCmd(root *cobra.Command) *cobra.Command {
	var systemWide bool

	cmd := &cobra.Command{
		Use:   "install [bash|zsh|fish]",
		Short: "Install shell completions to the standard directory",
		Long: `Install shell completion scripts to the appropriate directory.

By default, installs to the per-user directory (no root required):
  bash: $XDG_DATA_HOME/bash-completion/completions/
  zsh:  $XDG_DATA_HOME/zsh/site-functions/
  fish: $XDG_CONFIG_HOME/fish/completions/

With --system, installs to system-wide directories (requires root):
  bash: /usr/share/bash-completion/completions/
  zsh:  /usr/local/share/zsh/site-functions/
  fish: /usr/share/fish/vendor_completions.d/

bash-completion >=2.0 lazy-loads completions from these directories
automatically — no need to modify .bashrc or .zshrc.`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{shellBash, shellZsh, shellFish},
		RunE: func(_ *cobra.Command, args []string) error {
			return installCompletion(root, args[0], systemWide)
		},
	}

	cmd.Flags().BoolVar(
		&systemWide, "system", false,
		"install system-wide (requires root)",
	)

	return cmd
}

func installCompletion(
	root *cobra.Command, shell string, systemWide bool,
) error {
	paths, ok := completionPaths[shell]
	if !ok {
		return fmt.Errorf(
			"%w: %s (supported: bash, zsh, fish)", errUnsupportedShell, shell,
		)
	}

	target := paths.user
	if systemWide {
		target = paths.system
	}

	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create %s: %w", target, err)
	}
	defer f.Close()

	switch shell {
	case shellBash:
		err = root.GenBashCompletionV2(f, true)
	case shellZsh:
		err = root.GenZshCompletion(f)
	case shellFish:
		err = root.GenFishCompletion(f, true)
	}

	if err != nil {
		return fmt.Errorf("generate %s completion: %w", shell, err)
	}

	fmt.Fprintf(os.Stdout, "Completion installed: %s\n", target)

	printPostInstall(shell, systemWide)

	return nil
}

func printPostInstall(shell string, systemWide bool) {
	switch shell {
	case shellBash:
		if systemWide {
			fmt.Println("Completions will load automatically in new shells.")
		} else {
			fmt.Println(
				"bash-completion >=2.0 loads from this directory automatically.",
			)
			fmt.Println(
				"If completions don't work, ensure bash-completion is installed:",
			)
			if runtime.GOOS == "darwin" {
				fmt.Println("  brew install bash-completion@2")
			} else {
				fmt.Println(
					"  dnf install bash-completion  # or: apt install bash-completion",
				)
			}
		}
	case shellZsh:
		fmt.Println("Run 'compinit' or start a new shell to activate.")
	case shellFish:
		fmt.Println("Completions will load automatically in new shells.")
	}
}

func xdgDataHome() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v
	}

	return filepath.Join(homeDir(), ".local", "share")
}

func xdgConfigHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}

	return filepath.Join(homeDir(), ".config")
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}

	return h
}
