package main

import (
	"html/template"

	"github.com/stretchr/graceful"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/modules/gateway"
	"github.com/NebulousLabs/Sia/modules/host"
	"github.com/NebulousLabs/Sia/modules/hostdb"
	"github.com/NebulousLabs/Sia/modules/miner"
	"github.com/NebulousLabs/Sia/modules/renter"
	"github.com/NebulousLabs/Sia/modules/transactionpool"
	"github.com/NebulousLabs/Sia/modules/wallet"
	"github.com/NebulousLabs/Sia/network"
)

type DaemonConfig struct {
	// Network Variables
	APIAddr string
	RPCAddr string

	// Host Variables
	HostDir string

	// Miner Variables
	Threads int

	// Renter Variables
	DownloadDir string

	// Wallet Variables
	WalletDir string
}

type daemon struct {
	// Modules. TODO: Implement all modules.
	state   *consensus.State
	tpool   *transactionpool.TransactionPool
	network *network.TCPServer
	wallet  *wallet.Wallet
	miner   *miner.Miner
	host    *host.Host
	hostDB  *hostdb.HostDB
	renter  *renter.Renter
	gateway *gateway.Gateway

	styleDir    string
	downloadDir string

	template *template.Template

	apiServer *graceful.Server
}

func newDaemon(config DaemonConfig) (d *daemon, err error) {
	d = new(daemon)
	d.state = consensus.CreateGenesisState()
	d.network, err = network.NewTCPServer(config.RPCAddr)
	if err != nil {
		return
	}
	d.gateway, err = gateway.New(d.network, d.state)
	if err != nil {
		return
	}
	d.tpool, err = transactionpool.New(d.state, d.gateway)
	if err != nil {
		return
	}
	d.wallet, err = wallet.New(d.state, d.tpool, config.WalletDir)
	if err != nil {
		return
	}
	d.miner, err = miner.New(d.state, d.gateway, d.tpool, d.wallet)
	if err != nil {
		return
	}
	d.host, err = host.New(d.state, d.tpool, d.wallet, config.HostDir)
	if err != nil {
		return
	}
	d.hostDB, err = hostdb.New(d.state)
	if err != nil {
		return
	}
	d.renter, err = renter.New(d.state, d.hostDB, d.wallet)
	if err != nil {
		return
	}

	// register RPC handlers
	// TODO: register all RPCs in a separate function
	err = d.network.RegisterRPC("AcceptBlock", d.gateway.RelayBlock)
	if err != nil {
		return
	}
	err = d.network.RegisterRPC("AcceptTransaction", d.tpool.AcceptTransaction)
	if err != nil {
		return
	}
	err = d.network.RegisterRPC("AddMe", d.gateway.AddMe)
	if err != nil {
		return
	}
	err = d.network.RegisterRPC("SendBlocks", d.gateway.SendBlocks)
	if err != nil {
		return
	}
	err = d.network.RegisterRPC("HostSettings", d.host.Settings)
	if err != nil {
		return
	}
	err = d.network.RegisterRPC("NegotiateContract", d.host.NegotiateContract)
	if err != nil {
		return
	}
	err = d.network.RegisterRPC("RetrieveFile", d.host.RetrieveFile)
	if err != nil {
		return
	}

	return
}
