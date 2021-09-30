package cmd

import (
	"fmt"
	"io"
	"path/filepath"

	"os"
	"path"

	"github.com/blunghamer/creds"
	"github.com/blunghamer/systemd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installSudoCmd)
}

var installSudoCmd = &cobra.Command{
	Use:   "installsudo",
	Short: fmt.Sprintf("move %v binary to bin folder, run as sudo please", toolname),
	RunE:  runInstallSudo,
}

func runInstallSudo(_ *cobra.Command, _ []string) error {

	targetFolder := "/usr/local/bin/"

	unitTargetFolder := "/usr/lib/systemd/user/"
	serviceFile := toolname + ".service"

	exeFile := os.Args[0]
	outfile := path.Join(targetFolder, toolname)
	serviceOut := path.Join(unitTargetFolder, serviceFile)

	if err := cp(exeFile, outfile); err != nil {
		log.Printf("Unable to copy %v from %v to %v: %v", toolname, exeFile, outfile, err)
		return err
	}

	if err := os.Chmod(outfile, 0755); err != nil {
		log.Printf("Unable to chmod binary %v", err)
		return err
	}

	if err := binExists(targetFolder, toolname); err != nil {
		return err
	}

	log.Printf("Successfully installed binary %v to %v", toolname, targetFolder)

	src, err := creds.FS.Open(filepath.Join("static", serviceFile))
	if err != nil {
		return err
	}
	defer src.Close()

	if err := cpf(src, serviceOut); err != nil {
		log.Printf("Unable to copy %v from %v to %v: %v", serviceFile, serviceFile, serviceOut, err)
		return err
	}

	log.Printf("Successfully installed service file %v to %v", serviceFile, serviceOut)

	if err := systemd.DaemonReload(); err != nil {
		log.Printf("Unable to reload systemd daemon%v", err)
		return err
	}

	return nil
}

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}

func cpf(source io.Reader, dest string) error {

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, source)

	return err
}

func cp(source, dest string) error {
	file, err := os.Open(source)
	if err != nil {
		return err

	}
	defer file.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return err
}
