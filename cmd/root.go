package cmd

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/blunghamer/creds"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

/*
func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: w.NewBoolWrapper(true)},
	}
}
func write(filename, pass string) {

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// create root group
	rootGroup := gokeepasslib.NewGroup()
	rootGroup.Name = "root group"

	entry := gokeepasslib.NewEntry()
	entry.Values = append(entry.Values, mkValue("Title", "My GMail password"))
	entry.Values = append(entry.Values, mkValue("UserName", "example@gmail.com"))
	entry.Values = append(entry.Values, mkProtectedValue("Password", "hunter2"))

	rootGroup.Entries = append(rootGroup.Entries, entry)

	// demonstrate creating sub group (we'll leave it empty because we're lazy)
	subGroup := gokeepasslib.NewGroup()
	subGroup.Name = "sub group"

	subEntry := gokeepasslib.NewEntry()
	subEntry.Values = append(subEntry.Values, mkValue("Title", "Another password"))
	subEntry.Values = append(subEntry.Values, mkValue("UserName", "johndough"))
	subEntry.Values = append(subEntry.Values, mkProtectedValue("Password", "123456"))

	subGroup.Entries = append(subGroup.Entries, subEntry)

	rootGroup.Groups = append(rootGroup.Groups, subGroup)

	// now create the database containing the root group
	db := &gokeepasslib.Database{
		Header:      gokeepasslib.NewHeader(),
		Credentials: gokeepasslib.NewPasswordCredentials(pass),
		Content: &gokeepasslib.DBContent{
			Meta: gokeepasslib.NewMetaData(),
			Root: &gokeepasslib.RootData{
				Groups: []gokeepasslib.Group{rootGroup},
			},
		},
	}

	// Lock entries using stream cipher
	db.LockProtectedEntries()

	// and encode it into the file
	keepassEncoder := gokeepasslib.NewEncoder(file)
	if err := keepassEncoder.Encode(db); err != nil {
		panic(err)
	}

	log.Printf("Wrote kdbx file: %s", filename)
}
*/

func serve() {
	conf, err := creds.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to read server config")
	}

	wait := time.Second * 15
	fmt.Println()

	p := creds.NewCred(conf)

	r := mux.NewRouter()
	r.HandleFunc("/attach", p.Attach)
	r.HandleFunc("/seal", p.Seal)
	r.HandleFunc("/unseal", p.Unseal)
	r.HandleFunc("/secret/{key}", p.SecretHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Addr: conf.Listenaddress,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	log.Info().Str("address", conf.Listenaddress).Msg("Starting server")
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info().Msg("shutting down")
	os.Exit(0)
}

const toolname = "creds"

var (
	//keepassFile string
	rootCmd = &cobra.Command{
		Use:   "creds",
		Short: "creds exposes secrets after unsealing via http",
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
	}
)

func init() {
	//rootCmd.PersistentFlags().StringVar(&keepassFile, "kpfile", "/vagrant/dev.kdbx", "keepass file")
	//viper.BindPFlag("kpfile", rootCmd.PersistentFlags().Lookup("kpfile"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
