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

	config.SSHConfig = LoadSSHServerConfig()
	return
}

func LoadSSHServerConfig() *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "foo" && string(pass) == "bar" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		// NoClientAuth: true,
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

	return config
}
