package eccpow

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/Onther-Tech/go-ethereum/common"
	"github.com/Onther-Tech/go-ethereum/consensus"
	"github.com/Onther-Tech/go-ethereum/core/types"
	"github.com/Onther-Tech/go-ethereum/metrics"
	"github.com/Onther-Tech/go-ethereum/rpc"
)

type ECC struct {
	shared *ECC

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters
	hashrate metrics.Meter // Meter tracking the average hashrate

	// Remote sealer related fields
	workCh       chan *sealTask   // Notification channel to push new work and relative result channel to remote sealer
	fetchWorkCh  chan *sealWork   // Channel used for remote sealer to fetch mining work
	submitWorkCh chan *mineResult // Channel used for remote sealer to submit their mining result
	fetchRateCh  chan chan uint64 // Channel used to gather submitted hash rate for local or remote sealer.
	submitRateCh chan *hashrate   // Channel used for remote sealer to submit their mining hashrate

	lock      sync.Mutex      // Ensures thread safety for the in-memory caches and mining fields
	closeOnce sync.Once       // Ensures exit channel will not be closed twice.
	exitCh    chan chan error // Notification channel to exiting backend threads

}

// sealTask wraps a seal block with relative result channel for remote sealer thread.
type sealTask struct {
	block   *types.Block
	results chan<- *types.Block
}

// mineResult wraps the pow solution parameters for the specified block.
type mineResult struct {
	nonce     types.BlockNonce
	mixDigest common.Hash
	hash      common.Hash

	errc chan error
}

// hashrate wraps the hash rate submitted by the remote sealer.
type hashrate struct {
	id   common.Hash
	ping time.Time
	rate uint64

	done chan struct{}
}

// sealWork wraps a seal work package for remote sealer.
type sealWork struct {
	errc chan error
	res  chan [4]string
}

// hasher is a repetitive hasher allowing the same hash data structures to be
// reused between hash runs instead of requiring new ones to be created.
//var hasher func(dest []byte, data []byte)

var two256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

//const cross_err = 0.01

//type (
//	intMatrix   [][]int
//	floatMatrix [][]float64
//)

//RunOptimizedConcurrencyLDPC use goroutine for mining block
func RunOptimizedConcurrencyLDPC(header *types.Header) ([]int, []int, uint64) {
	//Need to set difficulty before running LDPC
	// Number of goroutines : 500, Number of attempts : 50000 Not bad

	var LDPCNonce uint64
	var hashVector []int
	var outputWord []int

	var wg sync.WaitGroup
	var outerLoopSignal = make(chan struct{})
	var innerLoopSignal = make(chan struct{})
	var goRoutineSignal = make(chan struct{})
	parameters, _ := setParameters(header)

outerLoop:
	for {
		select {
		// If outerLoopSignal channel is closed, then break outerLoop
		case <-outerLoopSignal:
			break outerLoop

		default:
			// Defined default to unblock select statement
		}

	innerLoop:
		for i := 0; i < runtime.NumCPU(); i++ {
			select {
			// If innerLoop signal is closed, then break innerLoop and close outerLoopSignal
			case <-innerLoopSignal:
				close(outerLoopSignal)
				break innerLoop

			default:
				// Defined default to unblock select statement
			}

			wg.Add(1)
			go func(goRoutineSignal chan struct{}) {
				defer wg.Done()
				//goRoutineNonce := generateRandomNonce()
				//fmt.Printf("Initial goroutine Nonce : %v\n", goRoutineNonce)

				var goRoutineHashVector []int
				var goRoutineOutputWord []int

				var serializedHeader = string(header.ParentHash[:])
				var serializedHeaderWithNonce string

				var encryptedHeaderWithNonce [32]byte
				H := generateH(parameters)
				colInRow, rowInCol := generateQ(parameters, H)

				select {
				case <-goRoutineSignal:
					break

				default:
				attemptLoop:
					for attempt := 0; attempt < 5000; attempt++ {
						goRoutineNonce := generateRandomNonce()
						serializedHeaderWithNonce = serializedHeader + strconv.FormatUint(goRoutineNonce, 10)
						encryptedHeaderWithNonce = sha256.Sum256([]byte(serializedHeaderWithNonce))

						goRoutineHashVector = generateHv(parameters, encryptedHeaderWithNonce)
						goRoutineHashVector, goRoutineOutputWord, _ = OptimizedDecoding(header, goRoutineHashVector, H, rowInCol, colInRow)
						flag := MakeDecision(header, colInRow, goRoutineOutputWord)

						select {
						case <-goRoutineSignal:
							// fmt.Println("goRoutineSignal channel is already closed")
							break attemptLoop
						default:
							if flag {
								close(goRoutineSignal)
								close(innerLoopSignal)
								fmt.Printf("Codeword is founded with nonce = %d\n", goRoutineNonce)
								LDPCNonce = goRoutineNonce
								hashVector = goRoutineHashVector
								outputWord = goRoutineOutputWord
								break attemptLoop
							}
						}
						//goRoutineNonce++
					}
				}
			}(goRoutineSignal)
		}
		// Need to wait to prevent memory leak
		wg.Wait()
	}

	return hashVector, outputWord, LDPCNonce
}

