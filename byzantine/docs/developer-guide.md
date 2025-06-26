# Byzantine Module Developer Guide

## Introduction

This guide provides instructions for developers working with the Byzantine module in the go-wemix-qbft project. The Byzantine module is designed to test the robustness of the QBFT consensus algorithm by simulating various Byzantine behaviors.

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Understanding of QBFT consensus algorithm
- Familiarity with Ethereum/Geth architecture

### Building with Byzantine Module

```bash
# Build geth with Byzantine module
go build -o geth ./cmd/geth

# Run with Byzantine enabled
./geth --byzantine.enabled --byzantine.config byzantine-config.json
```

## Configuration

### Command Line Flags

```bash
# Enable Byzantine module
--byzantine.enabled

# Specify configuration file
--byzantine.config <path>

# Enable monitoring
--byzantine.monitoring.enabled

# Set storage limits
--byzantine.storage.max-messages 10000
--byzantine.storage.retention-period 24h
```

### Configuration File Example

```json
{
  "enabled": true,
  "attacks": [
    {
      "name": "test_silent_attack",
      "type": "silentMessage",
      "enabled": true,
      "sequence": 100,
      "round": 0,
      "code": 1,
      "parameters": {
        "direction": 1,
        "targets": ["0x1234..."]
      }
    }
  ],
  "storage": {
    "max_messages": 10000,
    "retention_period": "24h",
    "prune_interval": "1h"
  },
  "monitoring": {
    "enabled": true,
    "metrics_interval": "30s"
  }
}
```

## API Usage

### RPC Methods

#### 1. Get Service Status
```javascript
// Request
{
  "jsonrpc": "2.0",
  "method": "byzantine_getStatus",
  "params": [],
  "id": 1
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "running": true,
    "started_at": "2024-01-01T10:00:00Z",
    "active_attacks": 2,
    "executed_attacks": 10,
    "failed_attacks": 1,
    "stored_messages": 5000
  }
}
```

#### 2. Register Attack
```javascript
// Request
{
  "jsonrpc": "2.0",
  "method": "byzantine_registerAttack",
  "params": [{
    "name": "test_attack",
    "type": "silentMessage",
    "sequence": 100,
    "round": 0,
    "code": 1,
    "parameters": {
      "direction": 1
    }
  }],
  "id": 2
}

// Response
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": 12345  // Attack UID
}
```

#### 3. List Attacks
```javascript
// Request
{
  "jsonrpc": "2.0",
  "method": "byzantine_listAttacks",
  "params": [],
  "id": 3
}

// Response
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": [
    {
      "uid": 12345,
      "name": "test_attack",
      "type": "silentMessage",
      "status": "active",
      "sequence": 100,
      "round": 0
    }
  ]
}
```

## Implementing New Attacks

### Step 1: Define Attack Structure

```go
package attacks

import (
    "context"
    "github.com/ethereum/go-ethereum/byzantine/registry"
    "github.com/ethereum/go-ethereum/byzantine/types"
)

type MyCustomAttack struct {
    *registry.BaseAttack
    // Add custom fields
    customParam string
}
```

### Step 2: Implement Attack Interface

```go
// CheckExecuteCondition determines if attack should execute
func (a *MyCustomAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
    config := a.GetConfig()
    
    // Check sequence and round
    if event.Sequence != config.Sequence || event.Round != config.Round {
        return false
    }
    
    // Check attack status
    if config.Status == types.AttackStatusCancelled {
        return false
    }
    
    // Add custom conditions
    return true
}

// Execute performs the attack
func (a *MyCustomAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
    // Implement attack logic
    
    result := &types.AttackResult{
        UID:        a.GetUID(),
        Success:    true,
        ExecutedAt: time.Now(),
        Details: map[string]interface{}{
            "action": "custom_attack_executed",
        },
    }
    
    return result, nil
}
```

### Step 3: Create Factory Function

```go
func MyCustomAttackFactory(config types.AttackConfig) (types.Attack, error) {
    // Parse parameters
    customParam := registry.GetStringParameter(config, "customParam", "default")
    
    return &MyCustomAttack{
        BaseAttack:  registry.NewBaseAttack(config),
        customParam: customParam,
    }, nil
}
```

### Step 4: Register Attack Type

```go
func init() {
    err := registry.Register("myCustomAttack", MyCustomAttackFactory)
    if err != nil {
        panic(err)
    }
}
```

## Attack Development Best Practices

### 1. Condition Checking
- Always check sequence and round first
- Verify attack status (not cancelled/completed)
- Validate event type matches expected type
- Check role constraints if applicable

### 2. Error Handling
- Return meaningful error messages
- Log important decisions and errors
- Handle context cancellation properly
- Clean up resources in case of failure

