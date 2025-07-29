# Byzantine Module API Reference

## RPC API

The Byzantine module exposes RPC methods under the `byzantine` namespace. These methods are available when the Byzantine module is enabled.

### byzantine_getStatus

Returns the current status of the Byzantine service.

#### Parameters
None

#### Returns
```typescript
{
  running: boolean             // Whether the service is running
  started_at: string           // ISO timestamp when service started
  active_attacks: number       // Number of currently active attacks
  executed_attacks: number     // Number of executed attacks
  failed_attacks: number       // Number of failed attacks
  stored_messages: number      // Number of messages in storage
}
```

#### Example
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"byzantine_getStatus","params":[],"id":1}' \
  http://localhost:8545
```

### byzantine_listAttacks

Lists all registered attacks with their current status.

#### Parameters
None

#### Returns
```typescript
Array<{
  uid: number                   // Unique identifier
  name: string                  // Attack name
  type: string                  // Attack type
  enabled: boolean              // Whether attack is enabled
  sequence: number              // Target sequence number
  round: number                 // Target round number
  code: number                  // Message code (optional)
  status: string                // Current status
  targets: string[]             // Target addresses (optional)
  parameters: object            // Attack-specific parameters
  created_at: string            // Creation timestamp
  executed_at: string           // Execution timestamp (optional)
}>
```

#### Example
```javascript
const attacks = await byzantine.listAttacks();
console.log(attacks);
```

### byzantine_registerAttack

Registers a new Byzantine attack.

#### Parameters
```typescript
{
  name: string                  // Attack name
  type: string                  // Attack type
  sequence: number              // Target sequence number
  round: number                 // Target round number
  code: number                  // Message code (optional)
  targets: string[]             // Target addresses (optional)
  parameters: object            // Attack-specific parameters
}
```

#### Returns
```typescript
number  // Attack UID
```

#### Example
```javascript
const uid = await byzantine.registerAttack({
  name: "test_silent_attack",
  type: "silentMessage",
  sequence: 100,
  round: 0,
  code: 1,
  parameters: {
    direction: 1
  }
});
```

### byzantine_cancelAttack

Cancels an active attack.

#### Parameters
```typescript
uid: number  // Attack UID to cancel
```

#### Returns
```typescript
boolean  // Success status
```

#### Example
```javascript
const success = await byzantine.cancelAttack(12345);
```

### byzantine_getAttackHistory

Retrieves the execution history for a specific attack.

#### Parameters
```typescript
uid: number  // Attack UID
```

#### Returns
```typescript
Array<{
  uid: number                    // Attack UID
  success: boolean               // Whether execution succeeded
  error?: string                 // Error message if failed
  details?: object               // Execution details
  executed_at: string           // Execution timestamp
  duration: number              // Execution duration in nanoseconds
  block_message?: boolean       // Whether message was blocked
  block_reason?: string         // Reason for blocking
}>
```

#### Example
```javascript
const history = await byzantine.getAttackHistory(12345);
```

### byzantine_getMetrics

Returns service metrics.

#### Parameters
None

#### Returns
```typescript
{
  attacks_registered: number     // Total attacks registered
  attacks_executed: number       // Total attacks executed
  attacks_failed: number         // Total attacks failed
  messages_stored: number        // Total messages stored
  events_processed: number       // Total events processed
  storage_size_bytes: number     // Storage size in bytes
  uptime_seconds: number         // Service uptime in seconds
}
```

#### Example
```javascript
const metrics = await byzantine.getMetrics();
```

## Attack Types Reference

### silentMessage

Drops messages without sending them.

#### Parameters
```typescript
{
  direction: number    // 1: send only, 2: receive only, 3: both
  targets?: string[]   // Optional target addresses
}
```

### doubleVote

Sends duplicate messages with potentially different content.

#### Parameters
```typescript
{
  targets?: string[]          // Optional target addresses
}
```

### tamperedMessage

Modifies message fields before sending.

#### Parameters
```typescript
{
  fields: Array<{
    target: string           // Field path (e.g., "Proposal.Header.Coinbase")
    value: any              // New value
  }>
  targets?: string[]        // Optional target addresses
}
```

### fakeMessage

Creates and sends fake messages.

#### Parameters
```typescript
{
  fields?: Array<{
    target: string          // Target field (e.g., "Tx.Count")
    value: any             // Value for the field
  }>
  targets?: string[]        // Optional target addresses
}
```

### omitMessage

Omits required fields from messages.

#### Parameters
```typescript
{
  cmd: number              // What to omit (depends on message type)
  cnt?: number             // Number of items to omit (0 = all)
  targets?: string[]       // Optional target addresses
}
```

#### Command values for PrePrepare:
- 1: Omit previous Prepare seals
- 2: Omit previous Commit seals

#### Command values for RoundChange-PrePrepare:
- 1: Omit RoundChangeMessages
- 2: Omit PrepareMessages

### roleSpoofed

Sends messages while impersonating a different role.

#### Parameters
```typescript
{
  fields?: Array<{
    target: string          // Target field
    value: any             // Value for the field
  }>
  targets?: string[]        // Optional target addresses
}
```

### replayAttack

Replays previously sent messages.

#### Parameters
```typescript
{
  ori_sequence: number      // Original message sequence
  ori_round: number        // Original message round
  useOriginalView: boolean // Use original sequence/round
  targets?: string[]       // Optional target addresses
}
```

### messageFlooding

Floods the network with excessive messages.

#### Parameters
```typescript
{
  rate: number             // Messages per second
  duration: number         // Attack duration (seconds)
  messageSize?: number     // Size of each message (bytes)
  targets?: string[]       // Optional target addresses
}
```

## Message Codes

```typescript
enum MessageCode {
  PrePrepare = 0x12,      // 18 in decimal
  Prepare = 0x13,         // 19 in decimal
  Commit = 0x14,          // 20 in decimal
  RoundChange = 0x15      // 21 in decimal
}
```

## Attack Status Values

```typescript
enum AttackStatus {
  Pending = "pending",       // Waiting for conditions
  Active = "active",         // Ready to execute
  Executed = "executed",     // Has been executed
  Completed = "completed",   // All executions done
  Failed = "failed",         // Execution failed
  Cancelled = "cancelled"    // Manually cancelled
}
```

## CLI Commands

### byzantine list

Lists all registered attacks.

```bash
geth byzantine list [--json]
```

### byzantine register

Registers a new attack from command line.

```bash
geth byzantine register \
  --name "test_attack" \
  --type "silentMessage" \
  --sequence 100 \
  --round 0 \
  --code 1
