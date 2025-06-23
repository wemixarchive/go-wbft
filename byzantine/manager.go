package byzantine

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/google/uuid"
)

// attackManager implements AttackManager interface
type attackManager struct {
	eventCollector types.EventCollector
	builder        types.AttackBuilder

	mu      sync.RWMutex
	attacks map[string]*attackInstance
}

// attackInstance wraps an attack with metadata
type attackInstance struct {
	id     string
	attack types.Attack
	status types.AttackStatus
}

var _ types.AttackManager = (*attackManager)(nil)

// NewAttackManager creates a new attack manager
func NewAttackManager(collector types.EventCollector, builder types.AttackBuilder) types.AttackManager {
	return &attackManager{
		eventCollector: collector,
		builder:        builder,
		attacks:        make(map[string]*attackInstance),
	}
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
	m.attacks[attack.ID()] = &attackInstance{
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

	m.attacks[attack.ID()] = &attackInstance{
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

	m.attacks[attack.ID()] = &attackInstance{
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

func (m *attackManager) GetAttacksToExecute(sequence, round uint64, msgCode uint64) []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var attacks []types.Attack
	for _, instance := range m.attacks {
		if instance.status == types.AttackStatusActive &&
			instance.attack.ShouldExecute(sequence, round, msgCode) {
			attacks = append(attacks, instance.attack)
		}
	}

	return attacks
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

// Other methods implementation...
// ConfigureAttack, ListAttacks, StopAttack, GetAttacksToExecute, Start, Stop

// baseAttack is a temporary implementation
//type baseAttack struct {
//	id     string
//	config types.AttackConfig
//}
//
//func (a *baseAttack) ID() string                 { return a.id }
//func (a *baseAttack) Config() types.AttackConfig { return a.config }
//func (a *baseAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
//	return a.config.Sequence == sequence && a.config.Round == round
//}
//func (a *baseAttack) Execute(ctx AttackContext) error {
//	// TODO: Implement actual attack logic
//	return nil
//}
