package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Service implements the CLIService interface
type Service struct{}

var (
	ErrInvalidEndpointProtocol       = errors.New("Invalid environment protocol: Portainer only supports unix://, npipe:// or tcp://")
	ErrSocketOrNamedPipeNotFound     = errors.New("Unable to locate Unix socket or named pipe")
	ErrInvalidSnapshotInterval       = errors.New("Invalid snapshot interval")
	ErrAdminPassExcludeAdminPassFile = errors.New("Cannot use --admin-password with --admin-password-file")
)

func CLIFlags() *portainer.CLIFlags {
	return &portainer.CLIFlags{
		Addr:                      kingpin.Flag("bind", "Address and port to serve Portainer").Default(defaultBindAddress).Short('p').String(),
		AddrHTTPS:                 kingpin.Flag("bind-https", "Address and port to serve Portainer via https").Default(defaultHTTPSBindAddress).String(),
		TunnelAddr:                kingpin.Flag("tunnel-addr", "Address to serve the tunnel server").Default(defaultTunnelServerAddress).String(),
		TunnelPort:                kingpin.Flag("tunnel-port", "Port to serve the tunnel server").Default(defaultTunnelServerPort).String(),
		Assets:                    kingpin.Flag("assets", "Path to the assets").Default(defaultAssetsDirectory).Short('a').String(),
		Data:                      kingpin.Flag("data", "Path to the folder where the data is stored").Default(defaultDataDirectory).Short('d').String(),
		EndpointURL:               kingpin.Flag("host", "Environment URL").Short('H').String(),
		FeatureFlags:              kingpin.Flag("feat", "List of feature flags").Strings(),
		EnableEdgeComputeFeatures: kingpin.Flag("edge-compute", "Enable Edge Compute features").Bool(),
		NoAnalytics:               kingpin.Flag("no-analytics", "Disable Analytics in app (deprecated)").Bool(),
		TLS:                       kingpin.Flag("tlsverify", "TLS support").Default(defaultTLS).Bool(),
		TLSSkipVerify:             kingpin.Flag("tlsskipverify", "Disable TLS server verification").Default(defaultTLSSkipVerify).Bool(),
		TLSCacert:                 kingpin.Flag("tlscacert", "Path to the CA").Default(defaultTLSCACertPath).String(),
		TLSCert:                   kingpin.Flag("tlscert", "Path to the TLS certificate file").Default(defaultTLSCertPath).String(),
		TLSKey:                    kingpin.Flag("tlskey", "Path to the TLS key").Default(defaultTLSKeyPath).String(),
		HTTPDisabled:              kingpin.Flag("http-disabled", "Serve portainer only on https").Default(defaultHTTPDisabled).Bool(),
		HTTPEnabled:               kingpin.Flag("http-enabled", "Serve portainer on http").Default(defaultHTTPEnabled).Bool(),
		SSL:                       kingpin.Flag("ssl", "Secure Portainer instance using SSL (deprecated)").Default(defaultSSL).Bool(),
		SSLCert:                   kingpin.Flag("sslcert", "Path to the SSL certificate used to secure the Portainer instance").String(),
		SSLKey:                    kingpin.Flag("sslkey", "Path to the SSL key used to secure the Portainer instance").String(),
		Rollback:                  kingpin.Flag("rollback", "Rollback the database to the previous backup").Bool(),
		SnapshotInterval:          kingpin.Flag("snapshot-interval", "Duration between each environment snapshot job").String(),
		AdminPassword:             kingpin.Flag("admin-password", "Set admin password with provided hash").String(),
		AdminPasswordFile:         kingpin.Flag("admin-password-file", "Path to the file containing the password for the admin user").String(),
		Labels:                    pairs(kingpin.Flag("hide-label", "Hide containers with a specific label in the UI").Short('l')),
		Logo:                      kingpin.Flag("logo", "URL for the logo displayed in the UI").String(),
		Templates:                 kingpin.Flag("templates", "URL to the templates definitions.").Short('t').String(),
		BaseURL:                   kingpin.Flag("base-url", "Base URL parameter such as portainer if running portainer as http://yourdomain.com/portainer/.").Short('b').Default(defaultBaseURL).String(),
		InitialMmapSize:           kingpin.Flag("initial-mmap-size", "Initial mmap size of the database in bytes").Int(),
		MaxBatchSize:              kingpin.Flag("max-batch-size", "Maximum size of a batch").Int(),
		MaxBatchDelay:             kingpin.Flag("max-batch-delay", "Maximum delay before a batch starts").Duration(),
		SecretKeyName:             kingpin.Flag("secret-key-name", "Secret key name for encryption and will be used as /run/secrets/<secret-key-name>.").Default(defaultSecretKeyName).String(),
		LogLevel:                  kingpin.Flag("log-level", "Set the minimum logging level to show").Default("INFO").Enum("DEBUG", "INFO", "WARN", "ERROR"),
		LogMode:                   kingpin.Flag("log-mode", "Set the logging output mode").Default("PRETTY").Enum("NOCOLOR", "PRETTY", "JSON"),
		KubectlShellImage:         kingpin.Flag("kubectl-shell-image", "Kubectl shell image").Envar(portainer.KubectlShellImageEnvVar).Default(portainer.DefaultKubectlShellImage).String(),
		PullLimitCheckDisabled:    kingpin.Flag("pull-limit-check-disabled", "Pull limit check").Envar(portainer.PullLimitCheckDisabledEnvVar).Default(defaultPullLimitCheckDisabled).Bool(),
		TrustedOrigins:            kingpin.Flag("trusted-origins", "List of trusted origins for CSRF protection. Separate multiple origins with a comma.").Envar(portainer.TrustedOriginsEnvVar).String(),
		CSP:                       kingpin.Flag("csp", "Content Security Policy (CSP) header").Envar(portainer.CSPEnvVar).Default("true").Bool(),
	}
}

