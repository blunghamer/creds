package creds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/tobischo/gokeepasslib/v3"
	"golang.org/x/crypto/ssh/terminal"
)

func PrintGroup(group gokeepasslib.Group) {
	for _, ent := range group.Entries {
		log.Printf("%v %v %v", ent.GetTitle(), ent.GetPassword(), ent.Tags)
		for _, val := range ent.Values {
			log.Printf(" %v = %v", val.Key, val.Value)
		}
	}
	for _, grp := range group.Groups {
		PrintGroup(grp)
	}
}

func (p *Cred) DebugPrint() {
	for _, group := range p.db.Content.Root.Groups {
		PrintGroup(group)
	}
}

type Cred struct {
	filename string
	db       *gokeepasslib.Database
	sealed   bool
}

func NewCred(conf *Config) *Cred {
	p := &Cred{filename: conf.Keypassfile, db: nil, sealed: true}
	return p
}

func WalkGroup(group gokeepasslib.Group, cb WalkCallback) {
	for _, ent := range group.Entries {
		cb(ent)
	}
	for _, grp := range group.Groups {
		WalkGroup(grp, cb)
	}
}

type WalkCallback func(entry gokeepasslib.Entry)

func (p *Cred) Walk(cb WalkCallback) {
	for _, group := range p.db.Content.Root.Groups {
		WalkGroup(group, cb)
	}
}

func (p *Cred) Read(filename, pass string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Msg("Unable to read keepass file")
		return err
	}

	p.db = gokeepasslib.NewDatabase()
	p.db.Credentials = gokeepasslib.NewPasswordCredentials(pass)
	err = gokeepasslib.NewDecoder(file).Decode(p.db)
	if err != nil {
		log.Error().Err(err).Msg("Unable to read keepass")
		return err
	}

	return nil
}

func (p *Cred) Seal(w http.ResponseWriter, r *http.Request) {
	if p.sealed {
		log.Warn().Msg("Cannot double seal")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.db == nil {
		log.Warn().Msg("Cannot seal, unseal db with creds unseal on command line first")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := p.db.LockProtectedEntries()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	p.sealed = true
}

type AttachMsg struct {
	Password   string
	SourceFile string
}

func (p *Cred) Attach(w http.ResponseWriter, r *http.Request) {

	/*
		if p.db != nil {
			log.Warn().Msg("Cannot double attach keepass")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("cannot attach multiple"))
			return
		}
	*/

	byts, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	at := &AttachMsg{}
	json.Unmarshal(byts, &at)

	err = p.Read(at.SourceFile, at.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (p *Cred) Unseal(w http.ResponseWriter, r *http.Request) {
	if !p.sealed {
		log.Warn().Msg("Cannot double unseal")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unable to unseal again"))
		return
	}

	if p.db == nil {
		log.Warn().Msg("Cannot unseal, unseal db with creds unseal on command line first")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to unseal locked db"))
		return
	}

	err := p.db.UnlockProtectedEntries()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	p.sealed = false
}

func (p *Cred) SecretHandler(w http.ResponseWriter, r *http.Request) {

	if p.db == nil {
		log.Warn().Msg("Cannot get secret, unseal db with creds unseal on command line first")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	sec := "sealed"
	p.Walk(func(ent gokeepasslib.Entry) {
		if ent.GetTitle() == vars["key"] {
			sec = ent.GetPassword()
		}
	})

	if sec == "sealed" {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Str("key", vars["key"]).Msg("key not found")
		fmt.Fprint(w, "key not found")
	} else {
		log.Debug().Str("key", vars["key"]).Msg("key found")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sec)
	}
}

func ReadPasswordBytes() ([]byte, error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		log.Error().Err(err).Msg("unable to read password from terminal")
		return []byte(""), err
	}
	return bytePassword, nil
}
