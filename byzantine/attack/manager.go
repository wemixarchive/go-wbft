package attack

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/google/uuid"
)

// EventCollector interface for data collection
//type EventCollector interface {
//	StartCollection(dataReqs []types.DataRequirement) error
//	StopCollection(attackID string) error
//}

//// AttackManager interface for managing attacks
//type AttackManager interface {
//	Register(attack Attack) error
//	Unregister(attackID string) error
//	GetAttack(attackID string) (Attack, bool)
//	GetAttacksToExecute(sequence, round uint64, msgCode uint64) []Attack
//	GetAttackStatus(attackID string) AttackStatus
//	MarkExecuted(attackID string)
//	ListAllAttacks() []Attack
//	CleanupExecuted() int
//}

// attackManager implements AttackManager interface
type attackManager struct {
	attacks   map[string]*attackInfo
	collector types.EventCollector
	builder   types.AttackBuilder

	mu sync.RWMutex
}

// attackInfo holds attack metadata
type attackInfo struct {
	id     string
	attack types.Attack
	status types.AttackStatus
}

var _ types.AttackManager = (*attackManager)(nil)

// NewAttackManager creates a new attack manager
func NewAttackManager(collector types.EventCollector, builder types.AttackBuilder) types.AttackManager {
	return &attackManager{
		attacks:   make(map[string]*attackInfo),
		builder:   builder,
		collector: collector,
	}
}

func (m *attackManager) Builder() types.AttackBuilder {
	return m.builder
}

// Register registers a new attack
func (m *attackManager) Register(attack types.Attack) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if attack already exists
	if _, exists := m.attacks[attack.ID()]; exists {
		return errors.New("attack already registered")
	}

	// Start data collection if required
	if m.collector != nil {
		dataReqs := attack.RequiresData()
		if len(dataReqs) > 0 {
			if err := m.collector.Start(dataReqs); err != nil {
				return fmt.Errorf("failed to start data collection: %w", err)
			}
		}
	}

	// Register the attack
	m.attacks[attack.ID()] = &attackInfo{
		attack: attack,
		status: types.AttackStatusActive,
	}

	return nil
}

// Unregister removes an attack
func (m *attackManager) Unregister(attackID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if attack exists
	if _, exists := m.attacks[attackID]; !exists {
		return errors.New("attack not found")
	}

	// Stop data collection
	if m.collector != nil {
		if err := m.collector.Stop(attackID); err != nil {
			// Log error but continue with unregistration
			// In real implementation, proper logging would be added
		}
	}

	// Remove the attack
	delete(m.attacks, attackID)

	return nil
}

// GetAttack retrieves a specific attack
func (m *attackManager) GetAttack(attackID string) (types.Attack, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if info, exists := m.attacks[attackID]; exists {
		return info.attack, true
	}

	return nil, false
}

// GetAttacksToExecute finds attacks that should execute for given conditions
func (m *attackManager) GetAttacksToExecute(sequence, round uint64, msgCode uint64) []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var attacksToExecute []types.Attack

	for _, info := range m.attacks {
		// Only check active attacks
		if info.status != types.AttackStatusActive {
			continue
		}

		// Check if attack should execute
		if info.attack.ShouldExecute(sequence, round, msgCode) {
			attacksToExecute = append(attacksToExecute, info.attack)
		}
	}

	return attacksToExecute
}

// GetAttackStatus returns the status of an attack
func (m *attackManager) GetAttackStatus(attackID string) types.AttackStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if info, exists := m.attacks[attackID]; exists {
		return info.status
	}

	return types.AttackStatusCancelled // Default for non-existent attacks
}

// MarkExecuted marks an attack as executed
func (m *attackManager) MarkExecuted(attackID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if info, exists := m.attacks[attackID]; exists {
		info.status = types.AttackStatusExecuted
	}
}

// ListAllAttacks returns all registered attacks
func (m *attackManager) ListAllAttacks() []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attacks := make([]types.Attack, 0, len(m.attacks))
	for _, info := range m.attacks {
		attacks = append(attacks, info.attack)
	}

	return attacks
}

// CleanupExecuted removes all executed attacks
func (m *attackManager) CleanupExecuted() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := []string{}

	// Find executed attacks
	for id, info := range m.attacks {
		if info.status == types.AttackStatusExecuted {
			toRemove = append(toRemove, id)
		}
	}

	// Remove executed attacks
	for _, id := range toRemove {
		// Stop data collection
		if m.collector != nil {
			m.collector.Stop(id)
		}
		delete(m.attacks, id)
	}

	return len(toRemove)
}

