package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func LoadConfig() (config Config, err error) {
	config.Users, err = loadUsers()
	if err != nil {
		return
	}

	config.SSHConfig, err = makeSSHServerConfig(config.Users)
	if err != nil {
		return
	}

	return
}

func loadUsers() (map[string]User, error) {
	authKeyPath := filepath.Join(configDirectory, "authorized_keys")
	authKeyBytes, err := ioutil.ReadFile(authKeyPath)
	if err != nil {
		return nil, err
	}

	users := make(map[string]User)
	for len(authKeyBytes) > 0 {
		pubKey, comment, _, rest, err := ssh.ParseAuthorizedKey(authKeyBytes)
		if err != nil {
			return nil, err
		}

		user, ok := users[comment]
		if !ok {
			users[comment] = User{
				Username: comment,
			}
			user = users[comment]
		}

		users[comment] = User{
			Username: user.Username,
			Keys:     append(user.Keys, pubKey),
		}
		authKeyBytes = rest
	}

	return users, nil
}

func makeSSHServerConfig(users Users) (*ssh.ServerConfig, error) {
	sshConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			user, _, ok := users.LookupUser(c.User())
			if ok && user.CheckKey(key) {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(key),
					},
				}, nil
			}
			return nil, fmt.Errorf("Invalid credentials")
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

	sshConfig.AddHostKey(private)

	return sshConfig, nil
}
