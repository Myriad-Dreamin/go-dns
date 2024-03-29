package config

import (
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	defaultServerConfig = serverConfig{
		UDPRange:             500,
		UDPBufferSize:        512,
		TCPRange:             10,
		TCPBUfferSize:        65536,
		ServerNetworkType:    "udp4",
		RemoteServerAddr:     "114.114.114.114",
		LocalServerAddr:      "127.0.0.1:53",
		TCPServerTimeout:     5,
		TCPServerTimeoutUnit: "s",
	}
	defaultHostsConfig = hostsConfig{
		RelativePath: true,
		HostsPath:    "./hosts",
	}
	defaultRedisConfig = redisConfig{
		RedisServer:   "127.0.0.1:6379",
		RedisPassword: "",
	}
	defaultConfig = &Configuration{
		defaultServerConfig,
		defaultHostsConfig,
		defaultRedisConfig,
	}

	cfg         *Configuration
	cfgContext  string = "./config.toml"
	cfgLock     sync.RWMutex
	parseConfig sync.Once
)

type Configuration struct {
	ServerConfig serverConfig `toml:"server"`
	HostsConfig  hostsConfig  `toml:"hosts"`
	RedisConfig  redisConfig  `toml:redis`
}

type serverConfig struct {
	UDPRange      int64 `toml:"udp_range"`
	UDPBufferSize int64 `toml:"udp_buffer_size"`

	TCPRange      uint16 `toml:"tcp_range"`
	TCPBUfferSize int64  `toml:"tcp_buffer_size"`

	ServerNetworkType string `toml:"server_network_type"`
	RemoteServerAddr  string `toml:"default_remote_server_address"`
	LocalServerAddr   string `toml:"default_local_server_address"`

	TCPServerTimeout     int64  `toml:"tcp_server_timeout"`
	TCPServerTimeoutUnit string `toml:"tcp_server_timeout_unit"`
}

type hostsConfig struct {
	RelativePath bool   `toml:"relative_path"`
	HostsPath    string `toml:"hosts_path"`
}

type redisConfig struct {
	RedisServer   string `toml:server`
	RedisPassword string `toml:password`
}

func Config() *Configuration {
	parseConfig.Do(func() { ReloadConfiguration() })
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return cfg
}

func ResetPath(path string) {
	cfgContext = path
	ReloadConfiguration()
}

func ReloadConfiguration() error {
	filePath, err := filepath.Abs(cfgContext)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("reseting: %s\n", filePath)
	config := new(Configuration)
	*config = *defaultConfig
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		return err
	}
	cfgLock.Lock()
	defer cfgLock.Unlock()
	cfg = config
	return nil
}

func DefaultHost() string {
	return defaultServerConfig.RemoteServerAddr
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
