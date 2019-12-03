package main

import (
	"github.com/jedevc/AppArea/tunnel"
)

func main() {
	config := tunnel.LoadSSHServerConfig()
	sessions := tunnel.Run("0.0.0.0:2200", config)

	for _ = range sessions {
	}

	select {}
}
