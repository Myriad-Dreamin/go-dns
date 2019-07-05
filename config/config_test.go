package config

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	fmt.Println(Config().ServerConfig.TCPBUfferSize, Config().ServerConfig.RemoteServerAddr)
}
