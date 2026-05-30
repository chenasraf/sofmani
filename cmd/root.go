package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/installer"
	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/machine"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var (
	// Flag variables
	debug           bool
	noDebug         bool
	update          bool
	noUpdate        bool
	summary         bool
	noSummary       bool
	filter          []string
	logFile         string
	machineID       bool
	showVars        bool
	ignoreFrequency bool
	startFrom       string

	// The parsed CLI config
	cliConfig *appconfig.AppCliConfig

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "sofmani [flags] [config_file]",
		Short: "Software manifest installer",
		Long: `Sofmani is a declarative software manifest installer.
It reads a configuration file and installs software based on the manifest.

For online documentation, see https://github.com/chenasraf/sofmani/tree/master/docs`,
		Args: cobra.MaximumNArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Build AppCliConfig from parsed flags
			cliConfig = buildCliConfig(cmd, args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Handle --log-file without value: show log file path and exit
			if cliConfig.ShowLogFile {
				fmt.Println(logger.GetLogFile())
				return
			}

			// Handle --machine-id: show machine ID and exit
			if cliConfig.ShowMachineID {
				fmt.Println(machine.GetMachineID())
				return
			}

			// Handle --vars: show template variables and exit
			if cliConfig.ShowVars {
				printTemplateVars(cliConfig.ConfigFile)
				return
			}

			// Run the main application logic
			RunMain(cliConfig)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Pre-process args to handle -l/--log-file with space-separated value
	// This maintains backward compatibility with the original CLI behavior
	os.Args = preprocessArgs(os.Args)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// preprocessArgs handles the -l/--log-file flag which has an optional value:
// - "-l" alone or "--log-file" alone -> show current log path
// - "-l value" or "--log-file value" -> set log file to value
// This transforms the args so Cobra can handle them properly.
func preprocessArgs(args []string) []string {
	result := make([]string, 0, len(args))
	i := 0
	for i < len(args) {
		arg := args[i]

		// Check if this is -l or --log-file without an = sign
		isLogFlag := arg == "-l" || arg == "--log-file"

		if isLogFlag {
			// Check if there's a next argument that could be the value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Next arg is the value - combine into -l=value format
				result = append(result, "-l="+args[i+1])
				i += 2
				continue
			} else {
				// No value provided - use sentinel to indicate "show log path"
				result = append(result, "-l=:show:")
				i++
				continue
			}
		}

		result = append(result, arg)
		i++
	}
	return result
}

// GetCliConfig returns the parsed CLI configuration.
func GetCliConfig() *appconfig.AppCliConfig {
	return cliConfig
}

func init() {
	// Disable alphabetical sorting to control flag order in help output
	rootCmd.Flags().SortFlags = false

	// Boolean flags with negation variants (grouped together)
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.Flags().BoolVarP(&noDebug, "no-debug", "D", false, "Disable debug mode")
	rootCmd.Flags().BoolVarP(&update, "update", "u", false, "Enable update checks")
	rootCmd.Flags().BoolVarP(&noUpdate, "no-update", "U", false, "Disable update checks")
	rootCmd.Flags().BoolVarP(&summary, "summary", "s", false, "Enable installation summary")
	rootCmd.Flags().BoolVarP(&noSummary, "no-summary", "S", false, "Disable installation summary")

	// Filter flag (repeatable)
	rootCmd.Flags().StringArrayVarP(&filter, "filter", "f", nil, "Filter by installer name (can be used multiple times)")

	// Log file flag - optional value handled via arg preprocessing
	rootCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Set log file path (use flag alone to show current path)")

	// Machine ID flag
	rootCmd.Flags().BoolVarP(&machineID, "machine-id", "m", false, "Show machine ID and exit")

	// Template variables flag
	rootCmd.Flags().BoolVar(&showVars, "vars", false, "Show template variables and their values for the current platform, then exit")

	// Ignore frequency flag
	rootCmd.Flags().BoolVar(&ignoreFrequency, "ignore-frequency", false, "Ignore frequency limits and run all installers")

	// Start-from flag
	rootCmd.Flags().StringVar(&startFrom, "start-from", "", "Skip all installers before the one with the given name")
}

