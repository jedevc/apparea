package tunnel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

var configDirectory string

func init() {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	configDirectory = filepath.Join(user.HomeDir, ".apparea")
}

type Config struct {
	Hostname string `json:"hostname"`

	SSHConfig *ssh.ServerConfig `json:"-"`
}

func DefaultConfig() Config {
	return Config{
		Hostname: "apparea.dev",
	}
}

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

	config.SSHConfig, err = LoadServerConfig()
	if err != nil {
		return
	}

	return
}

func LoadServerConfig() (*ssh.ServerConfig, error) {
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
