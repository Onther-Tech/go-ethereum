package eccpow

import (
	"testing"

	"github.com/Onther-Tech/go-ethereum/core/types"
)

func TestNonceDecoding(t *testing.T) {
	LDPCNonce := generateRandomNonce()
	EncodedNonce := types.EncodeNonce(LDPCNonce)
	DecodedNonce := EncodedNonce.Uint64()

	if LDPCNonce == DecodedNonce {
		t.Logf("LDPCNonce : %v\n", LDPCNonce)
		t.Logf("Decoded Nonce : %v\n", DecodedNonce)
	} else {
		t.Errorf("LDPCNonce : %v\n", LDPCNonce)
		t.Errorf("Decoded Nonce : %v\n", DecodedNonce)
	}
}
