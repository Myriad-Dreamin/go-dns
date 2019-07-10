package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	defaultServerConfig = ServerConfig{
		UDPRange:          500,
		UDPBufferSize:     520,
		TCPRange:          5,
		TCPBUfferSize:     52000,
		ServerNetworkType: "udp4",
		RemoteServerAddr:  "223.5.5.5",
	}
	defaultHostsConfig = HostsConfig{
		RelativePath: true,
		HostsPath:    "./hosts",
	}
	defaultConfig = &Configuration{
		defaultServerConfig,
		defaultHostsConfig,
	}
	cfg         *Configuration
	cfgContext  string = "config.toml"
	cfgLock     sync.RWMutex
	parseConfig sync.Once
)

type Configuration struct {
	ServerConfig ServerConfig `toml:"server"`
	HostsConfig  HostsConfig  `toml:"hosts"`
}

type ServerConfig struct {
	UDPRange          uint16 `toml:"udp_range"`
	UDPBufferSize     uint16 `toml:"udp_buffer_size"`
	TCPRange          uint16 `toml:"tcp_range"`
	TCPBUfferSize     uint16 `toml:"tcp_buffer_size"`
	ServerNetworkType string `toml:"server_network_type"`
	RemoteServerAddr  string `toml:"default_remote_server_address"`
}

type HostsConfig struct {
	RelativePath bool   `toml:"relative_path"`
	HostsPath    string `toml:"hosts_path"`
}

func Config() *Configuration {
	parseConfig.Do(ReloadConfiguration)
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return cfg
}

func ResetPath(path string) {
	cfgContext = path
	ReloadConfiguration()
}

func ReloadConfiguration() {
	filePath, err := filepath.Abs(cfgContext)
	if err != nil {
		panic(err)
	}
	fmt.Printf("reseting: %s\n", filePath)
	config := new(Configuration)
	*config = *defaultConfig
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		panic(err)
	}
	cfgLock.Lock()
	defer cfgLock.Unlock()
	cfg = config
}

// func init() {
// 	s := make(chan os.Signal, 1)
// 	signal.Notify(s, syscall.SIGUSR1)
// 	go func() {
// 		for {
// 			<-s
// 			ReloadConfiguration()
// 			fmt.Printf("reloading\n")
// 		}
// 	}()
// }
