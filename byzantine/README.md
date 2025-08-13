# Byzantine Module Documentation

## Overview

The Byzantine module is a comprehensive testing framework for simulating Byzantine behaviors in the WBFT (WEMIX Byzantine Fault Tolerant) consensus system. It provides a modular and extensible architecture for implementing various attack scenarios to test the robustness of the WBFT consensus algorithm.

## Documentation Structure

### 1. [Architecture Overview](docs/architecture.md)
- High-level system design
- Component relationships
- Integration with Geth and WBFT
- Key design patterns
- Security considerations

### 2. [Code Structure](docs/code-structure.md)
- Detailed package organization
- Interface definitions
- Data structures
- Configuration format
- API specifications

### 3. [Diagrams](docs/diagrams.md)
- Module architecture diagram
- Message flow sequences
- State diagrams
- Component interactions
- Event flow visualization

### 4. [Developer Guide](docs/developer-guide.md)
- Getting started
- Configuration guide
- API usage examples
- Implementing new attacks
- Best practices
- Troubleshooting

## Quick Start

### Enable Byzantine Module

```bash
# Build with Byzantine support
make geth

# Run with Byzantine enabled
./build/bin/geth --byzantine.enabled --byzantine.config byzantine-config.json
```

### Basic Configuration

```json
{
  "enabled": true,
  "attacks": [
    {
      "name": "network_partition",
      "type": "policy",
      "enabled": true,
      "seq_s": 100,
      "seq_e": 200,
      "round": -1,  // -1 means all rounds (wildcard)
      "max_execution_count": 0,  // 0 means unlimited
      "parameters": {
        "code": 31,  // Message code bitmap
        "fields": [
          {
            "target": "policy.direction",
            "value": 3  // 1=send, 2=receive, 3=both
          },
          {
            "target": "policy.targets",
            "value": [
              "0x2493a84a8f83cb87fdcbe0bb3b2d313f69a58d3c",
              "0x8c4a10b9108d49b9d23f764464090831d9c17764"
            ]
          }
        ]
      }
    }
  ]
}
```

### RPC API Methods

```javascript
// Get all registered Byzantine tests
const tests = await byzantine.byzantineTests();

// Stop Byzantine tests by UIDs
await byzantine.stopByzantineTests(["uid1", "uid2"]);

// Set message policy attack
await byzantine.setMessagePolicy({
  sequence: 100,
  round: 0,
  code: 1,  // 1=PrePrepare, 2=Prepare, 4=Commit, 8=RoundChange, 16=Propagation
  direction: 1,  // 1=send, 2=receive, 3=both
  targets: ["0x1234...", "0x5678..."]
});

// Send tampered message attack
await byzantine.sendTamperedMessage({
  sequence: 100,
  round: 0,
  code: 2,
  fields: {
    "msg.digest": "0xabcd..."
  },
  withValid: false
});

// Send fake message attack
await byzantine.sendFakeMessage({
  sequence: 100,
  round: 0,
  code: 4,
  fields: {
    "msg.proposal": ""
  }
});

// Other RPC methods
await byzantine.sendOmitMessage(...);
await byzantine.replayMessage(...);
await byzantine.roleSpoof(...);
await byzantine.storeMessage(...);
await byzantine.dosAttack(...);
```

## Key Features

### Attack Types

| Attack Type | Description | Key Parameters |
|------------|-------------|----------------|
| **policy** | Message policy control (drop/delay) | code, direction, targets |
| **tamper** | Modify message content | code, fields, withValid |
| **fake** | Generate invalid messages | code, fields |
| **omit** | Remove required message fields | code, cmd |
| **roleSpoof** | Impersonate consensus roles | code, fields |
| **replay** | Reuse previously stored messages | code, useOriginalView |
| **store** | Store messages for later replay | code |
| **dos** | Denial of Service attacks | code, delay, count |

### Message Codes (Bitmap)

| Code | Message Type | Description |
|------|-------------|-------------|
| 1 | PrePrepare | Pre-prepare phase message |
| 2 | Prepare | Prepare phase message |
| 4 | Commit | Commit phase message |
| 8 | RoundChange | Round change message |
| 16 | Propagation | Block propagation |
| 31 | All | All message types (1+2+4+8+16) |

### Integration Points
- **Consensus Hooks**: Non-intrusive integration with WBFT consensus engine
- **Ethereum Backend**: Integration with eth module for block propagation attacks
- **Event System**: Monitor and react to consensus events
- **Attack Registry**: Pluggable attack implementations
- **RPC API**: Runtime configuration and monitoring

### Safety Features
- Testing-only mode enforcement
- Maximum execution count limits
- Attack validation before execution
- Comprehensive logging with [BYZ] prefix
- Round wildcard support (round=-1 for all rounds)

## Architecture

```
Byzantine Module
├── service/           # Service lifecycle and orchestration
│   ├── byzantine_service.go
│   ├── consensus_hook.go
│   └── config_loader.go
├── manager/           # Attack execution management
│   └── attack_manager.go
├── attacks/           # Attack implementations
│   ├── dos_attack.go
│   ├── fake_attack.go
│   ├── message_policy_attack.go
│   ├── omit_attack.go
│   ├── replay_attack.go
│   ├── rolespoof_attack.go
│   ├── store_attack.go
│   └── tamper_attack.go
├── registry/          # Attack registration and factory
│   ├── attack_registry.go
│   ├── factory.go
│   └── attack_param_parser.go
├── api/              # RPC API interface
│   ├── api.go
│   ├── handler.go
│   └── converter.go
├── adapter/          # Integration adapters
│   ├── hook_adapter.go
│   └── event_observer.go
├── types/            # Type definitions
│   ├── attack_types.go
│   ├── attack_collection.go
│   ├── interface.go
│   └── ...
└── cmd/              # CLI commands and flags
    ├── commands.go
    ├── flags.go
    └── config.go
```

## Use Cases

1. **Consensus Testing**: Verify WBFT handles Byzantine behaviors correctly
2. **Network Partition Simulation**: Test consensus under network split scenarios
3. **Security Auditing**: Test network resilience against various attack vectors
4. **Performance Testing**: Measure impact of malicious behaviors on consensus
5. **Protocol Development**: Validate new consensus features and improvements

## Example Scenarios

### Network Partition Attack
Simulates network partition where certain validators cannot communicate:
```json
{
  "name": "network_partition",
  "type": "policy",
  "seq_s": 100,
  "seq_e": 200,
  "round": -1,
  "parameters": {
    "code": 31,
    "fields": [
      {"target": "policy.direction", "value": 3},
      {"target": "policy.targets", "value": ["0x1234...", "0x5678..."]}
    ]
  }
}
```

### Silent Validator Attack
Validator stops sending consensus messages:
```json
{
  "name": "silent_validator",
  "type": "policy",
  "seq_s": 50,
  "seq_e": 100,
  "round": 0,
  "parameters": {
    "code": 7,  // PrePrepare + Prepare + Commit
    "fields": [
      {"target": "policy.direction", "value": 1}  // Drop on send
    ]
  }
}
```

## Important Notes

**Warning**: The Byzantine module is designed for testing purposes only and should NEVER be enabled in production environments.

## Contributing

When contributing to the Byzantine module:
1. Follow the existing code structure in respective packages
2. Add attack implementations to `attacks/` directory
3. Register new attacks in the factory
4. Update type definitions in `types/`
5. Add RPC methods if needed in `api/`
6. Include comprehensive tests
7. Update documentation
8. Follow Go best practices and conventions

## License

This module is part of the WEMIX blockchain implementation and follows the same license terms.
