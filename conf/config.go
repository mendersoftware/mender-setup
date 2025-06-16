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
package conf

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultUpdateControlMapBootExpirationTimeSeconds = 600
)

// MenderServer is a placeholder for a full server definition used when
// multiple servers are given. The fields corresponds to the definitions
// given in MenderConfig.
type MenderServer struct {
	ServerURL string
	// TODO: Move all possible server specific configurations in
	//       MenderConfig over to this struct. (e.g. TenantToken?)
}

type Security struct {
	AuthPrivateKey string `json:",omitempty"`
	SSLEngine      string `json:",omitempty"`
}

type MenderConfigFromFile struct {
	// Path to the public key used to verify signed updates.
	// Only one of ArtifactVerifyKey/ArtifactVerifyKeys can be specified.
	ArtifactVerifyKey string `json:",omitempty"`
	// List of verification keys for verifying signed updates.
	// Starting in order from the first key in the list,
	// each key will try to verify the artifact until one succeeds.
	// Only one of ArtifactVerifyKey/ArtifactVerifyKeys can be specified.
	ArtifactVerifyKeys []string `json:",omitempty"`

	// HTTPS client parameters
	HttpsClient HttpsClient `json:",omitempty"`
	// Security parameters
	Security Security `json:",omitempty"`
	// Connectivity connection handling and transfer parameters
	Connectivity Connectivity `json:",omitempty"`

	// Rootfs device path
	RootfsPartA string `json:",omitempty"`
	RootfsPartB string `json:",omitempty"`

	// Command to set active partition.
	BootUtilitiesSetActivePart string `json:",omitempty"`
	// Command to get the partition which will boot next.
	BootUtilitiesGetNextActivePart string `json:",omitempty"`

	// Path to the device type file
	DeviceTypeFile string `json:",omitempty"`

	// Expiration timeout for the control map
	UpdateControlMapExpirationTimeSeconds int `json:",omitempty"`
	// Expiration timeout for the control map when just booted
	UpdateControlMapBootExpirationTimeSeconds int `json:",omitempty"`

	// Poll interval for checking for new updates
	UpdatePollIntervalSeconds int `json:",omitempty"`
	// Poll interval for periodically sending inventory data
	InventoryPollIntervalSeconds int `json:",omitempty"`

	// Skip CA certificate validation
	SkipVerify bool `json:",omitempty"`

	// Global retry polling max interval for fetching update, authorize wait and update status
	RetryPollIntervalSeconds int `json:",omitempty"`
	// Global max retry poll count
	RetryPollCount int `json:",omitempty"`

	// State script parameters
	StateScriptTimeoutSeconds      int `json:",omitempty"`
	StateScriptRetryTimeoutSeconds int `json:",omitempty"`
	// Poll interval for checking for update (check-update)
	StateScriptRetryIntervalSeconds int `json:",omitempty"`

	// Update module parameters:

	// The timeout for the execution of the update module, after which it
	// will be killed.
	ModuleTimeoutSeconds int `json:",omitempty"`

	// Path to server SSL certificate
	ServerCertificate string `json:",omitempty"`
	// Server URL (For single server conf)
	ServerURL string `json:",omitempty"`
	// Path to deployment log file
	UpdateLogPath string `json:",omitempty"`
	// Server JWT TenantToken
	TenantToken string `json:",omitempty"`
	// List of available servers, to which client can fall over
	Servers []MenderServer `json:",omitempty"`
	// Log level which takes effect right before daemon startup
	DaemonLogLevel string `json:",omitempty"`
}

// HttpsClient holds the configuration for the client side mTLS configuration
// NOTE: Careful when changing this, the struct is exposed directly in the
// 'mender.conf' file.
type HttpsClient struct {
	Certificate string `json:",omitempty"`
	Key         string `json:",omitempty"`
	SSLEngine   string `json:",omitempty"`
}

// Connectivity instructs the client how we want to treat the keep alive connections
// and when a connection is considered idle and therefore closed
// NOTE: Careful when changing this, the struct is exposed directly in the
// 'mender.conf' file.
type Connectivity struct {
	// If set to true, there will be no persistent connections, and every
	// HTTP transaction will try to establish a new connection
	DisableKeepAlive bool `json:",omitempty"`
	// A number of seconds after which a connection is considered idle and closed.
	// The longer this is the longer connections are up after the first call over HTTP
	IdleConnTimeoutSeconds int `json:",omitempty"`
}

