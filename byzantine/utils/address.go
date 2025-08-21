package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
)

// fakeSignerInfo holds fake signer's keys and address
type fakeSignerInfo struct {
	PrivateKey *ecdsa.PrivateKey
	BlsKey     bls.SecretKey
	Address    common.Address
}

// GenerateFakeSigners creates fake signers with valid ECDSA and BLS keys
func GenerateFakeSigners(addresses []common.Address) []fakeSignerInfo {
	signers := make([]fakeSignerInfo, len(addresses))

	for i, addr := range addresses {
		// Generate new ECDSA private key
		ecdsaKey, err := crypto.GenerateKey()
		if err != nil {
			log.Warn("BYZ: Failed to generate ECDSA key", "err", err)
			continue
		}

		// Derive BLS key from ECDSA key (deterministic)
		blsKey, err := bls.DeriveFromECDSA(ecdsaKey)
		if err != nil {
			log.Warn("BYZ: Failed to derive BLS key", "err", err)
			continue
		}

		signers[i] = fakeSignerInfo{
			PrivateKey: ecdsaKey,
			BlsKey:     blsKey,
			Address:    addr,
		}

		log.Debug("BYZ: Generated fake signer",
			"addr", addr,
			"ecdsaAddr", crypto.PubkeyToAddress(ecdsaKey.PublicKey),
			"blsPubKey", blsKey.PublicKey().Marshal())
	}

	return signers
}

// GenerateRandomAddress generates a random Ethereum address using hash of current timestamp and random bytes
func GenerateRandomAddress() common.Address {
	// Use timestamp and some fixed bytes to generate a deterministic but unique address
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(timestamp))
	return common.BytesToAddress(hash[:20])
}
