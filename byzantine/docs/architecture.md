# Byzantine Module Architecture

## Overview

The Byzantine module is a sophisticated testing framework integrated into the go-wemix-qbft project to simulate 
and test Byzantine behaviors in the QBFT consensus algorithm. It provides a modular, extensible architecture for 
implementing various Byzantine attack scenarios.

## Module Structure

### Core Components

#### 1. **Service Layer** (`byzantine/service/`)
- **ByzantineService**: Main service implementing lifecycle management and orchestration
- **ConsensusHook**: Integration point with the consensus engine (WBFT)
- **ConfigLoader**: Configuration management for Byzantine attacks

#### 2. **Types Layer** (`byzantine/types/`)
- **Interface Definitions**: Core interfaces (Attack, AttackManager, MessageStorage, etc.)
- **Attack Types**: Enumeration of Byzantine attack types
- **Message Types**: QBFT message type definitions
- **Event Types**: Event system for attack coordination
- **Configuration Types**: Configuration structures

#### 3. **Attack Implementations** (`byzantine/attacks/`)
- **Silent Attack**: Drops messages silently
- **Double Vote Attack**: Sends duplicate messages
- **Tamper Attack**: Modifies message content
- **Fake Attack**: Creates fake messages
- **Omit Attack**: Omits parts of messages
- **Role Spoof Attack**: Impersonates other roles
- **Replay Attack**: Replays old messages
- **Message Flooding Attack**: Floods network with messages

#### 4. **Storage Layer** (`byzantine/storage/`)
- **MessageStorage**: Stores and retrieves consensus messages
- **HistoryStorage**: Maintains attack execution history

#### 5. **Management Layer** (`byzantine/manager/`)
- **AttackManager**: Orchestrates attack execution
- **ChainHandler**: Handles blockchain state for attacks

#### 6. **Adapter Layer** (`byzantine/adapter/`)
- **HookAdapter**: Adapts Byzantine hooks to consensus hooks
- **EventObserver**: Observes and publishes consensus events

#### 7. **API Layer** (`byzantine/api/`)
- **PublicByzantineAPI**: RPC API for Byzantine operations
- **Handler**: HTTP/WebSocket handlers
- **Models**: API data models

#### 8. **Command Layer** (`byzantine/cmd/`)
- **Flags**: CLI flags for Byzantine configuration
- **Commands**: CLI commands for Byzantine operations
- **Actions**: Command implementations
- **Config**: Configuration parsing

#### 9. **Registry Layer** (`byzantine/registry/`)
- **AttackRegistry**: Registry for attack implementations
- **Factory**: Factory pattern for attack creation

## Integration with Geth

### 1. **Main Integration** (`cmd/geth/main.go`)
```go
// Register Byzantine flags
byzantine.RegisterFlags(app)

// Register Byzantine commands  
byzantine.RegisterByzantineCommands(app)
```

### 2. **Configuration Integration** (`cmd/geth/config.go`)
```go
// Register Byzantine service during node creation
if err := byzantine.Register(ctx, stack, backend, eth); err != nil {
    utils.Fatalf("Failed to register Byzantine service: %v", err)
}
```

### 3. **Registration Process** (`byzantine/register.go`)
- Loads Byzantine configuration
- Creates ByzantineService instance
- Registers service lifecycle with node
- Registers RPC APIs
- Integrates with consensus engine (WBFT)

## Integration with Consensus (WBFT)

### 1. **Backend Integration** (`consensus/wbft/backend/backend.go`)
```go
// Byzantine hook integration
func (sb *Backend) SetByzantineHook(hook btypes.ConsensusHook) {
    sb.byzantineHook = hook
}

// Message broadcast interception
if sb.byzantineHook != nil {
    if !sb.byzantineHook.BeforeBroadcast(code, sequence, round, sb.address) {
        return nil // Silent drop
    }
}
```

### 2. **Handler Integration** (`consensus/wbft/backend/handler.go`)
```go
// Message reception interception
if sb.byzantineHook != nil {
    if !sb.byzantineHook.BeforeProcessMessage(msg.Code, sequence, round, addr) {
        return true, nil // Message processed (dropped)
    }
}
```

## Message Flow

### Outbound Messages (Broadcasting)
1. Consensus engine prepares to broadcast a message
2. `BeforeBroadcast` hook is called
3. Byzantine module evaluates active attacks
4. Decision made to send or drop message
5. If allowed, message is broadcast normally

### Inbound Messages (Receiving)
1. Node receives a message from the network
2. `BeforeProcessMessage` hook is called
3. Byzantine module evaluates active attacks
4. Decision made to process or drop message
5. If allowed, message is processed normally

## Attack Execution Flow

1. **Registration**: Attack is registered via API or configuration
2. **Event Monitoring**: ConsensusHook monitors consensus events
3. **Condition Check**: Attack checks if execution conditions are met
4. **Execution**: Attack modifies behavior (drop, modify, or inject messages)
5. **Result Logging**: Attack results are stored in history

## Key Design Patterns

### 1. **Strategy Pattern**
- Different attack implementations share common interface
- Runtime selection of attack behavior

### 2. **Observer Pattern**
- Event system for monitoring consensus state
- Publishers and subscribers for event handling

### 3. **Factory Pattern**
- Attack registry with factory methods
- Dynamic attack instantiation

### 4. **Hook Pattern**
- Non-intrusive integration with consensus
- Minimal changes to legacy code

### 5. **Lifecycle Pattern**
- Service lifecycle management (Start/Stop)
- Clean resource management

## Configuration

Byzantine behavior is configured through:
1. **CLI Flags**: Command-line arguments
2. **JSON Configuration**: File-based configuration
3. **RPC API**: Runtime configuration via API

## Security Considerations

- Byzantine module is intended for testing only
- Should be disabled in production environments
- Access to Byzantine API should be restricted
- Attack configurations should be carefully validated

## Extension Points

The module is designed for easy extension:
1. **New Attack Types**: Implement Attack interface
2. **New Storage Backends**: Implement Storage interfaces
3. **New Event Types**: Extend event system
4. **Custom Hooks**: Add new consensus integration points
