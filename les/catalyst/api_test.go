// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package catalyst

import (
	"math/big"
	"testing"

	"github.com/ethxdao/go-ethereum/consensus/ethash"
	"github.com/ethxdao/go-ethereum/core"
	//"github.com/ethxdao/go-ethereum/core/beacon"
	"github.com/ethxdao/go-ethereum/core/rawdb"
	"github.com/ethxdao/go-ethereum/core/types"
	"github.com/ethxdao/go-ethereum/crypto"
	"github.com/ethxdao/go-ethereum/eth/downloader"
	"github.com/ethxdao/go-ethereum/eth/ethconfig"
	"github.com/ethxdao/go-ethereum/les"
	"github.com/ethxdao/go-ethereum/node"
	"github.com/ethxdao/go-ethereum/params"
)

var (
	// testKey is a private key to use for funding a tester account.
	testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

	// testAddr is the Ethereum address of the tester account.
	testAddr = crypto.PubkeyToAddress(testKey.PublicKey)

	testBalance = big.NewInt(2e18)
)

func generatePreMergeChain(n int) (*core.Genesis, []*types.Header, []*types.Block) {
	db := rawdb.NewMemoryDatabase()
	config := params.AllEthashProtocolChanges
	genesis := &core.Genesis{
		Config:    config,
		Alloc:     core.GenesisAlloc{testAddr: {Balance: testBalance}},
		ExtraData: []byte("test genesis"),
		Timestamp: 9000,
		BaseFee:   big.NewInt(params.InitialBaseFee),
	}
	gblock := genesis.MustCommit(db)
	engine := ethash.NewFaker()
	blocks, _ := core.GenerateChain(config, gblock, engine, db, n, nil)
	totalDifficulty := big.NewInt(0)

	var headers []*types.Header
	for _, b := range blocks {
		totalDifficulty.Add(totalDifficulty, b.Difficulty())
		headers = append(headers, b.Header())
	}
	config.TerminalTotalDifficulty = totalDifficulty

	return genesis, headers, blocks
}

func TestEth2DeepReorg(t *testing.T) {
	// TODO (MariusVanDerWijden) TestEth2DeepReorg is currently broken, because it tries to reorg
	// before the totalTerminalDifficulty threshold
	/*
		genesis, preMergeBlocks := generatePreMergeChain(core.TriesInMemory * 2)
		n, ethservice := startEthService(t, genesis, preMergeBlocks)
		defer n.Close()

		var (
			api    = NewConsensusAPI(ethservice, nil)
			parent = preMergeBlocks[len(preMergeBlocks)-core.TriesInMemory-1]
			head   = ethservice.BlockChain().CurrentBlock().NumberU64()
		)
		if ethservice.BlockChain().HasBlockAndState(parent.Hash(), parent.NumberU64()) {
			t.Errorf("Block %d not pruned", parent.NumberU64())
		}
		for i := 0; i < 10; i++ {
			execData, err := api.assembleBlock(AssembleBlockParams{
				ParentHash: parent.Hash(),
				Timestamp:  parent.Time() + 5,
			})
			if err != nil {
				t.Fatalf("Failed to create the executable data %v", err)
			}
			block, err := ExecutableDataToBlock(ethservice.BlockChain().Config(), parent.Header(), *execData)
			if err != nil {
				t.Fatalf("Failed to convert executable data to block %v", err)
			}
			newResp, err := api.ExecutePayload(*execData)
			if err != nil || newResp.Status != "VALID" {
				t.Fatalf("Failed to insert block: %v", err)
			}
			if ethservice.BlockChain().CurrentBlock().NumberU64() != head {
				t.Fatalf("Chain head shouldn't be updated")
			}
			if err := api.setCanonical(block.Hash()); err != nil {
				t.Fatalf("Failed to set head: %v", err)
			}
			if ethservice.BlockChain().CurrentBlock().NumberU64() != block.NumberU64() {
				t.Fatalf("Chain head should be updated")
			}
			parent, head = block, block.NumberU64()
		}
	*/
}

// startEthService creates a full node instance for testing.
func startLesService(t *testing.T, genesis *core.Genesis, headers []*types.Header) (*node.Node, *les.LightEthereum) {
	t.Helper()

	n, err := node.New(&node.Config{})
	if err != nil {
		t.Fatal("can't create node:", err)
	}
	ethcfg := &ethconfig.Config{
		Genesis:        genesis,
		Ethash:         ethash.Config{PowMode: ethash.ModeFake},
		SyncMode:       downloader.LightSync,
		TrieDirtyCache: 256,
		TrieCleanCache: 256,
		LightPeers:     10,
	}
	lesService, err := les.New(n, ethcfg)
	if err != nil {
		t.Fatal("can't create eth service:", err)
	}
	if err := n.Start(); err != nil {
		t.Fatal("can't start node:", err)
	}
	if _, err := lesService.BlockChain().InsertHeaderChain(headers, 0); err != nil {
		n.Close()
		t.Fatal("can't import test headers:", err)
	}
	return n, lesService
}

func encodeTransactions(txs []*types.Transaction) [][]byte {
	var enc = make([][]byte, len(txs))
	for i, tx := range txs {
		enc[i], _ = tx.MarshalBinary()
	}
	return enc
}
