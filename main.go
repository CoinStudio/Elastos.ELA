package main

import (
	"bytes"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/elastos/Elastos.ELA/blockchain"
	cmdcom "github.com/elastos/Elastos.ELA/cmd/common"
	"github.com/elastos/Elastos.ELA/common/config"
	"github.com/elastos/Elastos.ELA/common/log"
	"github.com/elastos/Elastos.ELA/core/types"
	"github.com/elastos/Elastos.ELA/dpos"
	"github.com/elastos/Elastos.ELA/dpos/state"
	"github.com/elastos/Elastos.ELA/dpos/store"
	"github.com/elastos/Elastos.ELA/elanet"
	"github.com/elastos/Elastos.ELA/mempool"
	"github.com/elastos/Elastos.ELA/p2p"
	"github.com/elastos/Elastos.ELA/p2p/msg"
	"github.com/elastos/Elastos.ELA/pow"
	"github.com/elastos/Elastos.ELA/servers"
	"github.com/elastos/Elastos.ELA/servers/httpjsonrpc"
	"github.com/elastos/Elastos.ELA/servers/httpnodeinfo"
	"github.com/elastos/Elastos.ELA/servers/httprestful"
	"github.com/elastos/Elastos.ELA/servers/httpwebsocket"
	"github.com/elastos/Elastos.ELA/utils/elalog"
	"github.com/elastos/Elastos.ELA/utils/signal"
)

var (
	// Build version generated when build program.
	Version string

	// The go source code version at build.
	GoVersion string

	// printStateInterval is the interval to print out peer-to-peer network
	// state.
	printStateInterval = time.Minute
)

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(10)

	log.Infof("Node version: %s", Version)
	log.Info(GoVersion)

	var interrupt = signal.NewInterrupt()

	// fixme remove singleton Ledger
	ledger := blockchain.Ledger{}

	// Initializes the foundation address
	blockchain.FoundationAddress = activeNetParams.Foundation

	var dposStore store.IDposStore
	chainStore, err := blockchain.NewChainStore(dataDir, activeNetParams.GenesisBlock)
	if err != nil {
		printErrorAndExit(err)
	}
	defer chainStore.Close()
	ledger.Store = chainStore // fixme

	dposStore, err = store.NewDposStore(dataDir)
	if err != nil {
		printErrorAndExit(err)
	}
	defer dposStore.Close()

	txMemPool := mempool.NewTxPool()
	blockMemPool := mempool.NewBlockPool(activeNetParams)
	blockMemPool.Store = chainStore

	blockchain.DefaultLedger = &ledger // fixme

	arbiters, err := state.NewArbitrators(activeNetParams, chainStore.GetHeight)
	if err != nil {
		printErrorAndExit(err)
	}
	ledger.Arbitrators = arbiters // fixme

	chain, err := blockchain.New(chainStore, activeNetParams, arbiters.State)
	if err != nil {
		printErrorAndExit(err)
	}
	ledger.Blockchain = chain // fixme
	blockMemPool.Chain = chain

	server, err := elanet.NewServer(dataDir, &elanet.Config{
		Chain:          chain,
		ChainParams:    activeNetParams,
		PermanentPeers: cfg.PermanentPeers,
		TxMemPool:      txMemPool,
		BlockMemPool:   blockMemPool,
	})
	if err != nil {
		printErrorAndExit(err)
	}

	blockMemPool.IsCurrent = server.IsCurrent

	var arbitrator *dpos.Arbitrator
	if config.Parameters.EnableArbiter {
		log.Info("Start the manager")
		pwd, err := cmdcom.GetFlagPassword()
		if err != nil {
			printErrorAndExit(err)
		}
		arbitrator, err = dpos.NewArbitrator(pwd, dpos.Config{
			EnableEventLog:    true,
			EnableEventRecord: false,
			Params:            cfg.ArbiterConfiguration,
			ChainParams:       activeNetParams,
			Arbitrators:       arbiters,
			Store:             dposStore,
			TxMemPool:         txMemPool,
			BlockMemPool:      blockMemPool,
			Broadcast: func(msg p2p.Message) {
				server.BroadcastMessage(msg)
			},
		})
		if err != nil {
			printErrorAndExit(err)
		}
		servers.Arbiter = arbitrator
	}

	servers.Compile = Version
	servers.Config = cfg
	servers.Chain = chain
	servers.Store = chainStore
	servers.TxMemPool = txMemPool
	servers.Server = server
	servers.Arbiters = arbiters
	servers.Pow = pow.NewService(&pow.Config{
		PayToAddr:   cfg.PowConfiguration.PayToAddr,
		MinerInfo:   cfg.PowConfiguration.MinerInfo,
		Chain:       chain,
		ChainParams: activeNetParams,
		TxMemPool:   txMemPool,
		BlkMemPool:  blockMemPool,
		BroadcastBlock: func(block *types.Block) {
			hash := block.Hash()
			server.RelayInventory(msg.NewInvVect(msg.InvTypeBlock, &hash), block)
		},
		Arbitrators: arbiters,
	})

	if arbitrator != nil {
		arbitrator.Start()
		defer arbitrator.Stop()
	}

	// initialize producer state after arbiters has initialized
	if err = chain.InitializeProducersState(interrupt.C, pgBar.Start,
		pgBar.Increase); err != nil {
		printErrorAndExit(err)
	}
	pgBar.Stop()

	if arbitrator != nil {
		arbitrator.Recover()
	}

	log.Info("Start the P2P networks")
	server.Start()
	defer server.Stop()

	log.Info("Start services")
	go httpjsonrpc.StartRPCServer()
	if config.Parameters.HttpRestStart {
		go httprestful.StartServer()
	}
	if config.Parameters.HttpWsStart {
		go httpwebsocket.Start()
	}
	if config.Parameters.HttpInfoStart {
		go httpnodeinfo.StartServer()
	}

	go printSyncState(chainStore, server)

	waitForSyncFinish(server, interrupt.C)
	if interrupt.Interrupted() {
		return
	}
	log.Info("Start consensus")
	if config.Parameters.PowConfiguration.AutoMining {
		log.Info("Start POW Services")
		go servers.Pow.Start()
	}

	<-interrupt.C
}

func printErrorAndExit(err error) {
	log.Error(err)
	os.Exit(-1)
}

func waitForSyncFinish(server elanet.Server, interrupt <-chan struct{}) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

out:
	for {
		select {
		case <-ticker.C:
			if server.IsCurrent() {
				break out
			}

		case <-interrupt:
			break out
		}
	}
}

func printSyncState(db blockchain.IChainStore, server elanet.Server) {
	statlog := elalog.NewBackend(logger.Writer()).Logger("STAT",
		elalog.LevelInfo)

	ticker := time.NewTicker(printStateInterval)
	defer ticker.Stop()

	for range ticker.C {
		var buf bytes.Buffer
		buf.WriteString("-> ")
		buf.WriteString(strconv.FormatUint(uint64(db.GetHeight()), 10))
		peers := server.ConnectedPeers()
		buf.WriteString(" [")
		for i, p := range peers {
			buf.WriteString(strconv.FormatUint(uint64(p.ToPeer().Height()), 10))
			buf.WriteString(" ")
			buf.WriteString(p.ToPeer().String())
			if i != len(peers)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString("]")
		statlog.Info(buf.String())
	}
}
