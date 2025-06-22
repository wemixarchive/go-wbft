package byzantine

import (
	"github.com/ethereum/go-ethereum/byzantine/types"
)

// baseAttack provides common functionality
type baseAttack struct {
	id     string
	config types.AttackConfig
}

func (a *baseAttack) ID() string {
	return a.id
}

func (a *baseAttack) Config() types.AttackConfig {
	return a.config
}

func (a *baseAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.config.Enabled &&
		a.config.Sequence == sequence &&
		a.config.Round == round
}

func (a *baseAttack) Execute(ctx AttackContext) error {
	// TODO: Implement base attack logic
	return nil
}

// silentAttack implementation
type silentAttack struct {
	baseAttack
}

func (a *silentAttack) Execute(ctx AttackContext) error {
	// TODO: Implement silent attack logic
	return nil
}

// tamperAttack implementation
type tamperAttack struct {
	baseAttack
}

func (a *tamperAttack) Execute(ctx AttackContext) error {
	// TODO: Implement tamper attack logic
	return nil
}

// fakeAttack implementation
type fakeAttack struct {
	baseAttack
}

func (a *fakeAttack) Execute(ctx AttackContext) error {
	// TODO: Implement fake attack logic
	return nil
}
