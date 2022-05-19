package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/spiffe/spire/cmd/spire-agent/cli"
	"github.com/spiffe/spire/pkg/agent"
	"github.com/spiffe/spire/pkg/agent/catalog"
	common_cli "github.com/spiffe/spire/pkg/common/cli"
	"github.com/spiffe/spire/pkg/common/log"
	"github.com/spiffe/spire/pkg/common/util"
)

const (
	rekorServer    = "https://log.testifyse.io"
	trust_domain   = "dev.testifysec.com"
	server_address = "10.28.1.5"
	server_port    = 8081

	socket_path        = "/run/spire/sockets/agent.sock"
	insecure_bootstrap = true
)

func main() {

	new(cli.CLI).Run(os.Args)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startSpire()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startWitness()
	}()

	wg.Wait()

}

func startSpire() {
	for {
		c, err := spireConfig(socket_path, server_address, server_port, trust_domain)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			time.Sleep(time.Second * 5)
			continue
		}

		ctx := context.Background()

		err = agent.New(&c).Run(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			time.Sleep(time.Second * 5)
			continue
		}

	}
}

func startWitness() {

}

func spireConfig(bindAddr string, serverAddr string, serverPort int, trustDomain string) (agent.Config, error) {
	c := agent.Config{}

	serverHostPort := net.JoinHostPort(serverAddr, strconv.Itoa(serverPort))
	c.ServerAddress = fmt.Sprintf("dns:///%s", serverHostPort)

	log, err := log.NewLogger()
	if err != nil {
		return c, err
	}

	c.Log = log

	bind, err := util.GetUnixAddrWithAbsPath(bindAddr)
	if err != nil {
		return c, err
	}

	td, err := common_cli.ParseTrustDomain(trustDomain, nil)
	if err != nil {
		return c, err
	}

	conf := catalog.Config{
		Log: log,
		PluginConfig: catalog.HCLPluginConfigMap{
			"KeyManager": {
				"memory": {},
			},
			"NodeAttestor": {
				"gcp_iit": {},
			},
			"WorkloadAttestor": {
				"unix": {},
			},
		},
	}

	c.TrustDomain = td
	c.BindAddress = bind
	c.InsecureBootstrap = true
	c.DataDir = "/tmp/spire-agent"
	c.PluginConfigs = conf.PluginConfig
	return c, nil
}
