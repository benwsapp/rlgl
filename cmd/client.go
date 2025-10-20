package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/benwsapp/rlgl/pkg/wsclient"
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
		viper.SetConfigName("rlgl")
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

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run the client that pushes config to the server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

		serverURL, err := cmd.Flags().GetString("server")
		if err != nil {
			return fmt.Errorf("failed to get server flag: %w", err)
		}

		clientID, err := cmd.Flags().GetString("client-id")
		if err != nil {
			return fmt.Errorf("failed to get client-id flag: %w", err)
		}

		interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			return fmt.Errorf("failed to get interval flag: %w", err)
		}

		once, err := cmd.Flags().GetBool("once")
		if err != nil {
			return fmt.Errorf("failed to get once flag: %w", err)
		}

		configPath, err := InitConfig(cmd)
		if err != nil {
			return err
		}

		slog.Info(
			"starting websocket client",
			"server", serverURL,
			"client_id", clientID,
			"config", configPath,
			"interval", interval,
			"once", once,
		)

		if once {
			return wsclient.RunOnce(serverURL, configPath, clientID)
		}

		return wsclient.Run(serverURL, configPath, clientID, interval)
	},
}

const defaultInterval = 30 * time.Second

func init() {
	clientCmd.Flags().String("server", "ws://localhost:8080/ws", "WebSocket server URL")
	clientCmd.Flags().String("client-id", "", "unique client identifier (required)")
	clientCmd.Flags().Duration("interval", defaultInterval, "interval between config pushes")
	clientCmd.Flags().Bool("once", false, "push config once and exit")
	clientCmd.Flags().String("config", "", "path to site configuration file (defaults to rlgl.yaml)")

	_ = clientCmd.MarkFlagRequired("client-id")

	_ = viper.BindEnv("server", "RLGL_REMOTE_HOST")
	_ = viper.BindEnv("client-id", "RLGL_CLIENT_ID")
	_ = viper.BindEnv("interval", "RLGL_CLIENT_INTERVAL")
	_ = viper.BindEnv("once", "RLGL_CLIENT_ONCE")

	_ = viper.BindPFlag("server", clientCmd.Flags().Lookup("server"))
	_ = viper.BindPFlag("client-id", clientCmd.Flags().Lookup("client-id"))
	_ = viper.BindPFlag("interval", clientCmd.Flags().Lookup("interval"))
	_ = viper.BindPFlag("once", clientCmd.Flags().Lookup("once"))

	RootCmd.AddCommand(clientCmd)
}
