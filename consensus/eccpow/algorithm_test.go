package eccpow

import (
	"fmt"
	"testing"

	"github.com/Onther-Tech/go-ethereum/common/hexutil"
)

func TestLDPC(t *testing.T) {
	// Create a block to verify
	prev_hash := hexutil.MustDecode("0xd783efa4d392943503f28438ad5830b2d5964696ffc285f338585e9fe0a37a05")
	cur_hash := hexutil.MustDecode("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
	n = 24
	wc = 3
	wr = 6

	//digest, result := hashimotoLight(32*1024, cache, hash, nonce)
	//if !bytes.Equal(digest, wantDigest) {
	//	t.Errorf("light hashimoto digest mismatch: have %x, want %x", digest, wantDigest)
	//}
	//if !bytes.Equal(result, wantResult) {
	//	t.Errorf("light hashimoto result mismatch: have %x, want %x", result, wantResult)
	//}
	nonce := runLDPC(prev_hash, cur_hash)
	fmt.Print(nonce)
	//if !bytes.Equal(digest, wantDigest) {
	//	t.Errorf("full hashimoto digest mismatch: have %x, want %x", digest, wantDigest)
	//}
	//if !bytes.Equal(result, wantResult) {
	//	t.Errorf("full hashimoto result mismatch: have %x, want %x", result, wantResult)
	//}
}
