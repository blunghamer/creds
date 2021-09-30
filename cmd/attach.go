package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"net/http"

	"github.com/blunghamer/creds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var kpSource string

func init() {
	rootCmd.AddCommand(attachCmd)
	attachCmd.Flags().StringVarP(&kpSource, "keyfile", "k", "/vagrant/dev.kdbx", "path to kdbx file")
}

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: fmt.Sprintf("attach %v secrets database", toolname),
	Run:   runAttach,
}

func runAttach(_ *cobra.Command, _ []string) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cnf, err := creds.ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("Config error")
		return
	}

	pass, err := creds.ReadPasswordBytes()
	if err != nil {
		log.Error().Err(err).Msg("Password read error")
		return
	}

	byts, err := json.Marshal(creds.AttachMsg{Password: string(pass), SourceFile: kpSource})
	if err != nil {
		log.Error().Err(err).Msg("Password read error")
		return
	}

	r, err := http.Post("http://"+cnf.Listenaddress+"/attach", "application/json", bytes.NewReader(byts))
	if err != nil {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	if r.StatusCode > 299 {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	log.Info().Int("statuscode", r.StatusCode).Str("file", kpSource).Msg("successfully attached ")
}
