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
	Users     []User
	SSHConfig *ssh.ServerConfig `json:"-"`
}

type User struct {
	Username string
	Key      ssh.PublicKey
}

func (user User) KeyString() string {
	return string(user.Key.Marshal())
}