### 3. Performance Considerations
- Avoid blocking operations in Execute()
- Use context timeout for long operations
- Minimize memory allocation
- Cache frequently used data

### 4. Testing
```go
func TestMyCustomAttack(t *testing.T) {
    // Create test configuration
    config := types.AttackConfig{
        UID:      1,
        Name:     "test",
        Type:     "myCustomAttack",
        Sequence: 100,
        Round:    0,
        Parameters: map[string]interface{}{
            "customParam": "test_value",
        },
    }
    
    // Create attack instance
    attack, err := MyCustomAttackFactory(config)
    require.NoError(t, err)
    
    // Create test event
    event := types.Event{
        Type:     types.EventTypeMessageSent,
        Sequence: 100,
        Round:    0,
    }
    
    // Test condition check
    assert.True(t, attack.CheckExecuteCondition(context.Background(), event))
    
    // Test execution
    result, err := attack.Execute(context.Background(), event)
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Integration with QBFT

### Understanding Hook Points

1. **BeforeBroadcast**: Called before sending any consensus message
   - Can block outbound messages
   - Access to message code, sequence, round, and sender

2. **BeforeProcessMessage**: Called before processing received messages
   - Can block inbound messages
   - Access to message code, sequence, round, and sender

### Message Codes
```go
const (
    PrePrepareCode  = 0x12
    PrepareCode     = 0x13
    CommitCode      = 0x14
    RoundChangeCode = 0x15
)
```

### Attack Examples by Category

#### Safety Attacks
- **Double Vote**: Send duplicate messages with different content
- **Conflicting Blocks**: Propose different blocks for same height

#### Liveness Attacks
- **Silent Proposer**: Proposer doesn't send PrePrepare
- **Silent Validator**: Validators don't send Prepare/Commit

#### Integrity Attacks
- **Tampered Header**: Modify block header fields
- **Fake Transactions**: Inject invalid transactions
- **Omit Seals**: Remove required signatures

#### Role Attacks
- **Fake Proposer**: Non-proposer sends PrePrepare
- **Role Spoofing**: Impersonate other validators

## Debugging

### Enable Debug Logging
```bash
./geth --byzantine.enabled --log.level=debug
```

### Common Log Patterns
```
# Attack registration
INFO [01-01|10:00:00.000] Byzantine attack registered uid=12345 type=silentMessage

# Attack execution
DEBUG [01-01|10:01:00.000] Byzantine attack condition met uid=12345 seq=100 round=0
INFO [01-01|10:01:00.001] Byzantine attack executed uid=12345 result=success

# Message blocking
INFO [01-01|10:01:00.002] Byzantine: Blocking outbound message attack_type=silentMessage reason="Silent attack"
```

### Monitoring Metrics
- Attack execution count
- Message block count
- Storage usage
- Event processing rate

## Troubleshooting

### Common Issues

1. **Attack Not Executing**
   - Check sequence/round matches
   - Verify attack is enabled and active
   - Ensure message code matches
   - Check target addresses if specified

2. **Performance Impact**
   - Monitor storage size
   - Check event processing rate
   - Review attack execution time
   - Enable pruning if needed

3. **Integration Issues**
   - Verify Byzantine hook is registered
   - Check consensus engine type (must be WBFT)
   - Ensure service is started
   - Review log for registration errors

## Security Considerations

1. **Production Usage**
   - Byzantine module should NEVER be enabled in production
   - Restrict API access in test environments
   - Monitor for unauthorized attack registration

2. **Resource Limits**
   - Set appropriate storage limits
   - Configure message retention period
   - Implement rate limiting for flooding attacks

3. **Attack Validation**
   - Validate all attack parameters
   - Check target address validity
   - Verify sequence/round ranges
   - Sanitize custom parameters

## Advanced Topics

### Custom Event Types
```go
// Define custom event
type CustomEvent struct {
    MyField string
}

// Publish event
event := types.Event{
    Type: "customEvent",
    Data: &CustomEvent{MyField: "value"},
}
publisher.Publish(event)
```

### Storage Extension
```go
// Implement custom storage
type MyStorage struct {
    // Implementation
}

func (s *MyStorage) Store(message *types.StoredMessage) error {
    // Custom storage logic
}
```

### Attack Chaining
```go
// Configure dependent attacks
{
  "attacks": [
    {
      "uid": 1,
      "name": "setup_attack",
      "type": "silentMessage",
      "sequence": 100
    },
    {
      "uid": 2,
      "name": "follow_up_attack",
      "type": "fakeMessage",
      "sequence": 101,
      "parameters": {
        "dependsOn": 1
      }
    }
  ]
}
```

## Conclusion

The Byzantine module provides a powerful framework for testing QBFT consensus robustness. 
By following this guide and best practices, developers can effectively implement and test various Byzantine behaviors 
to ensure the consensus algorithm's reliability and security.
