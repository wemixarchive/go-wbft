# Byzantine Module Code Structure

## Package Overview

### 1. **byzantine** (Root Package)
```
byzantine/
├── register.go           # Module registration and integration
├── web3ext/             # Web3 extension for RPC
│   └── web3ext.go       # JavaScript bindings for Byzantine API
└── tests/               # Integration tests
    └── example_config.json
```

### 2. **byzantine/types**
Core type definitions and interfaces:
```
types/
├── interface.go         # Core interfaces (Attack, AttackManager, etc.)
├── attack_types.go      # Attack type definitions and enums
├── message_types.go     # QBFT message type definitions
├── event_types.go       # Event system types
├── config_type.go       # Configuration structures
├── api_types.go         # API request/response types
├── wbft_types.go        # WBFT integration types
├── constants.go         # Module constants
└── errors.go            # Error definitions
```

### 3. **byzantine/service**
Core service implementation:
```
service/
├── byzantine_service.go  # Main service implementation
├── consensus_hook.go     # ConsensusHook implementation
└── config_loader.go      # Configuration loading logic
```

### 4. **byzantine/attacks**
Attack implementations:
```
attacks/
├── registry.go           # Attack registration
├── silent_attack.go      # Silent message dropping
├── double_vote_attack.go # Double voting attack
├── tamper_attack.go      # Message tampering
├── fake_attack.go        # Fake message generation
├── omit_attack.go        # Message field omission
├── rolespoof_attack.go   # Role spoofing
├── replay_attack.go      # Message replay
└── msg_flooding_attack.go # Message flooding DDoS
```

### 5. **byzantine/storage**
Storage implementations:
```
storage/
├── message_storage.go       # Message storage implementation
├── message_storage_test.go  # Storage tests
└── history_storage.go       # Attack history storage
```

### 6. **byzantine/manager**
Attack management:
```
manager/
├── attack_manager.go     # Attack orchestration
└── chain_handler.go      # Blockchain state handling
```

### 7. **byzantine/adapter**
Integration adapters:
```
adapter/
├── hook_adapter.go       # Consensus hook adapter
└── event_observer.go     # Event observation/publishing
```

### 8. **byzantine/api**
RPC API implementation:
```
api/
├── api.go               # Main API implementation
├── handler.go           # HTTP/WebSocket handlers
├── models.go            # API data models
├── converter.go         # Type conversion utilities
└── api_test.go         # API tests
```

### 9. **byzantine/cmd**
Command-line interface:
```
cmd/
├── flags.go             # CLI flag definitions
├── commands.go          # CLI command definitions
├── actions.go           # Command implementations
├── config.go            # Configuration handling
└── testdata/
    └── byzantine_config.json
```

### 10. **byzantine/registry**
Attack registry:
```
registry/
├── attack_registry.go    # Attack registration system
└── factory.go           # Attack factory methods
```

## Key Interfaces

### Attack Interface
```go
type Attack interface {
    GetUID() uint64
    GetType() AttackType
    CheckExecuteCondition(ctx context.Context, event Event) bool
    Execute(ctx context.Context, event Event) (*AttackResult, error)
    GetConfig() AttackConfig
    SetStatus(status AttackStatus)
}
```

### AttackManager Interface
```go
type AttackManager interface {
    RegisterAttack(attack Attack) error
    UnregisterAttack(uid uint64) error
    GetAttack(uid uint64) (Attack, error)
    ListAttacks() []Attack
    ProcessEvent(ctx context.Context, event Event) error
    GetActiveAttacks() []Attack
    EvaluateAndExecuteAttacks(ctx context.Context, event Event) (AttackDecision, error)
    ProcessEventAsync(ctx context.Context, event Event) error
}
```

### ConsensusHook Interface
```go
type ConsensusHook interface {
    BeforeBroadcast(msgCode, sequence, round uint64, from common.Address) bool
    BeforeProcessMessage(msgCode, sequence, round uint64, from common.Address) bool
}
```

