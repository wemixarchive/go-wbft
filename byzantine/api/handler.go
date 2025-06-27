package api

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// Handler handles API requests
type Handler struct {
	service        types.ByzantineService
	attackManager  types.AttackManager
	messageStorage types.MessageStorage
	historyStorage types.HistoryStorage
}

// NewHandler creates a new API handler
func NewHandler(svc types.ByzantineService) *Handler {
	return &Handler{
		service:        svc,
		attackManager:  svc.GetAttackManager(),
		messageStorage: svc.GetMessageStorage(),
		historyStorage: svc.GetHistoryStorage(),
	}
}

// GetByzantineTests returns all registered Byzantine tests
func (h *Handler) GetByzantineTests() ([]types.AttackConfig, error) {
	attacks := h.service.ListAttacks()

	// Convert to AttackInfo format
	infos := make([]types.AttackConfig, len(attacks))
	for i, attack := range attacks {
		infos[i] = types.AttackConfig{
			UID:        attack.UID,
			Name:       attack.Name,
			Type:       attack.Type,
			Sequence:   attack.Sequence,
			Round:      attack.Round,
			Code:       attack.Code,
			Status:     attack.Status,
			CreatedAt:  attack.CreatedAt,
			Targets:    attack.Targets,
			Parameters: attack.Parameters,
			//FakeMessage:      h.getFakeMessageFromOptions(attack.Options),
			//WithValidMessage: h.getWithValidMessageFromOptions(attack.Options),
		}
	}

	return infos, nil
}

// StopByzantineTests stops Byzantine tests by UIDs
func (h *Handler) StopByzantineTests(uids []uint64) error {
	var lastErr error
	successCount := 0

	for _, uid := range uids {
		if err := h.service.CancelAttack(uid); err != nil {
			log.Error("Failed to stop Byzantine test", "uid", uid, "error", err)
			lastErr = err
		} else {
			successCount++
		}
	}

	if lastErr != nil && successCount == 0 {
		return fmt.Errorf("failed to stop all attacks: %w", lastErr)
	}

	if lastErr != nil {
		return fmt.Errorf("stopped %d/%d attacks, last error: %w", successCount, len(uids), lastErr)
	}

	return nil
}

// RegisterSilentMessage registers a silent message attack
func (h *Handler) RegisterSilentMessage(params types.SilentMessageParams) error {
	// Validate parameters
	if err := h.validateSilentMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:     fmt.Sprintf("silent_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeSilentMessage,
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     types.MessageCode(params.Code),
		Parameters: map[string]interface{}{
			"direction": params.Direction,
		},
		Targets:   params.Targets,
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register silent attack: %w", err)
	}

	log.Info("Silent message attack registered", "uid", uid, "sequence", params.Sequence, "round", params.Round)
	return nil
}

