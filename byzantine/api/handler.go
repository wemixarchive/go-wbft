package api

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// Handler handles API requests
type Handler struct {
	service       types.ByzantineService
	attackManager types.AttackManager
}

// NewHandler creates a new API handler
func NewHandler(svc types.ByzantineService) *Handler {
	return &Handler{
		service:       svc,
		attackManager: svc.GetAttackManager(),
	}
}

// GetByzantineTests returns all registered Byzantine tests
func (h *Handler) GetByzantineTests() ([]types.AttackConfig, error) {
	attacks := h.service.ListAttacks()

	// Convert to AttackInfo format
	infos := make([]types.AttackConfig, len(attacks))
	for i, attack := range attacks {
		infos[i] = types.AttackConfig{
			UID:            attack.UID,
			Name:           attack.Name,
			Type:           attack.Type,
			Enabled:        attack.Enabled,
			SequenceStart:  attack.SequenceStart,
			SequenceEnd:    attack.SequenceEnd,
			Round:          attack.Round,
			ExecutionCount: attack.ExecutionCount,
			Status:         attack.Status,
			Parameters:     attack.Parameters,
			CreatedAt:      attack.CreatedAt,
		}
	}

	return infos, nil
}

// StopByzantineTests stops Byzantine tests by UIDs
func (h *Handler) StopByzantineTests(uids []string) error {
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

// RegisterSetMessagePolicy registers a set message policy
func (h *Handler) RegisterSetMessagePolicy(params types.SetMessagePolicyParams) error {
	// Validate parameters
	if err := h.validateMessagePolicyParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	fieldsMap := make([]map[string]interface{}, len(params.Fields))
	for i, field := range params.Fields {
		fieldsMap[i] = map[string]interface{}{
			"target": field.Target,
			"value":  field.Value,
		}
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:          fmt.Sprintf("policy_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeMessagePolicy,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code":   params.Code,
			"fields": fieldsMap,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register store attack: %w", err)
	}

	log.Info("Store attack registered", "uid", uid, "config", config)
	return nil
}

// RegisterTamperedMessage registers a tampered message attack
func (h *Handler) RegisterTamperedMessage(params types.TamperedMessageParams) error {
	// Validate parameters
	if err := h.validateTamperedMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	fieldsMap := make([]map[string]interface{}, len(params.Fields))
	for i, field := range params.Fields {
		fieldsMap[i] = map[string]interface{}{
			"target": field.Target,
			"value":  field.Value,
		}
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:          fmt.Sprintf("tampered_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeTamperedMessage,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code":   params.Code,
			"fields": fieldsMap,
		},
		Enabled:   params.Enabled,
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register tampered attack: %w", err)
	}

	log.Info("Tampered message attack registered", "uid", uid, "config", config)
	return nil
}

// RegisterFakeMessage registers a fake message attack
func (h *Handler) RegisterFakeMessage(params types.FakeMessageParams) error {
	// Validate parameters
	if err := h.validateFakeMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	fieldsMap := make([]map[string]interface{}, len(params.Fields))
	for i, field := range params.Fields {
		fieldsMap[i] = map[string]interface{}{
			"target": field.Target,
			"value":  field.Value,
		}
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:          fmt.Sprintf("fake_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeFakeMessage,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code":   params.Code,
			"fields": fieldsMap,
		},
		Enabled:   params.Enabled,
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register fake attack: %w", err)
	}

	log.Info("Fake message attack registered", "uid", uid, "config", config)
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
		Name:          fmt.Sprintf("omit_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeOmitMessage,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code": params.Code,
			"cmd":  params.Cmd,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register omit attack: %w", err)
	}

	log.Info("Omit message attack registered", "uid", uid, "config", config)
	return nil
}

// RegisterRoleSpoofedMessage registers a role spoofing attack
func (h *Handler) RegisterRoleSpoofedMessage(params types.RoleSpoofParams) error {
	// Validate parameters
	if err := h.validateRoleSpoofParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	fieldsMap := make([]map[string]interface{}, len(params.Fields))
	for i, field := range params.Fields {
		fieldsMap[i] = map[string]interface{}{
			"target": field.Target,
			"value":  field.Value,
		}
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:          fmt.Sprintf("rolespoof_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeRoleSpoofed,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code":   params.Code,
			"fields": fieldsMap,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register role spoof attack: %w", err)
	}

	log.Info("Role spoof attack registered", "uid", uid, "config", config)
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
		Name:          fmt.Sprintf("replay_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeReplay,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code":            params.Code,
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

	log.Info("Replay attack registered", "uid", uid, "config", config)
	return nil
}

// RegisterStoreMessage registers a store message
func (h *Handler) RegisterStoreMessage(params types.StoreMessageParams) error {
	// Validate parameters
	if err := h.validateStoreMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:          fmt.Sprintf("store_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeStoreMessage,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code": params.Code,
		},
		Status:    types.AttackStatusPending,
		CreatedAt: time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register store attack: %w", err)
	}

	log.Info("Store attack registered", "uid", uid, "config", config)
	return nil
}

// RegisterDosMessage registers a DoS message attack
func (h *Handler) RegisterDosMessage(params types.DosMessageParams) error {
	// Validate parameters
	if err := h.validateDosMessageParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create attack configuration
	config := types.AttackConfig{
		Name:          fmt.Sprintf("dos_%d_%d_%d", params.SequenceStart, params.SequenceEnd, params.Round),
		Type:          types.AttackTypeDos,
		SequenceStart: params.SequenceStart,
		SequenceEnd:   params.SequenceEnd,
		Round:         params.Round,
		Parameters: map[string]interface{}{
			"code":  params.Code,
			"cmd":   params.Cmd,
			"cnt":   params.Cnt,
			"delay": params.Delay,
		},
		MaxExecutionCount: params.MaxExecutionCount,
		Enabled:           params.Enabled,
		Status:            types.AttackStatusPending,
		CreatedAt:         time.Now(),
	}

	// Register attack
	uid, err := h.service.RegisterAttack(config)
	if err != nil {
		return fmt.Errorf("failed to register DoS attack: %w", err)
	}

	log.Info("DoS message attack registered", "uid", uid, "config", config)
	return nil
}

// UpgradeGovContract handles governance contract upgrade
func (h *Handler) UpgradeGovContract() error {
	// TODO: Implement governance contract upgrade logic
	log.Info("Governance contract upgrade requested")
	return nil
}

// RegisterAttacks registers multiple attacks at once
func (h *Handler) RegisterAttacks(params types.RegisterAttacksParams) (*types.BatchRegisterResponse, error) {
	response := &types.BatchRegisterResponse{
		Results: make([]types.RegisterAttackResponse, 0, len(params.Attacks)),
		Total:   len(params.Attacks),
	}

	for _, attackAPI := range params.Attacks {
		// Convert AttackAPIConfig to AttackConfig
		config := types.AttackConfig{
			Name:              attackAPI.Name,
			Type:              types.StringToAttackType(attackAPI.Type),
			Enabled:           attackAPI.Enabled,
			SequenceStart:     attackAPI.SequenceStart,
			SequenceEnd:       attackAPI.SequenceEnd,
			Round:             attackAPI.Round,
			MaxExecutionCount: attackAPI.MaxExecutionCount,
			Parameters:        attackAPI.Parameters,
			Status:            types.AttackStatusPending,
			CreatedAt:         time.Now(),
		}

		// Register attack
		uid, err := h.service.RegisterAttack(config)
		result := types.RegisterAttackResponse{
			UID:     uid,
			Success: err == nil,
		}
		if err != nil {
			result.Error = err.Error()
			response.Failed++
		} else {
			response.Success++
		}
		response.Results = append(response.Results, result)
	}

	return response, nil
}

// GetActiveAttacks returns all active attacks
func (h *Handler) GetActiveAttacks() ([]types.AttackStatusResponse, error) {
	attacks := h.service.ListAttacks()
	activeAttacks := make([]types.AttackStatusResponse, 0)

	for _, attack := range attacks {
		// Filter for active attacks
		if attack.Status == types.AttackStatusActive || attack.Status == types.AttackStatusPending {
			statusResp := h.convertToStatusResponse(attack)
			activeAttacks = append(activeAttacks, statusResp)
		}
	}

	return activeAttacks, nil
}

// GetAttackStatus returns the status of a specific attack
func (h *Handler) GetAttackStatus(uid string) (*types.AttackStatusResponse, error) {
	attack, err := h.attackManager.GetAttack(uid)
	if err != nil {
		return nil, fmt.Errorf("attack not found: %w", err)
	}

	config := attack.GetConfig()
	statusResp := h.convertToStatusResponse(config)
	return &statusResp, nil
}

// GetAttackMetrics returns metrics about attacks
func (h *Handler) GetAttackMetrics() (*types.Metrics, error) {
	status := h.service.GetStatus()

	metrics := &types.Metrics{
		AttacksRegistered: int64(len(h.service.ListAttacks())),
		AttacksExecuted:   int64(status.ExecutedAttacks),
		AttacksFailed:     int64(status.FailedAttacks),
		MessagesStored:    int64(status.StoredMessages),
	}

	// Calculate uptime if service is running
	if status.Running && status.StartedAt != nil {
		metrics.Uptime = int64(time.Since(*status.StartedAt).Seconds())
	}

	return metrics, nil
}

// convertToStatusResponse converts AttackConfig to AttackStatusResponse
func (h *Handler) convertToStatusResponse(attack types.AttackConfig) types.AttackStatusResponse {
	resp := types.AttackStatusResponse{
		UID:            attack.UID,
		Name:           attack.Name,
		Type:           string(attack.Type),
		Status:         string(attack.Status),
		Enabled:        attack.Enabled,
		SequenceRange:  attack.GetSequenceRange(),
		Round:          attack.Round,
		ExecutionCount: attack.ExecutionCount,
		Parameters:     attack.Parameters,
		CreatedAt:      attack.CreatedAt.Unix(),
	}

	if attack.ExecutedAt != nil {
		executedAt := attack.ExecutedAt.Unix()
		resp.ExecutedAt = &executedAt
	}

	return resp
}

// Validation methods
func (h *Handler) validateMessagePolicyParams(params types.SetMessagePolicyParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateTamperedMessageParams(params types.TamperedMessageParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number or sequence range is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	if len(params.Fields) == 0 {
		return fmt.Errorf("at least one field is required")
	}
	return nil
}

func (h *Handler) validateFakeMessageParams(params types.FakeMessageParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number or sequence range is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateOmitMessageParams(params types.OmitMessageParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
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
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateReplayMessageParams(params types.ReplayMessageParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateStoreMessageParams(params types.StoreMessageParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}
	return nil
}

func (h *Handler) validateDosMessageParams(params types.DosMessageParams) error {
	// Check sequence
	if params.SequenceStart == 0 || params.SequenceEnd == 0 {
		return fmt.Errorf("sequence number is required")
	}
	if params.Code == 0 {
		return fmt.Errorf("message code is required")
	}

	// Validate cmd value (0, 1, or 2)
	if params.Cmd < 0 || params.Cmd > 2 {
		return fmt.Errorf("invalid cmd value: %d, must be 0 (valid), 1 (sequence), or 2 (round)", params.Cmd)
	}

	if params.Cnt == 0 {
		return fmt.Errorf("cnt must be greater than 0")
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
