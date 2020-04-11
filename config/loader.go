package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func LoadConfig() (config Config, err error) {
	configPath := filepath.Join(configDirectory, "config.json")
	configFile, err := os.Open(configPath)
	if err != nil {
		err = fmt.Errorf("Failed to load config file: %s", err)
		return
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		err = fmt.Errorf("Could not parse config file: %s", err)
		return
	}

	config.SSHConfig, err = makeSSHServerConfig()
	if err != nil {
		return
	}

	return
}

func makeSSHServerConfig() (*ssh.ServerConfig, error) {
	authKeyPath := filepath.Join(configDirectory, "authorized_keys")
	authKeyBytes, err := ioutil.ReadFile(authKeyPath)
	if err != nil {
		return nil, err
	}

	authKeys := map[string]bool{}
	for len(authKeyBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authKeyBytes)
		if err != nil {
			return nil, err
		}

		authKeys[string(pubKey.Marshal())] = true
		authKeyBytes = rest
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if authKeys[string(key.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(key),
					},
				}, nil
			}
			return nil, fmt.Errorf("Unknown key provided")
		},
	}

	privateBytes, err := ioutil.ReadFile(filepath.Join(configDirectory, "id_rsa"))
	if err != nil {
		log.Fatalf("Could not load server private key: %s", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Could not parse server private key: %s", err)
	}

	config.AddHostKey(private)

	return config, nil
}