// LoadAttack loads an attack from configuration
func (m *attackManager) LoadAttack(cfg types.AttackConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use builder to create attack
	attack, err := m.builder.
		Reset().
		WithConfig(cfg).
		WithType(cfg.Type).
		WithTiming(cfg.Sequence, cfg.Round).
		WithTargets(cfg.GetTargetAddresses()).
		WithOptions(cfg.Params).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build attack: %w", err)
	}

	// Generate unique ID for the attack
	id := fmt.Sprintf("%s-%s", cfg.Name, uuid.New().String()[:8])
	log.Info("id generated", "id", id, "name", cfg.Name, "type", cfg.ConvertTypeToString())

	// Create attack instance
	m.attacks[attack.ID()] = &attackInfo{
		id:     id,
		attack: attack,
		status: types.AttackStatusPending,
	}

	//switch cfg.Type {
	//case types.AttackTypeSilent:
	//	// TODO: Create silent attack implementation
	//	attack = &attack2.baseAttack{
	//		id:     fmt.Sprintf("%s-%s", cfg.Name, uuid.New().String()[:8]),
	//		config: cfg,
	//	}
	//case types.AttackTypeTamper:
	//	// TODO: Create tamper attack implementation
	//	attack = &attack2.baseAttack{
	//		id:     fmt.Sprintf("%s-%s", cfg.Name, uuid.New().String()[:8]),
	//		config: cfg,
	//	}
	//case types.AttackTypeFake:
	//	// TODO: Create fake attack implementation
	//	attack = &attack2.baseAttack{
	//		id:     fmt.Sprintf("%s-%s", cfg.Name, uuid.New().String()[:8]),
	//		config: cfg,
	//	}
	//default:
	//	return fmt.Errorf("unsupported attack type: %s", cfg.ConvertTypeToString())
	//}

	m.attacks[attack.ID()] = &attackInfo{
		attack: attack,
		status: types.AttackStatusPending,
	}

	log.Debug("Attack loaded", "id", attack.ID(), "type", cfg.ConvertTypeToString())
	return err
}

// TODO:
// should be refactored to use a proper attack factory
func (m *attackManager) ConfigureAttack(params types.AttackParams) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Convert AttackParams to types.AttackType
	attackType := types.AttackType(params.Type)

	// Use builder to create attack
	attack, err := m.builder.
		Reset().
		WithType(attackType).
		WithTiming(params.Sequence, params.Round).
		WithTargets(params.Targets).
		WithOptions(params.Options).
		Build()

	if err != nil {
		return "", fmt.Errorf("failed to build attack: %w", err)
	}

	m.attacks[attack.ID()] = &attackInfo{
		attack: attack,
		status: types.AttackStatusActive,
	}

	log.Info("Attack configured via API", "id", attack.ID(), "type", params.Type)
	return attack.ID(), nil
}

func (m *attackManager) ListAttacks() []types.AttackInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]types.AttackInfo, 0, len(m.attacks))
	for _, instance := range m.attacks {
		attack := instance.attack
		config := attack.Config()

		infos = append(infos, types.AttackInfo{
			ID:       attack.ID(),
			Name:     config.Name,
			Type:     config.Type,
			Sequence: config.Sequence,
			Round:    config.Round,
			Status:   instance.status,
			Targets:  config.GetTargetAddresses(),
		})
	}

	return infos
}

func (m *attackManager) StopAttack(attackID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.attacks[attackID]
	if !exists {
		return fmt.Errorf("attack not found: %s", attackID)
	}

	instance.status = types.AttackStatusStopped
	log.Info("Attack stopped", "id", attackID)

	return nil
}

func (m *attackManager) Start() error {
	log.Info("Starting attack manager", "attacks", len(m.attacks))

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, instance := range m.attacks {
		if instance.status == types.AttackStatusPending {
			instance.status = types.AttackStatusActive
		}
	}

	return nil
}

func (m *attackManager) Stop() error {
	log.Info("Stopping attack manager")

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, instance := range m.attacks {
		instance.status = types.AttackStatusStopped
	}

	return nil
}

