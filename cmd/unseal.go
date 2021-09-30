package cmd

import (
	"net/http"
	"os"

	"github.com/blunghamer/creds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(unsealCmd)
}

var unsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "unseal kdbx",
	Run:   runUnseal,
}

func runUnseal(_ *cobra.Command, args []string) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cnf, err := creds.ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("Config error")
		return
	}

	r, err := http.Get("http://" + cnf.Listenaddress + "/unseal")
	if err != nil {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	if r.StatusCode > 299 {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	log.Info().Msg("Unseal success")
}
