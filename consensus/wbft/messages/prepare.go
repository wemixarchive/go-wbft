// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/wbft/types/prepare.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type StoredPrepare struct {
	Seq         *big.Int
	Round       *big.Int
	Digest      common.Hash
	PrepareSeal []byte
}

// A WBFT PREPARE message.
type Prepare struct {
	CommonPayload
	Digest      common.Hash
	PrepareSeal []byte
}

func NewPrepare(sequence *big.Int, round *big.Int, digest common.Hash, seal []byte) *Prepare {
	return &Prepare{
		CommonPayload: CommonPayload{
			code:     PrepareCode,
			Sequence: sequence,
			Round:    round,
		},
		Digest:      digest,
		PrepareSeal: seal,
	}
}

func NewPrepareWithSigAndSource(sequence *big.Int, round *big.Int, digest common.Hash, signature []byte, source common.Address, seal []byte) *Prepare {
	prepare := NewPrepare(sequence, round, digest, seal)
	prepare.signature = signature
	prepare.source = source
	return prepare
}

func (p *Prepare) String() string {
	return fmt.Sprintf("Prepare {seq=%v, round=%v, digest=%v}", p.Sequence, p.Round, p.Digest.Hex())
}

func (p *Prepare) EncodePayloadForSigning() ([]byte, error) {
	return rlp.EncodeToBytes(
		[]interface{}{
			p.Code(),
			[]interface{}{p.Sequence, p.Round, p.Digest, p.PrepareSeal},
		})
}

func (p *Prepare) EncodeRLP(w io.Writer) error {
	return rlp.Encode(
		w,
		[]interface{}{
			[]interface{}{
				p.Sequence,
				p.Round,
				p.Digest,
				p.PrepareSeal},
			p.signature,
		})
}

func (p *Prepare) DecodeRLP(stream *rlp.Stream) error {
	var message struct {
		Payload struct {
			Sequence    *big.Int
			Round       *big.Int
			Digest      common.Hash
			PrepareSeal []byte
		}
		Signature []byte
	}
	if err := stream.Decode(&message); err != nil {
		return err
	}
	p.code = PrepareCode
	p.Sequence = message.Payload.Sequence
	p.Round = message.Payload.Round
	p.Digest = message.Payload.Digest
	p.PrepareSeal = message.Payload.PrepareSeal
	p.signature = message.Signature
	return nil
}

func (p *Prepare) DeepCopy() *Prepare {
	if p == nil {
		return nil
	}

	// DeepCopy CommonPayload
	cp := p.CommonPayload
	return &Prepare{
		CommonPayload: CommonPayload{
			code:   cp.code,
			source: cp.source,
			Sequence: func() *big.Int {
				if cp.Sequence == nil {
					return nil
				}
				return new(big.Int).Set(cp.Sequence)
			}(),
			Round: func() *big.Int {
				if cp.Round == nil {
					return nil
				}
				return new(big.Int).Set(cp.Round)
			}(),
			signature: func() []byte {
				if cp.signature == nil {
					return nil
				}
				s := make([]byte, len(cp.signature))
				copy(s, cp.signature)
				return s
			}(),
		},
		Digest: p.Digest,
		PrepareSeal: func() []byte {
			if p.PrepareSeal == nil {
				return nil
			}
			s := make([]byte, len(p.PrepareSeal))
			copy(s, p.PrepareSeal)
			return s
		}(),
	}
}