### ByzantineService Interface
```go
type ByzantineService interface {
    Start() error
    Stop() error
    GetStatus() ServiceStatus
    Configure(config ByzantineConfig) error
    GetMetrics() Metrics
    RegisterAttack(config AttackConfig) (uint64, error)
    CancelAttack(uid uint64) error
    ListAttacks() []AttackConfig
    GetAttackHistory(uid uint64) ([]AttackResult, error)
    GetAttackManager() AttackManager
    GetMessageStorage() MessageStorage
    GetHistoryStorage() HistoryStorage
    GetConsensusHook() ConsensusHook
}
```

## Key Data Structures

### AttackConfig
```go
type AttackConfig struct {
    UID        uint64                 `json:"uid"`
    Name       string                 `json:"name"`
    Type       AttackType             `json:"type"`
    Enabled    bool                   `json:"enabled"`
    Sequence   uint64                 `json:"sequence"`
    Round      uint64                 `json:"round"`
    Code       MessageCode            `json:"code,omitempty"`
    Status     AttackStatus           `json:"status,omitempty"`
    Targets    []common.Address       `json:"targets,omitempty"`
    Parameters map[string]interface{} `json:"parameters,omitempty"`
    CreatedAt  time.Time              `json:"created_at"`
    ExecutedAt *time.Time             `json:"executed_at,omitempty"`
}
```

### Event
```go
type Event struct {
    Type      EventType
    Sequence  uint64
    Round     uint64
    Timestamp time.Time
    Data      interface{}
    Metadata  map[string]interface{}
}
```

### QBFTMessage
```go
type QBFTMessage struct {
    Code      MessageCode
    Address   common.Address
    Signature []byte
    View      *View
    Payload   interface{}
}
```

## Configuration Structure

### ByzantineConfig
```go
type ByzantineConfig struct {
    Enabled       bool            `json:"enabled"`
    ConfigFile    string          `json:"config_file,omitempty"`
    Attacks       []AttackConfig  `json:"attacks,omitempty"`
    StorageConfig StorageConfig   `json:"storage,omitempty"`
    Monitoring    MonitoringConfig `json:"monitoring,omitempty"`
}
```

### Example Configuration
```json
{
  "enabled": true,
  "attacks": [
    {
      "name": "silent_proposer",
      "type": "silentMessage",
      "enabled": true,
      "sequence": 100,
      "round": 0,
      "code": 1,
      "parameters": {
        "direction": 1
      }
    }
  ],
  "storage": {
    "max_messages": 10000,
    "retention_period": "24h",
    "prune_interval": "1h"
  }
}
```

## Message Codes

```go
const (
    MessageCodePrePrepare     MessageCode = 0x12
    MessageCodePrepare        MessageCode = 0x13
    MessageCodeCommit         MessageCode = 0x14
    MessageCodeRoundChange    MessageCode = 0x15
)
```

## Attack Types

```go
const (
    AttackTypeDoubleVote      AttackType = "doubleVote"
    AttackTypeSilentMessage   AttackType = "silentMessage"
    AttackTypeTamperedMessage AttackType = "tamperedMessage"
    AttackTypeFakeMessage     AttackType = "fakeMessage"
    AttackTypeOmitMessage     AttackType = "omitMessage"
    AttackTypeRoleSpoofed     AttackType = "roleSpoofed"
    AttackTypeReplay          AttackType = "replayAttack"
    AttackTypeMessageFlooding AttackType = "messageFlooding"
)
```

## RPC API Methods

### Byzantine Namespace
- `byzantine_getStatus`: Get service status
- `byzantine_listAttacks`: List all registered attacks
- `byzantine_registerAttack`: Register a new attack
- `byzantine_cancelAttack`: Cancel an active attack
- `byzantine_getAttackHistory`: Get attack execution history
- `byzantine_getMetrics`: Get service metrics

## CLI Commands

### Byzantine Command
```bash
geth byzantine [subcommand] [flags]
```

Subcommands:
- `list`: List all attacks
- `register`: Register a new attack
- `cancel`: Cancel an attack
- `status`: Show service status
- `history`: Show attack history

## Lifecycle Management

1. **Service Start**:
   - Load configuration
   - Initialize components
   - Subscribe to events
   - Load attacks from config

2. **Service Stop**:
   - Cancel context
   - Clean up resources
   - Update status

3. **Attack Lifecycle**:
   - Registration → Pending
   - Activation → Active
   - Execution → Executed
   - Completion → Completed/Failed
