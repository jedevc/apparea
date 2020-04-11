package config

import (
	"log"
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
