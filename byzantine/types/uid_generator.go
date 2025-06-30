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

func (ug *UIDGeneratorImpl) Generate(attackType AttackType, sequence, round uint64) string {
	return fmt.Sprintf("%s-%d-%d",
		AttachTypeToString(attackType), sequence, round)
}

func (ug *UIDGeneratorImpl) Parse(uid string) (AttackType, uint64, uint64, error) {
	parts := strings.Split(uid, "-")
	if len(parts) != 3 {
		return "", 0, 0, errors.New("invalid UID format")
	}

	attackType := StringToAttackType(parts[0])
	sequence, _ := strconv.ParseUint(parts[1], 10, 64)
	round, _ := strconv.ParseUint(parts[2], 10, 64)

	return attackType, sequence, round, nil
}
