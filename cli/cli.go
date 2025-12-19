// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	terminal "golang.org/x/term"

	setup_conf "github.com/mendersoftware/mender-setup/conf"

	"github.com/mendersoftware/mender-setup/conf"
)

const (
	appDescription = `mender-setup is a cli tool for generating the mender.conf` +
		` configuration files, either through specifying the parameters to the CLI,` +
		`or through running it interactively`
)

const (
	errMsgAmbiguousArgumentsGivenF = "Ambiguous arguments given - " +
		"unrecognized argument: %s"
	errMsgConflictingArgumentsF = "Conflicting arguments given, only one " +
		"of the following flags may be given: {%q, %q}"
)

type runOptionsType struct {
	config         string
	fallbackConfig string
	dataStore      string
	conf.HttpConfig
	setupOptions setupOptionsType // Options for setup subcommand
	logOptions   logOptionsType   // Options for logging
}

func ShowVersion() string {
	return fmt.Sprintf("%s\truntime: %s",
		setup_conf.VersionString(), runtime.Version())
}

// validateStringFlagValue validates that a string flag has a non-empty value
// that doesn't look like another flag (which would indicate a parsing issue).
func validateStringFlagValue(flagName string) func(*cli.Context, string) error {
	return func(ctx *cli.Context, value string) error {
		if value == "" || strings.HasPrefix(value, "-") {
			return fmt.Errorf("--%s requires a non-empty value", flagName)
		}
		return nil
	}
}

func SetupCLI(args []string) error {
	runOptions := &runOptionsType{}

	app := &cli.App{
		Description: appDescription,
		Name:        "mender-setup",
		Usage:       "Run to create a working mender-configuration file",
		ArgsUsage:   "[options]",
		Action:      runOptions.setupCLIHandler,
		Version:     ShowVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Destination: &runOptions.setupOptions.configPath,
				Value:       conf.DefaultConfFile,
				Usage:       "`PATH` to configuration file.",
				Action:      validateStringFlagValue("config"),
			},
			&cli.StringFlag{
				Name:    "data",
				Aliases: []string{"d"},
				Usage:   "Mender state data `DIR`ECTORY path.",
				Value:   conf.DefaultDataStore,
				Action:  validateStringFlagValue("data"),
			},
			&cli.StringFlag{
				Name:        "device-type",
				Destination: &runOptions.setupOptions.deviceType,
				Usage:       "Name of the device `type`.",
				Action:      validateStringFlagValue("device-type"),
			},
			&cli.StringFlag{
				Name:        "username",
				Destination: &runOptions.setupOptions.username,
				Usage:       "User `E-Mail` at hosted.mender.io.",
				Action:      validateStringFlagValue("username"),
			},
			&cli.StringFlag{
				Name:        "password",
				Destination: &runOptions.setupOptions.password,
				Usage:       "User `PASSWORD` at hosted.mender.io.",
				Action:      validateStringFlagValue("password"),
			},
			&cli.StringFlag{
				Name:        "server-url",
				Aliases:     []string{"url"},
				Destination: &runOptions.setupOptions.serverURL,
				Usage:       "`URL` to Mender server.",
				Value:       "https://docker.mender.io",
				Action:      validateStringFlagValue("server-url"),
			},
			&cli.StringFlag{
				Name:        "server-ip",
				Destination: &runOptions.setupOptions.serverIP,
				Usage:       "Server ip address.",
				Action:      validateStringFlagValue("server-ip"),
			},
			&cli.StringFlag{
				Name:        "server-cert",
				Aliases:     []string{"E"},
				Destination: &runOptions.setupOptions.serverCert,
				Usage:       "`PATH` to trusted server certificates",
				// No validator - empty string is valid (indicates no custom certificate)
			},
			&cli.StringFlag{
				Name:        "tenant-token",
				Destination: &runOptions.setupOptions.tenantToken,
				Usage:       "Hosted Mender tenant `token`",
				Action:      validateStringFlagValue("tenant-token"),
			},
			&cli.IntFlag{
				Name:        "inventory-poll",
				Destination: &runOptions.setupOptions.invPollInterval,
				Usage:       "Inventory poll interval in `sec`onds.",
				Value:       defaultInventoryPoll,
			},
			&cli.IntFlag{
				Name:        "retry-poll",
				Destination: &runOptions.setupOptions.retryPollInterval,
				Usage:       "Retry poll interval in `sec`onds.",
				Value:       defaultRetryPoll,
			},
			&cli.IntFlag{
				Name:        "update-poll",
				Destination: &runOptions.setupOptions.updatePollInterval,
				Usage:       "Update poll interval in `sec`onds.",
				Value:       defaultUpdatePoll,
			},
			&cli.BoolFlag{
				Name:        "hosted-mender",
				Destination: &runOptions.setupOptions.hostedMender,
				Usage:       "Setup device towards Hosted Mender.",
			},
			&cli.BoolFlag{
				Name:        "demo",
				Destination: &runOptions.setupOptions.demo,
				Usage: "Use demo configuration. DEPRECATED: use --demo-server and/or" +
					" --demo-polling instead",
			},
			&cli.BoolFlag{
				Name:        "demo-server",
				Destination: &runOptions.setupOptions.demoServer,
				Usage:       "Use demo server configuration.",
			},
			&cli.BoolFlag{
				Name:        "demo-polling",
				Destination: &runOptions.setupOptions.demoIntervals,
				Usage:       "Use demo polling intervals.",
			},
			&cli.BoolFlag{
				Name:  "quiet",
				Usage: "Suppress informative prompts.",
			},
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "Set logging `level`.",
				Value:       "warning",
				Destination: &runOptions.logOptions.logLevel,
				Action:      validateStringFlagValue("log-level"),
			},
		},
	}

	cli.HelpPrinter = upgradeHelpPrinter(cli.HelpPrinter)
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%s\n", ShowVersion())
	}
	return app.Run(args)
}

