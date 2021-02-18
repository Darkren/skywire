//+build darwin

package updater

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/skycoin/skywire/pkg/util/osutil"
)

const (
	packageInstallationName    = "com.skycoin.skywire.visor"
	packageInstalledBinaryPath = "/Applications/Skywire.app/Contents/" + visorBinary
)

// InstalledViaPackageInstaller checks if the visor is installed via package installer.
func (u *Updater) InstalledViaPackageInstaller() (bool, error) {
	cmd := "/usr/sbin/pkgutil --pkgs=" + packageInstallationName
	outputBytes, err := osutil.RunWithResult("sh", "-c", cmd)
	if err != nil {
		return false, fmt.Errorf("failed to execute command %s: %w", cmd, err)
	}

	outputBytes = bytes.TrimSpace(outputBytes)

	if string(outputBytes) != packageInstallationName {
		return false, nil
	}

	binaryPath := filepath.Join(filepath.Dir(u.restartCtx.CmdPath()), visorBinary)

	return binaryPath == packageInstalledBinaryPath, nil
}

func (u *Updater) updateWithPackage() (bool, error) {
	// we don't update it with package automatically,
	// so just return false
	return false, nil
}