// ParseFlags parse the CLI flags and return a portainer.Flags struct
func (*Service) ParseFlags(version string) (*portainer.CLIFlags, error) {
	kingpin.Version(version)

	flags := CLIFlags()

	kingpin.Parse()

	if !filepath.IsAbs(*flags.Assets) {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}

		*flags.Assets = filepath.Join(filepath.Dir(ex), *flags.Assets)
	}

	return flags, nil
}

// ValidateFlags validates the values of the flags.
func (*Service) ValidateFlags(flags *portainer.CLIFlags) error {
	displayDeprecationWarnings(flags)

	if err := validateEndpointURL(*flags.EndpointURL); err != nil {
		return err
	}

	if err := validateSnapshotInterval(*flags.SnapshotInterval); err != nil {
		return err
	}

	if *flags.AdminPassword != "" && *flags.AdminPasswordFile != "" {
		return ErrAdminPassExcludeAdminPassFile
	}

	return nil
}

func displayDeprecationWarnings(flags *portainer.CLIFlags) {
	if *flags.NoAnalytics {
		log.Warn().Msg("the --no-analytics flag has been kept to allow migration of instances running a previous version of Portainer with this flag enabled, to version 2.0 where enabling this flag will have no effect")
	}

	if *flags.SSL {
		log.Warn().Msg("SSL is enabled by default and there is no need for the --ssl flag, it has been kept to allow migration of instances running a previous version of Portainer with this flag enabled")
	}
}

func validateEndpointURL(endpointURL string) error {
	if endpointURL == "" {
		return nil
	}

	if !strings.HasPrefix(endpointURL, "unix://") && !strings.HasPrefix(endpointURL, "tcp://") && !strings.HasPrefix(endpointURL, "npipe://") {
		return ErrInvalidEndpointProtocol
	}

	if strings.HasPrefix(endpointURL, "unix://") || strings.HasPrefix(endpointURL, "npipe://") {
		socketPath := strings.TrimPrefix(endpointURL, "unix://")
		socketPath = strings.TrimPrefix(socketPath, "npipe://")

		if _, err := os.Stat(socketPath); err != nil {
			if os.IsNotExist(err) {
				return ErrSocketOrNamedPipeNotFound
			}

			return err
		}
	}

	return nil
}

func validateSnapshotInterval(snapshotInterval string) error {
	if snapshotInterval == "" {
		return nil
	}

	if _, err := time.ParseDuration(snapshotInterval); err != nil {
		return ErrInvalidSnapshotInterval
	}

	return nil
}