func (runOptions *runOptionsType) commonCLIHandler(
	ctx *cli.Context) (*conf.MenderConfig, error) {

	log.Debug("commonCLIHandler config file: ", runOptions.config)

	// Handle config flags
	config, err := conf.LoadConfig(
		runOptions.config, runOptions.fallbackConfig)
	if err != nil {
		return nil, err
	}

	// Make sure that paths that are not configurable via the config file is consistent with
	// --data flag
	config.ArtifactScriptsPath = path.Join(runOptions.dataStore, "scripts")
	config.ModulesWorkPath = path.Join(runOptions.dataStore, "modules", "v3")

	// Checks if the DeviceTypeFile is defined in config file.
	if config.MenderConfigFromFile.DeviceTypeFile != "" {
		// Sets the config.DeviceTypeFile to the value in config file.
		config.DeviceTypeFile = config.MenderConfigFromFile.DeviceTypeFile

	} else {
		config.MenderConfigFromFile.DeviceTypeFile = path.Join(
			runOptions.dataStore, "device_type")
		config.DeviceTypeFile = path.Join(
			runOptions.dataStore, "device_type")
	}

	if runOptions.HttpConfig.NoVerify {
		config.SkipVerify = true
	}

	return config, nil
}

func (runOptions *runOptionsType) handleCLIOptions(ctx *cli.Context) error {
	config, err := runOptions.commonCLIHandler(ctx)
	if err != nil {
		return err
	}

	// Execute commands
	// switch ctx.Command.Name {

	// Check that user has permission to directories so that
	// the user doesn't have to perform the setup before raising
	// an error.
	log.Debug("handleCLIOptions config file: ", runOptions.config)
	if err = checkWritePermissions(path.Dir(runOptions.config)); err != nil {
		return err
	}
	log.Debug("handleCLIOptions dataStore file: ", runOptions.dataStore)
	if err = checkWritePermissions(runOptions.dataStore); err != nil {
		return err
	}
	// Run cli setup prompts.
	if err := doSetup(ctx, &config.MenderConfigFromFile,
		&runOptions.setupOptions); err != nil {
		return err
	}
	if !ctx.Bool("quiet") {
		fmt.Println(promptDone)
	}

	return err
}