// RegisterTamperedMessage registers a tampered message attack
func (h *Handler) RegisterTamperedMessage(params types.TamperedMessageParams) error {
	// Validate parameters
	if err := h.validateTamperedMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	tamperFieldsMaps := make([]map[string]interface{}, len(params.TamperFields))
	for i, field := range params.TamperFields {
		tamperFieldsMaps[i] = map[string]interface{}{
			"target": field.Target,
			"value":  field.Value,
		}
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:     fmt.Sprintf("tampered_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeTamperedMessage,
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     types.MessageCode(params.Code),
		Targets:  params.Targets,
		Parameters: map[string]interface{}{
			"tamperFields":     tamperFieldsMaps,
			"withValidMessage": params.WithValidMessage,
			"delay":            params.Delay,
			"code":             params.Code,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register tampered attack: %w", err)
	}

	log.Info("Tampered message attack registered", "uid", uid, "sequence", params.Sequence, "round", params.Round)
	return nil
}

// RegisterFakeMessage registers a fake message attack
func (h *Handler) RegisterFakeMessage(params types.FakeMessageParams) error {
	// Validate parameters
	if err := h.validateFakeMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:     fmt.Sprintf("fake_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeFakeMessage,
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     types.MessageCode(params.Code),
		Targets:  params.Targets,
		Parameters: map[string]interface{}{
			"fakeMessage": params.FakeMessage,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register fake attack: %w", err)
	}

	log.Info("Fake message attack registered", "uid", uid, "sequence", params.Sequence, "round", params.Round)
	return nil
}

// RegisterOmitMessage registers an omit message attack
func (h *Handler) RegisterOmitMessage(params types.OmitMessageParams) error {
	// Validate parameters
	if err := h.validateOmitMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:     fmt.Sprintf("omit_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeOmitMessage,
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     types.MessageCode(params.Code),
		Targets:  params.Targets,
		Parameters: map[string]interface{}{
			"cmd": params.Cmd,
			"cnt": params.Cnt,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register omit attack: %w", err)
	}

	log.Info("Omit message attack registered", "uid", uid, "sequence", params.Sequence, "round", params.Round)
	return nil
}

// RegisterRoleSpoofedMessage registers a role spoofing attack
func (h *Handler) RegisterRoleSpoofedMessage(params types.RoleSpoofParams) error {
	// Validate parameters
	if err := h.validateRoleSpoofParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:     fmt.Sprintf("rolespoof_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeRoleSpoofed,
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     types.MessageCode(params.Code),
		Targets:  params.Targets,
		Parameters: map[string]interface{}{
			"fakeMessage": params.FakeMessage,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register role spoof attack: %w", err)
	}

	log.Info("Role spoof attack registered", "uid", uid, "sequence", params.Sequence, "round", params.Round)
	return nil
}

// RegisterReplayMessage registers a replay attack
func (h *Handler) RegisterReplayMessage(params types.ReplayMessageParams) error {
	// Validate parameters
	if err := h.validateReplayMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:     fmt.Sprintf("replay_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeReplay,
		Sequence: params.Sequence,
		Round:    params.Round,
		Code:     types.MessageCode(params.Code),
		Targets:  params.Targets,
		Parameters: map[string]interface{}{
			"oriSequence":     params.OriSequence,
			"oriRound":        params.OriRound,
			"useOriginalView": params.UseOriginalView,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register replay attack: %w", err)
	}

	log.Info("Replay attack registered", "uid", uid, "sequence", params.Sequence, "round", params.Round)
	return nil
}

// UpgradeGovContract handles governance contract upgrade
func (h *Handler) UpgradeGovContract() error {
	// TODO: Implement governance contract upgrade logic
	log.Info("Governance contract upgrade requested")
	return nil
}

// Validation methods
func (h *Handler) validateSilentMessageParams(params types.SilentMessageParams) error {
	if params.Sequence == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	if params.Direction == 0 || params.Direction > 3 {
		return fmt.Errorf("invalid direction: must be 1, 2, or 3")
	}
	return nil
}

func (h *Handler) validateTamperedMessageParams(params types.TamperedMessageParams) error {
	if params.Sequence == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	if len(params.TamperFields) == 0 {
		return fmt.Errorf("at least one tamper field is required")
	}
	return nil
}

func (h *Handler) validateFakeMessageParams(params types.FakeMessageParams) error {
	if params.Sequence == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateOmitMessageParams(params types.OmitMessageParams) error {
	if params.Sequence == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	if params.Cmd == 0 {
		return fmt.Errorf("command is required")
	}
	return nil
}

func (h *Handler) validateRoleSpoofParams(params types.RoleSpoofParams) error {
	if params.Sequence == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateReplayMessageParams(params types.ReplayMessageParams) error {
	if params.OriSequence == 0 {
		return fmt.Errorf("original sequence number is required")
	}
	if params.Sequence == 0 {
		return fmt.Errorf("target sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

// Helper methods
func (h *Handler) getFakeMessageFromOptions(options map[string]interface{}) string {
	if options == nil {
		return ""
	}
	if msg, ok := options["fakeMessage"].(string); ok {
		return msg
	}
	return ""
}

func (h *Handler) getWithValidMessageFromOptions(options map[string]interface{}) bool {
	if options == nil {
		return false
	}
	if withValid, ok := options["withValidMessage"].(bool); ok {
		return withValid
	}
	return false
}
