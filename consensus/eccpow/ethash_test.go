// Copyright 2017 The go-ethereum Authors
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

package eccpow

import (
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Onther-Tech/go-ethereum/common"
	"github.com/Onther-Tech/go-ethereum/common/hexutil"
	"github.com/Onther-Tech/go-ethereum/core/types"
)

// Tests that eccpow works correctly in test mode.
func TestTestMode(t *testing.T) {
	header := &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(100)}

	eccpow := NewTester(nil, false)
	defer eccpow.Close()

	results := make(chan *types.Block)
	err := eccpow.Seal(nil, types.NewBlockWithHeader(header), results, nil)
	if err != nil {
		t.Fatalf("failed to seal block: %v", err)
	}
	select {
	case block := <-results:
		header.Nonce = types.EncodeNonce(block.Nonce())
		header.MixDigest = block.MixDigest()
		if err := eccpow.VerifySeal(nil, header); err != nil {
			t.Fatalf("unexpected verification error: %v", err)
		}
	case <-time.NewTimer(time.Second).C:
		t.Error("sealing result timeout")
	}
}

// This test checks that cache lru logic doesn't crash under load.
// It reproduces https://github.com/Onther-Tech/go-ethereum/issues/14943
func TestCacheFileEvict(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "eccpow-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	e := New(Config{CachesInMem: 3, CachesOnDisk: 10, CacheDir: tmpdir, PowMode: ModeTest}, nil, false)
	defer e.Close()

	workers := 8
	epochs := 100
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go verifyTest(&wg, e, i, epochs)
	}
	wg.Wait()
}

func verifyTest(wg *sync.WaitGroup, e *EccPoW, workerIndex, epochs int) {
	defer wg.Done()

	const wiggle = 4 * epochLength
	r := rand.New(rand.NewSource(int64(workerIndex)))
	for epoch := 0; epoch < epochs; epoch++ {
		block := int64(epoch)*epochLength - wiggle/2 + r.Int63n(wiggle)
		if block < 0 {
			block = 0
		}
		header := &types.Header{Number: big.NewInt(block), Difficulty: big.NewInt(100)}
		e.VerifySeal(nil, header)
	}
}

func TestRemoteSealer(t *testing.T) {
	eccpow := NewTester(nil, false)
	defer eccpow.Close()

	api := &API{eccpow}
	if _, err := api.GetWork(); err != errNoMiningWork {
		t.Error("expect to return an error indicate there is no mining work")
	}
	header := &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(100)}
	block := types.NewBlockWithHeader(header)
	sealhash := eccpow.SealHash(header)

	// Push new work.
	results := make(chan *types.Block)
	eccpow.Seal(nil, block, results, nil)

	var (
		work [4]string
		err  error
	)
	if work, err = api.GetWork(); err != nil || work[0] != sealhash.Hex() {
		t.Error("expect to return a mining work has same hash")
	}

	if res := api.SubmitWork(types.BlockNonce{}, sealhash, common.Hash{}); res {
		t.Error("expect to return false when submit a fake solution")
	}
	// Push new block with same block number to replace the original one.
	header = &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(1000)}
	block = types.NewBlockWithHeader(header)
	sealhash = eccpow.SealHash(header)
	eccpow.Seal(nil, block, results, nil)

	if work, err = api.GetWork(); err != nil || work[0] != sealhash.Hex() {
		t.Error("expect to return the latest pushed work")
	}
}

func TestHashRate(t *testing.T) {
	var (
		hashrate = []hexutil.Uint64{100, 200, 300}
		expect   uint64
		ids      = []common.Hash{common.HexToHash("a"), common.HexToHash("b"), common.HexToHash("c")}
	)
	eccpow := NewTester(nil, false)
	defer eccpow.Close()

	if tot := eccpow.Hashrate(); tot != 0 {
		t.Error("expect the result should be zero")
	}

	api := &API{eccpow}
	for i := 0; i < len(hashrate); i += 1 {
		if res := api.SubmitHashRate(hashrate[i], ids[i]); !res {
			t.Error("remote miner submit hashrate failed")
		}
		expect += uint64(hashrate[i])
	}
	if tot := eccpow.Hashrate(); tot != float64(expect) {
		t.Error("expect total hashrate should be same")
	}
}

func TestClosedRemoteSealer(t *testing.T) {
	eccpow := NewTester(nil, false)
	time.Sleep(1 * time.Second) // ensure exit channel is listening
	eccpow.Close()

	api := &API{eccpow}
	if _, err := api.GetWork(); err != erreccpowStopped {
		t.Error("expect to return an error to indicate eccpow is stopped")
	}

	if res := api.SubmitHashRate(hexutil.Uint64(100), common.HexToHash("a")); res {
		t.Error("expect to return false when submit hashrate to a stopped eccpow")
	}
}