type HttpConfig struct {
	ServerCert string
	*HttpsClient
	*Connectivity
	NoVerify bool
}

type MenderConfig struct {
	MenderConfigFromFile

	// Additional fields that are in our config struct for convenience, but
	// not actually configurable via the config file.
	ModulesPath     string
	ModulesWorkPath string

	ArtifactScriptsPath string
	RootfsScriptsPath   string

	BootstrapArtifactFile string
}

func NewMenderConfig() *MenderConfig {
	return &MenderConfig{
		MenderConfigFromFile:  MenderConfigFromFile{},
		ModulesPath:           DefaultModulesPath,
		ModulesWorkPath:       DefaultModulesWorkPath,
		ArtifactScriptsPath:   DefaultArtScriptsPath,
		RootfsScriptsPath:     DefaultRootfsScriptsPath,
		BootstrapArtifactFile: DefaultBootstrapArtifactFile,
	}
}

func LoadConfig(mainConfigFile string, fallbackConfigFile string) (*MenderConfig, error) {
	// Load fallback configuration first, then main configuration.
	// It is OK if either file does not exist, so long as the other one does exist.
	// It is also OK if both files exist.
	// Because the main configuration is loaded last, its option values
	// override those from the fallback file, for options present in both files.

	var filesLoadedCount int
	config := NewMenderConfig()

	if loadErr := loadConfigFile(fallbackConfigFile, config, &filesLoadedCount); loadErr != nil {
		return nil, loadErr
	}

	if loadErr := loadConfigFile(mainConfigFile, config, &filesLoadedCount); loadErr != nil {
		return nil, loadErr
	}

	log.Debugf("Loaded %d configuration file(s)", filesLoadedCount)

	checkConfigDefaults(config)

	if filesLoadedCount == 0 {
		log.Info("No configuration files present. Using defaults")
		return config, nil
	}

	log.Debugf("Loaded configuration = %#v", config)

	return config, nil
}

func loadConfigFile(configFile string, config *MenderConfig, filesLoadedCount *int) error {
	// Do not treat a single config file not existing as an error here.
	// It is up to the caller to fail when both config files don't exist.
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Debug("Configuration file does not exist: ", configFile)
		return nil
	}

	if err := readConfigFile(&config.MenderConfigFromFile, configFile); err != nil {
		log.Errorf("Error loading configuration from file: %s (%s)", configFile, err.Error())
		return err
	}

	if config.ArtifactVerifyKey != "" {
		if len(config.ArtifactVerifyKeys) > 0 {
			return errors.New("both ArtifactVerifyKey and ArtifactVerifyKeys are set")
		}
		// Unify the logic for verification key processing by moving
		// the single ArtifactVerifyKey to the list version.
		config.ArtifactVerifyKeys = append(config.ArtifactVerifyKeys, config.ArtifactVerifyKey)
		config.ArtifactVerifyKey = ""
	}

	(*filesLoadedCount)++
	log.Info("Loaded configuration file: ", configFile)
	return nil
}

func readConfigFile(config interface{}, fileName string) error {
	// Reads mender configuration (JSON) file.

	log.Debug("Reading Mender configuration from file " + fileName)
	conf, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(conf, &config); err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return errors.New("Error parsing mender configuration file: " + err.Error())
		}
		return errors.New("Error parsing config file: " + err.Error())
	}

	return nil
}

func checkConfigDefaults(config *MenderConfig) {
	if config.MenderConfigFromFile.UpdateControlMapExpirationTimeSeconds == 0 {
		log.Info(
			"'UpdateControlMapExpirationTimeSeconds' is not set " +
				"in the Mender configuration file." +
				" Falling back to the default of 2*UpdatePollIntervalSeconds")
	}

	if config.MenderConfigFromFile.UpdateControlMapBootExpirationTimeSeconds == 0 {
		log.Infof(
			"'UpdateControlMapBootExpirationTimeSeconds' is not set "+
				"in the Mender configuration file."+
				" Falling back to the default of %d seconds",
			DefaultUpdateControlMapBootExpirationTimeSeconds,
		)
	}
}

func SaveConfigFile(config *MenderConfigFromFile, filename string) error {
	configJson, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return errors.Wrap(err, "Error encoding configuration to JSON")
	}
	f, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	) // for mode see MEN-3762
	if err != nil {
		return errors.Wrap(err, "Error opening configuration file")
	}
	defer f.Close()

	if _, err = f.Write(configJson); err != nil {
		return errors.Wrap(err, "Error writing to configuration file")
	}
	return nil
}
