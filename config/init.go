package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

func InitializeConfigs(force bool) error {
	if force {
		err := os.RemoveAll(configDirectory)
		if err != nil {
			return err
		}
	}

	err := os.Mkdir(configDirectory, os.ModeDir|0o700)
	if err != nil {
		return err
	}

	config := DefaultConfig()

	configFile, err := os.Create(filepath.Join(configDirectory, "config.json"))
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"ssh-keygen",
		"-N", "",
		"-f", filepath.Join(configDirectory, "id_rsa"),
		"-t", "rsa", "-b", "4096")
	err = cmd.Run()
	if err != nil {
		return err
	}

	authKeyPath := filepath.Join(configDirectory, "authorized_keys")
	authKeyFile, err := os.Create(authKeyPath)
	if err != nil {
		return err
	}
	authKeyFile.Close()

	return nil
}
