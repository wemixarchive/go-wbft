// Modification Copyright 2024 The Wemix Authors
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
//
// This file is derived from quorum/consensus/istanbul/wbft/core/final_committed.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"math/big"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func (c *Core) handleFinalCommittedMsg() error {
	c.currentLogger(true, nil).Info("WBFT: handle final committed")
	c.startNewRound(common.Big0)
	c.byzantineFakeRoundChange()
	return nil
}

func (c *Core) byzantineFakeRoundChange() {
	hook := c.backend.ByzantineHook()
	if hook == nil {
		return
	}

	attacks := hook.GetExecutableAttacks(btypes.MessageCodeRoundChange, c.current.Sequence().Uint64(), c.current.Round().Uint64())

	if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil {
		for _, field := range at.FakeParams.Fields {
			switch field.Target {
			case btypes.FakeTargetRound:
				val, err := field.ValueToUint64()
				if err != nil {
					log.Error("[BYZ] Conversion failed", "err", err)
				} else {
					log.Info("[BYZ] attack", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.FakeParams)
					hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
					round := new(big.Int).SetUint64(val)
					c.broadcastRoundChange(round)
				}
			}
		}
	}
}
