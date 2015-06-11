package blockexplorer

import (
	"path/filepath"
	"testing"

	"github.com/NebulousLabs/Sia/build"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/modules/consensus"
	"github.com/NebulousLabs/Sia/modules/gateway"
	"github.com/NebulousLabs/Sia/modules/miner"
	"github.com/NebulousLabs/Sia/modules/transactionpool"
	"github.com/NebulousLabs/Sia/modules/wallet"
)

// Explorer tester struct is the helper object for explorer
// testing. It holds the helper modules for its testing
type explorerTester struct {
	cs      *consensus.State
	gateway modules.Gateway
	miner   modules.Miner
	tpool   modules.TransactionPool
	wallet  modules.Wallet

	explorer *BlockExplorer

	csUpdateChan     <-chan struct{}
	tpoolUpdateChan  <-chan struct{}
	minerUpdateChan  <-chan struct{}
	walletUpdateChan <-chan struct{}

	t *testing.T
}

// csUpdateWait blocks until a consensus update has propagated to all
// modules.
func (et *explorerTester) csUpdateWait() {
	<-et.csUpdateChan
	et.tpUpdateWait()
}

// tpUpdateWait blocks until a transaction pool update has propagated to all
// modules.
func (ht *explorerTester) tpUpdateWait() {
	<-ht.tpoolUpdateChan
	<-ht.minerUpdateChan
	<-ht.walletUpdateChan
}

func CreateExplorerTester(name string, t *testing.T) *explorerTester {
	testdir := build.TempDir(modules.HostDir, name)

	// Create the gateway.
	g, err := gateway.New(":0", filepath.Join(testdir, modules.GatewayDir))
	if err != nil {
		t.Fatal(err)
	}

	// Create the consensus set.
	cs, err := consensus.New(g, filepath.Join(testdir, modules.ConsensusDir))
	if err != nil {
		t.Fatal(err)
	}

	// Create the transaction pool.
	tp, err := transactionpool.New(cs, g)
	if err != nil {
		t.Fatal(err)
	}

	// Create the wallet.
	w, err := wallet.New(cs, tp, filepath.Join(testdir, modules.WalletDir))
	if err != nil {
		t.Fatal(err)
	}

	// Create the miner.
	m, err := miner.New(cs, tp, w)
	if err != nil {
		t.Fatal(err)
	}

	// Create the blockexplorer
	be, err := New(cs)
	if err != nil {
		t.Fatal(err)
	}

	et := &explorerTester{
		cs:      cs,
		gateway: g,
		miner:   m,
		tpool:   tp,
		wallet:  w,

		explorer: be,

		csUpdateChan:     cs.ConsensusSetNotify(),
		tpoolUpdateChan:  tp.TransactionPoolNotify(),
		minerUpdateChan:  m.MinerNotify(),
		walletUpdateChan: w.WalletNotify(),

		t: t,
	}
	return et
}
