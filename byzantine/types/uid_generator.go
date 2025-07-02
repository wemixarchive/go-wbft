package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// UIDGeneratorImpl implements UIDGenerator
type UIDGeneratorImpl struct{}

var _ UIDGenerator = (*UIDGeneratorImpl)(nil)

func NewUIDGenerator() *UIDGeneratorImpl {
	return &UIDGeneratorImpl{}
}

// Generate creates UID - 3 parameters for backward compatibility
func (ug *UIDGeneratorImpl) Generate(attackType AttackType, sequence, round uint64) string {
	typeStr := AttackTypeToString(attackType)
	return fmt.Sprintf("%s-%d-%d", typeStr, sequence, round)
}

// GenerateWithRange creates UID with range support - 4 parameters
// Format: "attackType-sequenceStart[-sequenceEnd]-round"
func (ug *UIDGeneratorImpl) GenerateWithRange(attackType AttackType, sequenceStart, sequenceEnd, round uint64) string {
	typeStr := AttackTypeToString(attackType)

	//if sequenceEnd == 0 || sequenceEnd == sequenceStart {
	//	return fmt.Sprintf("%s-%d-%d", typeStr, sequenceStart, round)
	//}

	return fmt.Sprintf("%s-%d-%d-%d", typeStr, sequenceStart, sequenceEnd, round)
}

// GenerateForLookup create UID for looking up attacks at specific sequence
func (ug *UIDGeneratorImpl) GenerateForLookup(attackType AttackType, sequence, round uint64) []string {
	typeStr := AttackTypeToString(attackType)

	patterns := []string{
		fmt.Sprintf("%s-%d-%d", typeStr, sequence, round),
		fmt.Sprintf("%s-%d-*", typeStr, sequence),
		fmt.Sprintf("%s-*-*", typeStr),
	}

	return patterns
}

// ParseRange parse UID to extract attack info including range
func (ug *UIDGeneratorImpl) ParseRange(uid string) (AttackType, uint64, uint64, uint64, error) {
	parts := strings.Split(uid, "-")
	if len(parts) < 4 {
		return "", 0, 0, 0, fmt.Errorf("invalid UID format: %s", uid)
	}

	attackType := StringToAttackType(parts[0])
	sequenceStart, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil && parts[1] != "*" {
		return "", 0, 0, 0, fmt.Errorf("invalid sequence start: %s", parts[1])
	}

	var sequenceEnd uint64
	var round uint64

	if len(parts) == 3 {
		sequenceEnd = sequenceStart
		round, err = strconv.ParseUint(parts[2], 10, 64)
		if err != nil && parts[2] != "*" {
			return "", 0, 0, 0, fmt.Errorf("invalid round: %s", parts[2])
		}
	} else if len(parts) == 4 {
		sequenceEnd, err = strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return "", 0, 0, 0, fmt.Errorf("invalid sequence end: %s", parts[2])
		}

		round, err = strconv.ParseUint(parts[3], 10, 64)
		if err != nil && parts[3] != "*" {
			return "", 0, 0, 0, fmt.Errorf("invalid round: %s", parts[3])
		}
	}

	return attackType, sequenceStart, sequenceEnd, round, nil
}

// IsWildcard IsWildCard checks if UID contains wildcards
func (ug *UIDGeneratorImpl) IsWildcard(uid string) bool {
	return strings.Contains(uid, "*")
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
