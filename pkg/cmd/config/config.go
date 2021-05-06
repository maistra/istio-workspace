package config

import (
	"path"
	"strings"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// SetupConfig defines Viper env var prefixes and type handling when inferring key value.
func SetupConfig() {
	viper.Reset()
	viper.SetEnvPrefix("IKE")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetTypeByDefaultValue(true)
}

// SetupConfigSources sets up Viper configuration sources.
//
// If specific file path is provided but fails when loading it will return an error.
//
// In case of default config location it will not fail if file does not exist,
// but will in any other case such as parse error.
//
// Config precedence (each item takes precedence over the item below it):
// . Flags
// . Env variables
// . Config file
//
// Environment variables are prefixed with `IKE` and have fully qualified names, for example
// in case of `develop` command and its `port` flag corresponding environment variable is
// `IKE_DEVELOP_PORT`.
func SetupConfigSources(configFile string, defaultConfigFile bool) error {
	ext := path.Ext(configFile)
	viper.SetConfigName(strings.TrimSuffix(path.Base(configFile), ext))
	if !contains(SupportedExtensions(), strings.TrimPrefix(path.Ext(ext), ".")) {
		return errors.Errorf("'%s' extension is not supported. Use one of [%s]", ext, strings.Join(SupportedExtensions(), ", "))
	}
	viper.SetConfigType(ext[1:])
	viper.AddConfigPath(path.Dir(configFile))

	if err := viper.ReadInConfig(); err != nil {
		if !defaultConfigFile {
			return errors.WrapWithDetails(err, "failed reading config file", "path", configFile)
		}

		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return errors.WrapWithDetails(err, "failed reading config file", "path", configFile)
		}
	}

	return nil
}

// SupportedExtensions returns a slice of all supported config format (as file extensions).
func SupportedExtensions() []string {
	return viper.SupportedExts
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

// SyncFullyQualifiedFlag ensures that if configuration provides a value for a given cmd.flag it will be set back to the flag itself,
// but only if the flag was not set through CLI.
//
// This way we can make flags required but still have their values provided by the configuration source.
func SyncFullyQualifiedFlag(cmd *cobra.Command, flagName string) error {
	value := viper.GetString(cmd.Name() + "." + flagName)
	if value != "" && !cmd.Flag(flagName).Changed {
		err := cmd.Flags().Set(flagName, value)

		return errors.Wrapf(err, "failed setting flag %s with value %v", flagName, value)
	}
	value = viper.GetString(flagName)
	if value != "" && !cmd.Flag(flagName).Changed {
		err := cmd.Flags().Set(flagName, value)

		return errors.Wrapf(err, "failed setting flag %s with value %v", flagName, value)
	}

	return nil
}

// SyncFullyQualifiedFlags ensures that if configuration provide a value for any of defined flags it will be set
// back to the flag itself.
//
// This function iterates over all flags defined for cobra.Command and accumulates errors if they occur while
// calling SyncFullyQualifiedFlag for every flag.
func SyncFullyQualifiedFlags(cmd *cobra.Command) error {
	var errs []error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		syncFlagErr := SyncFullyQualifiedFlag(cmd, flag.Name)
		errs = append(errs, syncFlagErr)
	})

	return errors.Wrap(errors.Combine(errs...), "failed to sync flags")
}

// BindFullyQualifiedFlag ensures that each flag used in commands is bound to a key using fully qualified name
// which has a following form:
//
// 		commandName.flagName
//
// This lets us  keep structure of yaml file:
//
//	commandName:
//		flagName: value
func BindFullyQualifiedFlag(cmd *cobra.Command) func(flag *pflag.Flag) {
	return func(flag *pflag.Flag) {
		_ = viper.BindPFlag(cmd.Name()+"."+flag.Name, flag)
		_ = viper.BindPFlag(flag.Name, flag)
	}
}