func (runOptions *runOptionsType) setupCLIHandler(ctx *cli.Context) error {
	if ctx.Args().Len() > 0 {
		return errors.Errorf(
			errMsgAmbiguousArgumentsGivenF,
			ctx.Args().First())
	}

	if ctx.Bool("quiet") {
		log.SetLevel(log.ErrorLevel)
	} else {
		if lvl, err := log.ParseLevel(ctx.String("log-level")); err == nil {
			log.SetLevel(lvl)
		} else {
			log.Warnf(
				"Failed to parse set log level '%s'.", ctx.String("log-level"))
		}
	}

	if err := runOptions.setupOptions.handleImplicitFlags(ctx); err != nil {
		return err
	}

	// Handle overlapping global flags
	if ctx.IsSet("config") && !ctx.IsSet("config") {
		runOptions.setupOptions.configPath = runOptions.config
	} else {
		runOptions.config = runOptions.setupOptions.configPath
	}
	runOptions.dataStore = ctx.String("data")
	if runOptions.HttpConfig.ServerCert != "" &&
		runOptions.setupOptions.serverCert == "" {
		runOptions.setupOptions.serverCert = runOptions.HttpConfig.ServerCert
	} else {
		runOptions.HttpConfig.ServerCert = runOptions.setupOptions.serverCert
	}
	return runOptions.handleCLIOptions(ctx)
}

func upgradeHelpPrinter(defaultPrinter func(w io.Writer, templ string, data interface{})) func(
	w io.Writer, templ string, data interface{}) {
	// Applies the ordinary help printer with column post processing
	return func(stdout io.Writer, templ string, data interface{}) {
		// Need at least 10 characters for last column in order to
		// pretty print; otherwise the output is unreadable.
		const minColumnWidth = 10
		isLowerCase := func(c rune) bool {
			// returns true if c in [a-z] else false
			asciiVal := int(c)
			if asciiVal >= 0x61 && asciiVal <= 0x7A {
				return true
			}
			return false
		}
		// defaultPrinter parses the text-template and outputs to buffer
		var buf bytes.Buffer
		defaultPrinter(&buf, templ, data)
		terminalWidth, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil || terminalWidth <= 0 {
			// Just write help as is.
			stdout.Write(buf.Bytes())
			return
		}
		for line, err := buf.ReadString('\n'); err == nil; line, err = buf.ReadString('\n') {
			if len(line) <= terminalWidth+1 {
				stdout.Write([]byte(line))
				continue
			}
			newLine := line
			indent := strings.LastIndex(
				line[:terminalWidth], "  ")
			// find indentation of last column
			if indent == -1 {
				indent = 0
			}
			indent += strings.IndexFunc(
				strings.ToLower(line[indent:]), isLowerCase) - 1
			if indent >= terminalWidth-minColumnWidth ||
				indent == -1 {
				indent = 0
			}
			// Format the last column to be aligned
			for len(newLine) > terminalWidth {
				// find word to insert newline
				idx := strings.LastIndex(newLine[:terminalWidth], " ")
				if idx == indent || idx == -1 {
					idx = terminalWidth
				}
				stdout.Write([]byte(newLine[:idx] + "\n"))
				newLine = newLine[idx:]
				newLine = strings.Repeat(" ", indent) + newLine
			}
			stdout.Write([]byte(newLine))
		}
		if err != nil {
			log.Fatalf("CLI HELP: error writing help string: %v\n", err)
		}
	}
}

func checkWritePermissions(dir string) error {
	log.Debug("Checking the permissions for: ", dir)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Wrapf(err, "Error creating "+
				"directory %q", dir)
		}
	} else if os.IsPermission(err) {
		return errors.Wrapf(os.ErrPermission,
			"Error trying to stat directory %q", dir)
	} else if err != nil {
		return errors.Errorf("Error trying to stat directory %q", dir)
	}
	f, err := os.MkdirTemp(dir, "temporaryFile")
	if os.IsPermission(err) {
		return errors.Wrapf(err, "User does not have "+
			"permission to write to data store "+
			"directory %q", dir)
	} else if err != nil {
		return errors.Wrapf(err,
			"Error checking write permissions to "+
				"directory %q", dir)
	}
	os.Remove(f)
	return nil
}
