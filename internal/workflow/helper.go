package workflow

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

const helperBinaryName = "chatgpt-helper"

func HelperBinaryPath(dataDir string) string {
	return filepath.Join(dataDir, helperBinaryName)
}

func EnsureHelperBinary(dataDir string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if filepath.Base(exe) == helperBinaryName {
		return nil
	}

	dest := HelperBinaryPath(dataDir)

	srcInfo, err := os.Stat(exe)
	if err != nil {
		return err
	}

	if destInfo, err := os.Stat(dest); err == nil {
		if destInfo.Size() == srcInfo.Size() && !srcInfo.ModTime().After(destInfo.ModTime().Add(time.Second)) {
			return nil
		}
	}

	if err := copyFile(exe, dest); err != nil {
		return err
	}
	return os.Chmod(dest, 0o755)
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, dest)
}

func HelperExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if filepath.Base(exe) == helperBinaryName {
		return exe
	}
	return ""
}
