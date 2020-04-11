package config

import (
	"bytes"
	"log"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"
)

var IsValidUsername = regexp.MustCompile(`^([a-zA-Z0-9]+)(\.[a-zA-Z0-9]+)*$`).MatchString

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

func (config Config) LookupUser(username string) (User, []string, bool) {
	if !IsValidUsername(username) {
		return User{}, nil, false
	}

	userParts := strings.Split(username, ".")
	if len(userParts) == 0 {
		return User{}, nil, false
	}

	user, ok := config.Users[userParts[0]]
	if !ok {
		return User{}, nil, false
	}

	userParts = userParts[1:]
	for i := 0; i < len(userParts)/2; i++ {
		j := len(userParts) - 1
		userParts[i], userParts[j] = userParts[j], userParts[i]
	}
	return user, userParts, true
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
