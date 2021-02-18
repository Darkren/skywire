package osutil

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
)

// LinuxDistro is a Linux distribution type.
type LinuxDistro int

const (
	// LinuxDistroUnknown is an unidentified distribution.
	LinuxDistroUnknown LinuxDistro = iota
	// LinuxDistroDebian stands for Debian-based distros.
	LinuxDistroDebian
)

type packageInstalledChecker func(pkgName string) (bool, error)

var (
	packageUpdateSupportedDistros = map[LinuxDistro]packageInstalledChecker{
		LinuxDistroDebian: isPackageInstalledDebian,
	}

	errFuncNotSupported = errors.New("func not supported for this OS")
)

// DetectLinuxDistro detects current Linux distribution.
func DetectLinuxDistro() (LinuxDistro, error) {
	if runtime.GOOS != "linux" {
		fmt.Println("GOOS NOT LINUX")
		return LinuxDistroUnknown, errFuncNotSupported
	}

	outputBytes, err := RunWithResult("apt", "-v")
	if err != nil {
		return LinuxDistroUnknown, err
	}

	fmt.Printf("DISTRO OUTPUT: %s\n", string(outputBytes))

	outputBytes = bytes.TrimSpace(outputBytes)
	if bytes.Contains(outputBytes, []byte("command not found")) {
		fmt.Println("RETURNING UNKNOWN")
		return LinuxDistroUnknown, nil
	}

	fmt.Println("RETURNING DEBIAN")
	return LinuxDistroDebian, nil
}

// PackageUpdateSupported checks if package update is supported for distro `d`.
func PackageUpdateSupported(d LinuxDistro) bool {
	_, ok := packageUpdateSupportedDistros[d]
	return ok
}

// IsPackageInstalled checks if package `pkgName` is installed with package for distro `d`.
func IsPackageInstalled(d LinuxDistro, pkgName string) (bool, error) {
	if !PackageUpdateSupported(d) {
		return false, fmt.Errorf("package update is not supported")
	}

	installationChecker := packageUpdateSupportedDistros[d]
	return installationChecker(pkgName)
}

func isPackageInstalledUnknown(_ string) (bool, error) {
	return false, errFuncNotSupported
}

func isPackageInstalledDebian(pkgName string) (bool, error) {
	cmd := "dpkg --get-selections | grep -v deinstall | grep skywire | awk '{print $1}'"
	outputBytes, err := RunWithResult("sh", "-c", cmd)
	if err != nil {
		return false, fmt.Errorf("failed to execute command %s: %w", cmd, err)
	}

	outputBytes = bytes.TrimSpace(outputBytes)

	fmt.Printf("OUTPUT: %s\n", string(outputBytes))

	return string(outputBytes) == pkgName, nil
}
