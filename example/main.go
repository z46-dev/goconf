package main

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/z46-dev/goconf"
)

type Configuration struct {
	HostIPv4 string `toml:"host_ipv4" validate:"required,ipv4"`
	HostIPv6 string `toml:"host_ipv6" validate:"required,ipv6"`
	HostPort int    `toml:"host_port" validate:"required,gt=0,lt=65536"`
	SSL      struct {
		Enabled  bool   `toml:"enabled"`
		CertPath string `toml:"cert_path" validate:"required_if=SSL.Enabled true"`
		KeyPath  string `toml:"key_path" validate:"required_if=SSL.Enabled true"`
	}
}

func resolveToWhereThisProgramIs() (filename string) {
	var ok bool
	if _, filename, _, ok = runtime.Caller(0); !ok {
		panic(fmt.Sprintf("failed to resolve caller info: %v", ok))
	}

	filename = strings.TrimSuffix(filename, "main.go")
	return
}

func main() {
	var (
		configFilePath string        = path.Join(resolveToWhereThisProgramIs(), "config.toml")
		start          time.Time     = time.Now()
		config         Configuration = goconf.MustLoadConfig[Configuration](configFilePath)
	)

	fmt.Printf("filePath=%s\ntime=%s\nConfiguration%+v\n", configFilePath, time.Since(start), config)
}
