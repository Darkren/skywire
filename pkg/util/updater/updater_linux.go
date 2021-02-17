//+build linux

package updater

func (u *Updater) InstalledViaPackageInstaller() (bool, error) {
	return false, nil
}
