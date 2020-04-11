package config

import (
	"bytes"
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
	Users     map[string]User
	SSHConfig *ssh.ServerConfig `json:"-"`
}

type User struct {
	Username string
	Keys     []ssh.PublicKey
}

func (user User) CheckKey(target ssh.PublicKey) bool {
	targetBytes := target.Marshal()
	for _, key := range user.Keys {
		keyBytes := key.Marshal()
		if bytes.Equal(targetBytes, keyBytes) {
			return true
		}
	}

	return false
}
