package cmd

import (
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

		_ = viper.BindPFlag("addr", cmd.Flags().Lookup("addr"))
		_ = viper.BindPFlag("trusted-origins", cmd.Flags().Lookup("trusted-origins"))
		_ = viper.BindPFlag("token", cmd.Flags().Lookup("token"))

		addr := viper.GetString("addr")
		trustedOrigins := viper.GetStringSlice("trusted-origins")
		token := viper.GetString("token")

		slog.Info("starting server", "addr", addr, "trusted_origins", trustedOrigins)

		store := wsserver.GetStore()

		return server.Run(addr, store, trustedOrigins, token)
	},
}

func init() {
	serveCmd.Flags().String("addr", ":8080", "address to bind the server to")
	serveCmd.Flags().StringSlice("trusted-origins", []string{}, "comma-separated list of trusted CORS origins")
	serveCmd.Flags().String("token", "", "authentication token (generates one if not provided)")

	_ = viper.BindEnv("addr", "RLGL_SERVER_ADDR")
	_ = viper.BindEnv("trusted-origins", "RLGL_TRUSTED_ORIGINS")
	_ = viper.BindEnv("token", "RLGL_TOKEN")

	RootCmd.AddCommand(serveCmd)
}
