package api

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine"
)

// byzantineAPI implementation
type byzantineAPI struct {
	builder AttackBuilder
	attacks map[string]Attack
	mu      sync.RWMutex
	manager byzantine.AttackManager // Will be implemented in next step
}

var _ byzantine.ByzantineAPI = (*byzantineAPI)(nil)

// NewByzantineAPI creates a new Byzantine API instance
func NewByzantineAPI(builder AttackBuilder) byzantine.ByzantineAPI {
	return &byzantineAPI{
		builder: builder,
		attacks: make(map[string]Attack),
	}
}

// NewByzantineAPIWithManager creates a new Byzantine API instance with manager
func NewByzantineAPIWithManager(builder AttackBuilder, manager byzantine.AttackManager) byzantine.ByzantineAPI {
	return &byzantineAPI{
		builder: builder,
		attacks: make(map[string]Attack),
		manager: manager,
	}
}

// ConfigureAttack configures a new attack
func (api *byzantineAPI) ConfigureAttack(params AttackParams) (attackID string, err error) {
	// Validate parameters
	if err := api.validateParams(params); err != nil {
		return "", err
	}

	// Build attack using builder pattern
	attack, err := api.builder.
		WithType(params.Type).
		WithTiming(params.Sequence, params.Round).
		WithTargets(params.Target).
		WithOptions(params.Options).
		Build()

	if err != nil {
		return "", fmt.Errorf("failed to build attack: %w", err)
	}

	// Register with manager if available
	if api.manager != nil {
		if err := api.manager.Register(attack); err != nil {
			return "", fmt.Errorf("failed to register attack: %w", err)
		}
	} else {
		// Fallback to local storage
		api.mu.Lock()
		api.attacks[attack.ID()] = attack
		api.mu.Unlock()
	}

	// TODO: Register with manager
	// TODO: Start event collection

	return attack.ID(), nil
}

// StopAttack stops an active attack
func (api *byzantineAPI) StopAttack(attackID string) error {
	api.mu.Lock()
	defer api.mu.Unlock()

	if _, exists := api.attacks[attackID]; !exists {
		return errors.New("attack not found")
	}

	delete(api.attacks, attackID)

	// TODO: Notify manager
	// TODO: Stop event collection

	return nil
}

// ListAttacks returns list of configured attacks
func (api *byzantineAPI) ListAttacks() []byzantine.AttackInfo {
	api.mu.RLock()
	defer api.mu.RUnlock()

	infos := make([]byzantine.AttackInfo, 0, len(api.attacks))
	for _, attack := range api.attacks {
		// For now, return basic info
		// TODO: Get detailed status from manager
		infos = append(infos, byzantine.AttackInfo{
			ID:     attack.ID(),
			Status: "active",
		})
	}

	return infos
}

// validateParams validates attack parameters
func (api *byzantineAPI) validateParams(params AttackParams) error {
	if params.Type == 0 {
		return errors.New("attack type is required")
	}

	if params.Sequence == 0 {
		return errors.New("sequence number is required")
	}

	// Add more validation as needed

	return nil
}

// ByzantineTests returns all registered Byzantine tests
func (api *byzantineAPI) ByzantineTests() []byzantine.AttackInfo {
	attacks := api.manager.ListAttacks()
	log.Warn("ByzantineTests called", "count", len(attacks))
	return attacks
}

// StopByzantineTests stops Byzantine tests by UIDs
func (api *byzantineAPI) StopByzantineTests(uids []uint64) error {
	log.Info("StopByzantineTests called", "uids", uids)

	for _, uid := range uids {
		if err := api.manager.UnregisterAttack(uid); err != nil {
			log.Error("Failed to stop Byzantine test", "uid", uid, "error", err)
			return err
		}
	}

	return nil
}

// SilentMessage configures a silent message attack
func (api *byzantineAPI) SilentMessage(params SilentMessageParams) error {
	log.Info("SilentMessage called",
		"sequence", params.Sequence,
		"round", params.Round,
		"code", params.Code,
		"direction", params.Direction)

	return api.manager.ConfigureSilentMessage(
		params.Sequence,
		params.Round,
		params.Code,
		params.Direction,
		params.Targets,
	)
}

// SendTamperedMessage configures a tampered message attack
func (api *byzantineAPI) SendTamperedMessage(params TamperedMessageParams) error {
	log.Info("SendTamperedMessage called",
		"sequence", params.Sequence,
		"round", params.Round,
		"code", params.Code)

	// Convert TamperFields to types.TamperField
	tamperFields := make([]types.TamperField, len(params.TamperFields))
	for i, field := range params.TamperFields {
		tamperFields[i] = types.TamperField{
			Target: field.Target,
			Value:  field.Value,
		}
	}

	return api.manager.ConfigureTamperedMessage(types.TamperMessageParams{
		Sequence:         params.Sequence,
		Round:            params.Round,
		Code:             params.Code,
		TamperFields:     tamperFields,
		WithValidMessage: params.WithValidMessage,
		Delay:            params.Delay,
		Targets:          params.Targets,
	})
}

// SendFakeMessage configures a fake message attack
func (api *byzantineAPI) SendFakeMessage(params FakeMessageParams) error {
	log.Info("SendFakeMessage called",
		"sequence", params.Sequence,
		"round", params.Round,
		"code", params.Code)

	return api.manager.ConfigureFakeMessage(types.FakeMessageParams{
		Sequence:    params.Sequence,
		Round:       params.Round,
		Code:        params.Code,
		FakeMessage: params.FakeMessage,
		Targets:     params.Targets,
	})
}

// SendOmitMessage configures an omit message attack
func (api *byzantineAPI) SendOmitMessage(params OmitMessageParams) error {
	log.Info("SendOmitMessage called",
		"sequence", params.Sequence,
		"round", params.Round,
		"code", params.Code,
		"cmd", params.Cmd,
		"cnt", params.Cnt)

	return api.manager.ConfigureOmitMessage(types.OmitMessageParams{
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     params.Code,
		Cmd:      params.Cmd,
		Cnt:      params.Cnt,
		Targets:  params.Targets,
	})
}

// SendRoleSpoofedMessage configures a role spoofing attack
func (api *byzantineAPI) SendRoleSpoofedMessage(params RoleSpoofParams) error {
	log.Info("SendRoleSpoofedMessage called",
		"sequence", params.Sequence,
		"round", params.Round,
		"code", params.Code)

	return api.manager.ConfigureRoleSpoofedMessage(types.RoleSpoofParams{
		Sequence:    params.Sequence,
		Round:       params.Round,
		Code:        params.Code,
		FakeMessage: params.FakeMessage,
		Targets:     params.Targets,
	})
}

// SendReplayMessage configures a replay attack
func (api *byzantineAPI) SendReplayMessage(params ReplayMessageParams) error {
	log.Info("SendReplayMessage called",
		"oriSequence", params.OriSequence,
		"oriRound", params.OriRound,
		"sequence", params.Sequence,
		"round", params.Round,
		"code", params.Code,
		"useOriginalView", params.UseOriginalView)

	return api.manager.ConfigureReplayMessage(types.ReplayMessageParams{
		OriSequence:     params.OriSequence,
		OriRound:        params.OriRound,
		Sequence:        params.Sequence,
		Round:           params.Round,
		UseOriginalView: params.UseOriginalView,
		Code:            params.Code,
		Targets:         params.Targets,
	})
}

// UpgradeGovContract configures governance contract upgrade
func (api *byzantineAPI) UpgradeGovContract() error {
	log.Info("UpgradeGovContract called")
	return nil
}