// ConfigureSilentMessage configures a silent message attack
func (m *attackManager) ConfigureSilentMessage(sequence, round, code, direction uint64, targets []common.Address) error {
	log.Info("ConfigureSilentMessage",
		"sequence", sequence,
		"round", round,
		"code", code,
		"direction", direction,
		"targets", targets)

	// Create attack config
	attackConfig := &types.AttackConfig{
		Name:     fmt.Sprintf("SilentMessage_%d_%d", sequence, round),
		Type:     types.AttackTypeSilent,
		Enabled:  true,
		Sequence: sequence,
		Round:    round,
		Params: map[string]interface{}{
			"code":      code,
			"direction": direction,
			"targets":   targets,
		},
	}

	// Create and register attack
	attack, err := DefaultRegistry.CreateAttack(attackConfig)
	if err != nil {
		return fmt.Errorf("failed to create silent attack: %w", err)
	}

	// TODO:
	// should refactoring
	err = m.Register(attack)
	if err != nil {
		return err
	}

	// Register hook with interceptor
	//hook := interceptor.NewAttackHook(
	//	fmt.Sprintf("silent_%d", uid),
	//	100, // High priority
	//	attack,
	//	m.logger,
	//)
	//
	//if err := m.interceptor.RegisterHook(hook); err != nil {
	//	m.UnregisterAttack(uid)
	//	return fmt.Errorf("failed to register hook: %w", err)
	//}

	log.Info("Configured silent message attack",
		//"uid", uid,
		"sequence", sequence,
		"round", round,
		"code", code)

	return nil
}

