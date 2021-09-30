package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"net/http"

	"github.com/blunghamer/creds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: fmt.Sprintf("get key from %v secrets database", toolname),
	Run:   runGet,
}

func runGet(_ *cobra.Command, args []string) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if len(args) != 1 {
		log.Error().Msg("please provide keyname as argument")
		return
	}

	cnf, err := creds.ReadConfig()
	if err != nil {
		log.Error().Err(err).Msg("Config error")
		return
	}

	r, err := http.Get("http://" + cnf.Listenaddress + "/secret/" + args[0])
	if err != nil {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	if r.StatusCode > 299 {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	byts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Unseal request error")
		return
	}

	fmt.Print(string(byts))
}
