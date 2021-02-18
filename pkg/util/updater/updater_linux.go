//+build linux

package updater

import (
	"fmt"
	"path/filepath"

	"github.com/skycoin/skywire/pkg/util/osutil"
)

const (
	packageInstallationName    = "skywire"
	packageInstalledBinaryPath = "/opt/skywire/" + visorBinary
)

// InstalledViaPackageInstaller checks if the visor is installed via package installer.
func (u *Updater) InstalledViaPackageInstaller() (bool, error) {
	distro, err := osutil.DetectLinuxDistro()
	if err != nil {
		return false, fmt.Errorf("failed to detect distro")
	}

	u.log.Infof("DISTRO: %v", distro)

	if !osutil.PackageUpdateSupported(distro) {
		return false, nil
	}

	isInstalled, err := osutil.IsPackageInstalled(distro, packageInstallationName)
	if err != nil {
		return false, fmt.Errorf("failed to check if package is installed: %w", err)
	}

	if !isInstalled {
		return false, nil
	}

	binaryPath := filepath.Join(filepath.Dir(u.restartCtx.CmdPath()), visorBinary)

	return binaryPath == packageInstalledBinaryPath, nil
}

func (u *Updater) updateWithPackage() (bool, error) {
	uid, err := osutil.GainRoot()
	defer osutil.ReleaseRoot(uid)

	cmd := "apt-get update && apt-get install --only-upgrade skywire"
	if err := osutil.Run("sh", "-c", cmd); err != nil {
		return false, err
	}

	return true, nil
}
