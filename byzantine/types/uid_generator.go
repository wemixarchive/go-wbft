package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type UIDGeneratorImpl struct{}

func NewUIDGenerator() *UIDGeneratorImpl {
	return &UIDGeneratorImpl{}
}

func (ug *UIDGeneratorImpl) Generate(attackType AttackType, code MessageCode,
	sequence, round uint64) string {
	return fmt.Sprintf("%s-%d-%d-%d",
		AttachTypeToString(attackType), code, sequence, round)
}

func (ug *UIDGeneratorImpl) Parse(uid string) (AttackType, MessageCode,
	uint64, uint64, error) {
	parts := strings.Split(uid, "-")
	if len(parts) != 4 {
		return "", 0, 0, 0, errors.New("invalid UID format")
	}

	attackType := StringToAttackType(parts[0])
	code, _ := strconv.ParseUint(parts[1], 10, 64)
	sequence, _ := strconv.ParseUint(parts[2], 10, 64)
	round, _ := strconv.ParseUint(parts[3], 10, 64)

	return attackType, MessageCode(code), sequence, round, nil
}
