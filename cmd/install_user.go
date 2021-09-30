package cmd

import (
	"fmt"
	"path/filepath"

	"os"
	"path"

	"github.com/blunghamer/creds"
	"github.com/blunghamer/systemd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installUserCmd)
}

var installUserCmd = &cobra.Command{
	Use:   "installuser",
	Short: fmt.Sprintf("install user service of %v run as normal user", toolname),
	RunE:  runInstallUser,
}

const workingDirectoryPermission = 0755
const userconfigdir = ".config/" + toolname

func runInstallUser(_ *cobra.Command, _ []string) error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := ensureWorkingDir(path.Join(home, userconfigdir)); err != nil {
		return err
	}

	absConfig := path.Join(home, userconfigdir, toolname+".yaml")

	src, err := creds.FS.Open(filepath.Join("static", toolname+".yaml"))
	if err != nil {
		return err
	}
	defer src.Close()

	if err := cpf(src, absConfig); err != nil {
		log.Error().Str("toolname", toolname).Str("configpath", absConfig).Err(err).Msg("Unable to copy config")
		return err
	}

	err = systemd.Enable(toolname, true)
	if err != nil {
		return err
	}

	err = systemd.Start(toolname, true)
	if err != nil {
		return err
	}

	msg := `Successfully installed, check status with:
	systemctl --user status %v
	journalctl --user --user-unit %v --lines 100 -f`

	fmt.Println(fmt.Printf(msg, toolname, toolname))

	return nil
}

func ensureWorkingDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, workingDirectoryPermission)
		if err != nil {
			return err
		}
	}
	return nil
}
