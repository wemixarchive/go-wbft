// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

package common

// SecretKey represents a BLS secret or private key.
type SecretKey interface {
	PublicKey() PublicKey
	Sign(msg []byte) Signature
	Marshal() []byte
}

// PublicKey represents a BLS public key.
type PublicKey interface {
	Marshal() []byte
	Copy() PublicKey
	Aggregate(p2 PublicKey) PublicKey
	IsInfinite() bool
	Equals(p2 PublicKey) bool
}

// Signature represents a BLS signature.
type Signature interface {
	Verify(pubKey PublicKey, msg []byte) bool
	// Deprecated: Use FastAggregateVerify or use this method in spectests only.
	AggregateVerify(pubKeys []PublicKey, msgs [][32]byte) bool
	FastAggregateVerify(pubKeys []PublicKey, msg [32]byte) bool
	Eth2FastAggregateVerify(pubKeys []PublicKey, msg [32]byte) bool
	Marshal() []byte
	Copy() Signature
}