```

### byzantine cancel

Cancels an active attack.

```bash
geth byzantine cancel --uid 12345
```

### byzantine status

Shows Byzantine service status.

```bash
geth byzantine status
```

### byzantine history

Shows attack execution history.

```bash
geth byzantine history [--uid 12345] [--limit 100]
```

## Configuration File Reference

### Full Configuration Example

```json
{
  "enabled": true,
  "config_file": "byzantine-config.json",
  "attacks": [
    {
      "name": "silent_proposer_attack",
      "type": "silentMessage",
      "enabled": true,
      "sequence": 100,
      "round": 0,
      "code": 1,
      "parameters": {
        "direction": 1
      }
    },
    {
      "name": "double_vote_attack",
      "type": "doubleVote",
      "enabled": true,
      "sequence": 200,
      "round": 1,
      "code": 2,
      "parameters": {
      },
      "targets": ["0x1234..."]
    }
  ],
  "storage": {
    "max_messages": 10000,
    "retention_period": "24h",
    "prune_interval": "1h"
  },
  "monitoring": {
    "enabled": true,
    "metrics_interval": "30s",
    "export_metrics": true,
    "metrics_endpoint": "http://localhost:9090"
  }
}
```

### Storage Configuration

```typescript
{
  max_messages: number        // Maximum messages to store
  retention_period: string    // How long to keep messages (e.g., "24h")
  prune_interval: string      // How often to prune old messages
}
```

### Monitoring Configuration

```typescript
{
  enabled: boolean            // Enable monitoring
  metrics_interval: string    // Metrics collection interval
  export_metrics: boolean     // Export metrics to external system
  metrics_endpoint?: string   // External metrics endpoint
}
```

## Error Codes

| Code | Description |
|------|-------------|
| -32000 | Generic Byzantine error |
| -32001 | Service not running |
| -32002 | Attack not found |
| -32003 | Invalid attack configuration |
| -32004 | Attack already exists |
| -32005 | Storage error |
| -32006 | Invalid parameters |
| -32007 | Operation timeout |

## Best Practices

1. **Always specify sequence and round** for precise attack timing
2. **Use meaningful attack names** for easy identification
3. **Set appropriate storage limits** to prevent memory issues
4. **Monitor metrics** to track attack impact
5. **Cancel completed attacks** to free resources
6. **Test attacks in isolation** before combining
7. **Document attack parameters** for reproducibility
8. **Use target addresses** to limit attack scope