//MakeDecision check outputWord is valid or not using colInRow
func MakeDecision(header *types.Header, colInRow [][]int, outputWord []int) bool {
	parameters, difficultyLevel := setParameters(header)
	for i := 0; i < parameters.m; i++ {
		sum := 0
		for j := 0; j < parameters.wr; j++ {
			//	fmt.Printf("i : %d, j : %d, m : %d, wr : %d \n", i, j, m, wr)
			sum = sum + outputWord[colInRow[j][i]]
		}
		if sum%2 == 1 {
			return false
		}
	}

	var numOfOnes int
	for _, val := range outputWord {
		numOfOnes += val
	}

	if numOfOnes >= Table[difficultyLevel].decisionFrom &&
		numOfOnes <= Table[difficultyLevel].decisionTo &&
		numOfOnes%Table[difficultyLevel].decisionStep == 0 {
		return true
	}

	return false
}

//func newIntMatrix(rows, cols int) intMatrix {
//	m := intMatrix(make([][]int, rows))
//	for i := range m {
//		m[i] = make([]int, cols)
//	}
//	return m
//}
//
//func newFloatMatrix(rows, cols int) floatMatrix {
//	m := floatMatrix(make([][]float64, rows))
//	for i := range m {
//		m[i] = make([]float64, cols)
//	}
//	return m
//}

// NewShared creates a full sized ecc PoW shared between all requesters running
// in the same process.
//func NewShared() *ECC {
//	return &ecc{shared: sharedecc}
//}

func NewTester(notify []string, noverify bool) *ECC {
	ecc := &ECC{
		update:       make(chan struct{}),
		hashrate:     metrics.NewMeterForced(),
		workCh:       make(chan *sealTask),
		fetchWorkCh:  make(chan *sealWork),
		submitWorkCh: make(chan *mineResult),
		fetchRateCh:  make(chan chan uint64),
		submitRateCh: make(chan *hashrate),
		exitCh:       make(chan chan error),
	}
	go ecc.remote(notify, noverify)
	return ecc
}

// Close closes the exit channel to notify all backend threads exiting.
func (ecc *ECC) Close() error {
	var err error
	ecc.closeOnce.Do(func() {
		// Short circuit if the exit channel is not allocated.
		if ecc.exitCh == nil {
			return
		}
		errc := make(chan error)
		ecc.exitCh <- errc
		err = <-errc
		close(ecc.exitCh)
	})
	return err
}

// Threads returns the number of mining threads currently enabled. This doesn't
// necessarily mean that mining is running!
func (ecc *ECC) Threads() int {
	ecc.lock.Lock()
	defer ecc.lock.Unlock()

	return ecc.threads
}

// SetThreads updates the number of mining threads currently enabled. Calling
// this method does not start mining, only sets the thread count. If zero is
// specified, the miner will use all cores of the machine. Setting a thread
// count below zero is allowed and will cause the miner to idle, without any
// work being done.
func (ecc *ECC) SetThreads(threads int) {
	ecc.lock.Lock()
	defer ecc.lock.Unlock()

	// If we're running a shared PoW, set the thread count on that instead
	if ecc.shared != nil {
		ecc.shared.SetThreads(threads)
		return
	}
	// Update the threads and ping any running seal to pull in any changes
	ecc.threads = threads
	select {
	case ecc.update <- struct{}{}:
	default:
	}
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
// Note the returned hashrate includes local hashrate, but also includes the total
// hashrate of all remote miner.
func (ecc *ECC) Hashrate() float64 {
	// Short circuit if we are run the ecc in normal/test mode.

	var res = make(chan uint64, 1)

	select {
	case ecc.fetchRateCh <- res:
	case <-ecc.exitCh:
		// Return local hashrate only if ecc is stopped.
		return ecc.hashrate.Rate1()
	}

	// Gather total submitted hash rate of remote sealers.
	return ecc.hashrate.Rate1() + float64(<-res)
}

// APIs implements consensus.Engine, returning the user facing RPC APIs.
func (ecc *ECC) APIs(chain consensus.ChainReader) []rpc.API {
	// In order to ensure backward compatibility, we exposes ecc RPC APIs
	// to both eth and ecc namespaces.
	return []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   &API{ecc},
			Public:    true,
		},
		{
			Namespace: "ecc",
			Version:   "1.0",
			Service:   &API{ecc},
			Public:    true,
		},
	}
}

//// SeedHash is the seed to use for generating a verification cache and the mining
//// dataset.
func SeedHash(block *types.Block) []byte {
	return block.ParentHash().Bytes()
}
