package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/benwsapp/rlgl/pkg/server"
	"github.com/benwsapp/rlgl/pkg/wsserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the server that receives client configs",
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

		slog.Info("starting server", "addr", addr, "trusted_origins", trustedOrigins)

		store := wsserver.GetStore()

		return server.Run(addr, store, trustedOrigins)
	},
}

func init() {
	serveCmd.Flags().String("addr", ":8080", "address to bind the server to")
	serveCmd.Flags().StringSlice("trusted-origins", []string{}, "comma-separated list of trusted CORS origins")

	_ = viper.BindEnv("addr", "RLGL_SERVER_ADDR")
	_ = viper.BindEnv("trusted-origins", "RLGL_TRUSTED_ORIGINS")

	_ = viper.BindPFlag("addr", serveCmd.Flags().Lookup("addr"))
	_ = viper.BindPFlag("trusted-origins", serveCmd.Flags().Lookup("trusted-origins"))

	RootCmd.AddCommand(serveCmd)
}
