package workflow

import (
	"os/exec"

	plist "howett.net/plist"
)

func WriteMetadata(field, text, path string) error {
	data, err := plist.Marshal(text, plist.XMLFormat)
	if err != nil {
		return err
	}
	cmd := exec.Command("/usr/bin/xattr", "-w", "com.apple.metadata:"+field, string(data), path)
	return cmd.Run()
}

func ReadMetadata(field, path string) (string, error) {
	cmd := exec.Command("/usr/bin/xattr", "-p", "com.apple.metadata:"+field, path)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var result string
	if _, err := plist.Unmarshal(out, &result); err != nil {
		return "", err
	}
	return result, nil
}
