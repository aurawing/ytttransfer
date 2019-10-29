package ytttransfer

import (
	"bytes"
	"crypto/sha256"
	"strings"

	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

//Verify verify signature for data by public key
func Verify(publicKey string, data []byte, signature string) bool {
	if !strings.HasPrefix(signature, "SIG_K1_") {
		return false
	}
	signature = strings.TrimPrefix(signature, "SIG_K1_")
	sigbytes, _ := base58.Decode(signature)
	sign := sigbytes[0:65]
	checksum := sigbytes[65:]
	sign1 := append([]byte{}, sign...)
	ck := ripemd160Sum(append(sign1, 'K', '1'))
	if !bytes.Equal(checksum, ck[0:4]) {
		return false
	}
	sign[0] -= 4
	sign[0] -= 27
	signx := append(sign[1:65], sign[0])
	recPubkey, err := ecrypto.Ecrecover(sha256Sum(data), signx)
	if err != nil {
		return false
	}
	recPublicKey, err := ecrypto.UnmarshalPubkey(recPubkey)
	if err != nil {
		return false
	}
	recPublicKeyBytes := ecrypto.CompressPubkey(recPublicKey)
	checksum = ripemd160Sum(recPublicKeyBytes)
	rawRecPublicKeyBytes := append(recPublicKeyBytes, checksum[0:4]...)
	rawPublicKeyBytes, err := base58.Decode(publicKey)
	if err != nil {
		return false
	}
	if len(rawPublicKeyBytes) == 33 {
		return bytes.Equal(rawPublicKeyBytes, rawRecPublicKeyBytes[0:33])
	} else if len(rawPublicKeyBytes) == 37 {
		return bytes.Equal(rawPublicKeyBytes, rawRecPublicKeyBytes[0:37])
	} else {
		return false
	}
}

func sha256Sum(bytes []byte) []byte {
	h := sha256.New()
	h.Write(bytes)
	return h.Sum(nil)
}

func ripemd160Sum(bytes []byte) []byte {
	h := ripemd160.New()
	h.Write(bytes)
	return h.Sum(nil)
}