// SetVersion sets the version for the root command.
func SetVersion(version string) {
	rootCmd.Version = version
	// Use custom template to match original output format (just version number)
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	appconfig.SetVersion(version)
}

// buildCliConfig creates an AppCliConfig from the parsed Cobra flags.
func buildCliConfig(cmd *cobra.Command, args []string) *appconfig.AppCliConfig {
	config := &appconfig.AppCliConfig{
		ConfigFile:      "",
		Debug:           nil,
		CheckUpdates:    nil,
		Summary:         nil,
		Filter:          filter,
		LogFile:         nil,
		ShowLogFile:     false,
		ShowMachineID:   machineID,
		ShowVars:        showVars,
		IgnoreFrequency: ignoreFrequency,
		StartFrom:       startFrom,
	}

	// Handle debug flag
	if cmd.Flags().Changed("debug") {
		config.Debug = lo.ToPtr(true)
	}
	if cmd.Flags().Changed("no-debug") {
		config.Debug = lo.ToPtr(false)
	}

	// Handle update flag
	if cmd.Flags().Changed("update") {
		config.CheckUpdates = lo.ToPtr(true)
	}
	if cmd.Flags().Changed("no-update") {
		config.CheckUpdates = lo.ToPtr(false)
	}

	// Handle summary flag
	if cmd.Flags().Changed("summary") {
		config.Summary = lo.ToPtr(true)
	}
	if cmd.Flags().Changed("no-summary") {
		config.Summary = lo.ToPtr(false)
	}

	// Handle log file flag
	if cmd.Flags().Changed("log-file") {
		if logFile == ":show:" {
			// Flag was provided without a value
			config.ShowLogFile = true
		} else {
			config.LogFile = &logFile
		}
	}

	// Handle config file positional argument
	switch {
	case len(args) > 0:
		config.ConfigFile = args[0]
	case config.ShowVars:
		// --vars tries to read machine_aliases from a config if one exists, but does not
		// require it. Best-effort lookup only.
		config.ConfigFile = appconfig.FindConfigFile()
	case !config.ShowLogFile && !config.ShowMachineID:
		// Find config file if not showing log file or machine ID
		file := appconfig.FindConfigFile()
		if file == "" {
			fmt.Fprintln(os.Stderr, "No config file found")
			os.Exit(1)
		}
		config.ConfigFile = file
	}

	return config
}

// printTemplateVars prints all template variables with their resolved values for the current
// platform. If configFile is non-empty and parses successfully, machine_aliases is loaded so
// {{ .DeviceIDAlias }} can be resolved; otherwise it is shown as unset.
func printTemplateVars(configFile string) {
	var machineAliases map[string]string
	if configFile != "" {
		if cfg, err := appconfig.ParseConfigFrom(configFile); err == nil && cfg != nil && cfg.MachineAliases != nil {
			machineAliases = *cfg.MachineAliases
		}
	}
	vars := installer.NewTemplateVars("", machineAliases)
	descs := installer.DescribeTemplateVars(vars)

	nameWidth := 0
	for _, d := range descs {
		if len(d.Name) > nameWidth {
			nameWidth = len(d.Name)
		}
	}

	for _, d := range descs {
		value := d.Value
		if value == "" {
			if d.Note != "" {
				value = "<" + d.Note + ">"
			} else {
				value = "<unset>"
			}
		} else if d.Note != "" {
			value = fmt.Sprintf("%s  <%s>", value, d.Note)
		}
		fmt.Printf("%-*s  %s\n", nameWidth, d.Name, value)
	}
}

// RunMain is set by main.go to run the main application logic.
var RunMain func(cliConfig *appconfig.AppCliConfig)