// ConfigureTamperedMessage configures a tampered message attack
func (m *attackManager) ConfigureTamperedMessage(params types.TamperMessageParams) error {
	// Convert tamper fields
	tamperFields := make([]map[string]interface{}, len(params.TamperFields))
	for i, field := range params.TamperFields {
		tamperFields[i] = map[string]interface{}{
			"target": field.Target,
			"value":  field.Value,
		}
	}

	// Create attack config
	attackConfig := &types.AttackConfig{
		Name:     fmt.Sprintf("TamperMessage_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeTamper,
		Enabled:  true,
		Sequence: params.Sequence,
		Round:    params.Round,
		Params: map[string]interface{}{
			"code":             params.Code,
			"tamperFields":     tamperFields,
			"withValidMessage": params.WithValidMessage,
			"delay":            params.Delay,
			"targets":          params.Targets,
		},
	}

	// Create and register attack
	// TODO:
	// should refactoring
	attack, err := DefaultRegistry.CreateAttack(attackConfig)
	if err != nil {
		return fmt.Errorf("failed to create tamper attack: %w", err)
	}

	err = m.Register(attack)
	if err != nil {
		return err
	}

	// Register hook with interceptor
	//hook := interceptor.NewAttackHook(
	//	fmt.Sprintf("tamper_%d", uid),
	//	100,
	//	attack,
	//	m.logger,
	//)
	//
	//if err := m.interceptor.RegisterHook(hook); err != nil {
	//	m.UnregisterAttack(uid)
	//	return fmt.Errorf("failed to register hook: %w", err)
	//}

	log.Info("Configured tampered message attack",
		//"uid", uid,
		"sequence", params.Sequence,
		"round", params.Round)

	return nil
}

// ConfigureFakeMessage configures a fake message attack
func (m *attackManager) ConfigureFakeMessage(params types.FakeMessageParams) error {
	// Create attack config
	attackConfig := &types.AttackConfig{
		Name:     fmt.Sprintf("FakeMessage_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeFake,
		Enabled:  true,
		Sequence: params.Sequence,
		Round:    params.Round,
		Params: map[string]interface{}{
			"code":        params.Code,
			"fakeMessage": params.FakeMessage,
			"targets":     params.Targets,
		},
	}

	// Create and register attack
	// TODO:
	// should refactoring
	attack, err := DefaultRegistry.CreateAttack(attackConfig)
	if err != nil {
		return fmt.Errorf("failed to create fake attack: %w", err)
	}

	err = m.Register(attack)
	if err != nil {
		return err
	}

	// Register hook with interceptor
	//hook := interceptor.NewAttackHook(
	//	fmt.Sprintf("fake_%d", uid),
	//	100,
	//	attack,
	//	m.logger,
	//)
	//
	//if err := m.interceptor.RegisterHook(hook); err != nil {
	//	m.UnregisterAttack(uid)
	//	return fmt.Errorf("failed to register hook: %w", err)
	//}

	log.Info("Configured fake message attack",
		//"uid", uid,
		"sequence", params.Sequence,
		"round", params.Round)

	return nil
}

// ConfigureOmitMessage configures an omit message attack
func (m *attackManager) ConfigureOmitMessage(params types.OmitMessageParams) error {
	// Create attack config
	attackConfig := &types.AttackConfig{
		Name:     fmt.Sprintf("OmitMessage_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeOmit,
		Enabled:  true,
		Sequence: params.Sequence,
		Round:    params.Round,
		Params: map[string]interface{}{
			"code":    params.Code,
			"cmd":     params.Cmd,
			"cnt":     params.Cnt,
			"targets": params.Targets,
		},
	}

	// Create and register attack
	// TODO:
	// should refactoring
	attack, err := DefaultRegistry.CreateAttack(attackConfig)
	if err != nil {
		return fmt.Errorf("failed to create omit attack: %w", err)
	}

	err = m.Register(attack)
	if err != nil {
		return err
	}

	// Register hook with interceptor
	//hook := interceptor.NewAttackHook(
	//	fmt.Sprintf("omit_%d", uid),
	//	100,
	//	attack,
	//	m.logger,
	//)
	//
	//if err := m.interceptor.RegisterHook(hook); err != nil {
	//	m.UnregisterAttack(uid)
	//	return fmt.Errorf("failed to register hook: %w", err)
	//}

	log.Info("Configured omit message attack",
		//"uid", uid,
		"sequence", params.Sequence,
		"round", params.Round)

	return nil
}

// ConfigureRoleSpoofedMessage configures a role spoofing attack
func (m *attackManager) ConfigureRoleSpoofedMessage(params types.RoleSpoofParams) error {
	// Create attack config
	attackConfig := &types.AttackConfig{
		Name:     fmt.Sprintf("RoleSpoof_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeRoleSpoof,
		Enabled:  true,
		Sequence: params.Sequence,
		Round:    params.Round,
		Params: map[string]interface{}{
			"code":        params.Code,
			"fakeMessage": params.FakeMessage,
			"targets":     params.Targets,
		},
	}

	// Create and register attack
	// TODO:
	// should refactoring
	attack, err := DefaultRegistry.CreateAttack(attackConfig)
	if err != nil {
		return fmt.Errorf("failed to create role spoof attack: %w", err)
	}

	err = m.Register(attack)
	if err != nil {
		return err
	}

	// Register hook with interceptor
	//hook := interceptor.NewAttackHook(
	//	fmt.Sprintf("rolespoof_%d", uid),
	//	100,
	//	attack,
	//	m.logger,
	//)
	//
	//if err := m.interceptor.RegisterHook(hook); err != nil {
	//	m.UnregisterAttack(uid)
	//	return fmt.Errorf("failed to register hook: %w", err)
	//}

	log.Info("Configured role spoofed message attack",
		//"uid", uid,
		"sequence", params.Sequence,
		"round", params.Round)

	return nil
}

// ConfigureReplayMessage configures a replay attack
func (m *attackManager) ConfigureReplayMessage(params types.ReplayMessageParams) error {
	// Create attack config
	attackConfig := &types.AttackConfig{
		Name:     fmt.Sprintf("ReplayMessage_%d_%d", params.Sequence, params.Round),
		Type:     types.AttackTypeReplay,
		Enabled:  true,
		Sequence: params.Sequence,
		Round:    params.Round,
		Params: map[string]interface{}{
			"ori_sequence":    params.OriSequence,
			"ori_round":       params.OriRound,
			"useOriginalView": params.UseOriginalView,
			"code":            params.Code,
			"targets":         params.Targets,
		},
	}

	// Create and register attack
	attack, err := DefaultRegistry.CreateAttack(attackConfig)
	if err != nil {
		return fmt.Errorf("failed to create replay attack: %w", err)
	}

	// TODO:
	// should refactoring
	err = m.Register(attack)
	if err != nil {
		return err
	}

	// Register hook with interceptor
	//hook := interceptor.NewAttackHook(
	//	fmt.Sprintf("replay_%d", uid),
	//	100,
	//	attack,
	//	m.logger,
	//)
	//
	//if err := m.interceptor.RegisterHook(hook); err != nil {
	//	m.UnregisterAttack(uid)
	//	return fmt.Errorf("failed to register hook: %w", err)
	//}

	log.Info("Configured replay message attack",
		//"uid", uid,
		"sequence", params.Sequence,
		"round", params.Round,
		"originalSequence", params.OriSequence,
		"originalRound", params.OriRound)

	return nil
}
