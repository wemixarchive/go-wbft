# Byzantine Module Documentation

## Overview

The Byzantine module is a comprehensive testing framework for simulating Byzantine behaviors in the go-wemix-qbft consensus system. It provides a modular and extensible architecture for implementing various attack scenarios to test the robustness of the QBFT consensus algorithm.

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
      "name": "silent_validator",
      "type": "policy",
      "enabled": false,
      "seq_s": 92,
      "seq_e": 100,
      "round": 0,
      "parameters": {
        "code": 1,
        "fields": [
          {
            "target": "policy.direction",
            "value": 1
          }
        ],
        "targets":[]
      }
    }
  ]
}
```

### Register Attack via RPC

```javascript
// Register a new attack
byzantine.registerAttack({
  name: "silent_validator",
  type: "policy",
  enabled: false,
  seq_s: 92,
  seq_e: 100,
  round: 0,
  parameters: {
    code: 1,
    fields: [
      {
        target: "policy.direction",
        value: 1
      }
    ],
    targets:[]
  }
})
```

## Key Features

### Attack Types
- **Set Message Policy**: Control original message sending
- **Tampered Message**: Modify message content
- **Fake Message**: Generate invalid messages
- **Omit Message**: Remove required fields
- **Role Spoofing**: Impersonate other roles
- **Replay Attack**: Reuse old messages
- **Store Message**: Store messages and reuse them to perform Byzantine attacks
- **Message Flooding**: DDoS with excessive messages

### Integration Points
- **Consensus Hooks**: Non-intrusive integration with WBFT
- **Event System**: Monitor and react to consensus events
- **Storage Layer**: Persist messages and attack history
- **RPC API**: Runtime configuration and monitoring

### Safety Features
- Intended for testing only
- Configurable resource limits
- Attack validation
- Comprehensive logging

## Architecture Highlights

```
Byzantine Module
├── Service Layer (Lifecycle & Orchestration)
├── Attack Manager (Attack Execution)
├── Storage Layer (Message & History)
├── Event System (Consensus Monitoring)
├── API Layer (RPC Interface)
└── Consensus Hooks (WBFT Integration)
```

## Use Cases

1. **Consensus Testing**: Verify QBFT handles Byzantine behaviors correctly
2. **Security Auditing**: Test network resilience against attacks
3. **Performance Testing**: Measure impact of malicious behaviors
4. **Protocol Development**: Validate new consensus features

## Important Notes

⚠️ **Warning**: The Byzantine module is designed for testing purposes only and should NEVER be enabled in production environments.

## Contributing

When contributing to the Byzantine module:
1. Follow the existing code structure
2. Add comprehensive tests
3. Update documentation
4. Follow Go best practices
5. Ensure attacks are configurable and safe

## License

This module is part of go-wemix-qbft and follows the same license terms.
