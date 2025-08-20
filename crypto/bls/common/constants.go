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

const (
	BLS_SECRET_KEY_LENGTH = 32
	BLS_PUBLIC_KEY_LENGTH = 48
	BLS_SIGNATURE_LENGTH  = 96
)

var (
	// ZeroSecretKey represents a zero secret key.
	ZeroSecretKey = [32]byte{}
	// InfinitePublicKey represents an infinite public key (G1 Point at Infinity).
	InfinitePublicKey = [BLS_PUBLIC_KEY_LENGTH]byte{0xC0}
	// InfiniteSignature represents an infinite signature (G2 Point at Infinity).
	InfiniteSignature = [96]byte{0xC0}
)
