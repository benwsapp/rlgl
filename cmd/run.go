package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/benwsapp/rlgl/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitConfig(cmd *cobra.Command) (string, error) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", fmt.Errorf("failed to get config flag: %w", err)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("site")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("config")
	}

	err = viper.ReadInConfig()
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	return viper.ConfigFileUsed(), nil
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the rlgl web server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return fmt.Errorf("failed to get addr flag: %w", err)
		}

		trustedOrigins, err := cmd.Flags().GetStringSlice("trusted-origins")
		if err != nil {
			return fmt.Errorf("failed to get trusted-origins flag: %w", err)
		}

		configPath, err := InitConfig(cmd)
		if err != nil {
			return err
		}

		slog.Info("starting server", "addr", addr, "config", configPath, "trusted_origins", trustedOrigins)

		return server.Run(addr, configPath, trustedOrigins)
	},
}

func init() {
	runCmd.Flags().String("addr", ":8080", "address to bind the server to")
	runCmd.Flags().String("config", "", "path to site configuration file (defaults to site.yaml)")
	runCmd.Flags().StringSlice("trusted-origins", []string{}, "comma-separated list of trusted CORS origins")

	_ = viper.BindEnv("addr", "RLGL_ADDR")
	_ = viper.BindEnv("config", "RLGL_CONFIG")
	_ = viper.BindEnv("trusted-origins", "RLGL_TRUSTED_ORIGINS")

	_ = viper.BindPFlag("addr", runCmd.Flags().Lookup("addr"))
	_ = viper.BindPFlag("config", runCmd.Flags().Lookup("config"))
	_ = viper.BindPFlag("trusted-origins", runCmd.Flags().Lookup("trusted-origins"))

	RootCmd.AddCommand(runCmd)
}
