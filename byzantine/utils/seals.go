package utils

import (
	"bytes"

	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/crypto/bls"

	"github.com/ethereum/go-ethereum/log"
)

// CreateFakeSeal creates a fake seal with the given sealer index
func CreateFakeSeal(fakeIndex int) wbft.SealData {
	// Use the provided fake index which should be outside the valid validator range
	fakeSealerIndex := uint32(fakeIndex)

	return wbft.SealData{
		Sealer: fakeSealerIndex,
		Seal:   generateFakeSignature(),
	}
}

// generateFakeSignature generates a valid BLS signature using a fake key
func generateFakeSignature() []byte {
	// Generate a fake BLS secret key with a fixed seed
	seed := bytes.Repeat([]byte{0x42}, 32) // 32 bytes seed
	fakeSecretKey, err := bls.GenerateKey(seed)
	if err != nil {
		log.Debug("BYZ: Failed to generate fake BLS key", "err", err)
		// Fallback to simple fake signature
		sig := make([]byte, 96)
		for i := range sig {
			sig[i] = byte(i % 256)
		}
		return sig
	}

	// Create a fake seal message similar to PrepareSeal function
	// This creates a 32-byte hash that looks like a real seal
	fakeBlockHash := bytes.Repeat([]byte{0xAB}, 32)   // Fake block hash
	fakeSealMessage := append(fakeBlockHash, byte(0)) // 0 for SealTypePrepare, 1 for SealTypeCommit

	// Sign with the fake key to get a valid BLS signature
	signature := fakeSecretKey.Sign(fakeSealMessage)

	return signature.Marshal()
}

// CountSealers counts the number of sealers in a SealerSet
func CountSealers(sealers []byte, validatorCount int) int {
	count := 0
	for i := 0; i < validatorCount; i++ {
		byteIndex := i / 8
		bitIndex := uint(i % 8)
		if byteIndex < len(sealers) && (sealers[byteIndex]&(1<<bitIndex)) != 0 {
			count++
		}
	}
	return count
}

// ModifySealerSet replaces some existing sealers with fake signers
// Returns the number of sealers actually modified
func ModifySealerSet(sealers []byte, fakeSignerCount int, validatorCount int) int {
	currentSealerCount := CountSealers(sealers, validatorCount)

	// Limit modifications to current sealer count
	toModify := fakeSignerCount
	if toModify > currentSealerCount {
		toModify = currentSealerCount
	}

	// Clear existing sealers
	cleared := 0
	for i := 0; i < validatorCount && cleared < toModify; i++ {
		byteIndex := i / 8
		bitIndex := uint(i % 8)
		if byteIndex < len(sealers) && (sealers[byteIndex]&(1<<bitIndex)) != 0 {
			// Clear this sealer
			sealers[byteIndex] &^= (1 << bitIndex)
			cleared++
			log.Debug("BYZ: Cleared sealer", "index", i)
		}
	}

	// Add fake signers by setting cleared positions
	added := 0
	for i := 0; i < validatorCount && added < cleared; i++ {
		byteIndex := i / 8
		bitIndex := uint(i % 8)
		if byteIndex < len(sealers) && (sealers[byteIndex]&(1<<bitIndex)) == 0 {
			// Set this position as fake sealer
			sealers[byteIndex] |= (1 << bitIndex)
			added++
			log.Debug("BYZ: Added fake sealer", "index", i)
		}
	}

	return added
}
