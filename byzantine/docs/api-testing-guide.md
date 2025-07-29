# Byzantine API Testing Guide

This guide shows how to test the updated Byzantine API that now supports config-file-like attack registration via RPC.

## New API Methods

### 1. RegisterAttacks - Batch Attack Registration

Register multiple attacks at once, similar to `loadAttacksFromConfig()`:

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc": "2.0",
    "method": "byzantine_registerAttacks",
    "params": [{
      "attacks": [
        {
          "name": "silent_attack_range",
          "type": "policy",
          "enabled": true,
          "seq_s": 100,
          "seq_e": 200,
          "round": 0,
          "parameters": {
            "code": 2,
            "fields": [
              {
                "target": "policy.direction",
                "value": 1
              }
            ],
            "targets": []
          }
        },
        {
          "name": "fake_attack_single",
          "type": "fake",
          "enabled": true,
          "seq_s": 150,
          "seq_e": 150,
          "round": 0,
          "parameters": {
            "code": 2,
            "fakeMessage": [
              {
                "target": "PrevPrePareSeal",
                "value": "nil"
              }
            ],
            "targets": []
          }
        }
      ]
    }],
    "id": 1
  }' \
  http://localhost:8545
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "results": [
      {
        "uid": "silent-1-100-200-0",
        "success": true
      },
      {
        "uid": "fake-2-150-0-0",
        "success": true
      }
    ],
    "total": 2,
    "success": 2,
    "failed": 0
  }
}
```

### 2. GetActiveAttacks - List Active Attacks

Get all currently active or pending attacks:

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc": "2.0",
    "method": "byzantine_getActiveAttacks",
    "params": [],
    "id": 1
  }' \
  http://localhost:8545
```

### 3. GetAttackStatus - Get Specific Attack Status

Get detailed status of a specific attack by UID:

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc": "2.0",
    "method": "byzantine_getAttackStatus",
    "params": ["silent-1-100-200-0"],
    "id": 1
  }' \
  http://localhost:8545
```

### 4. GetAttackMetrics - Get Attack Metrics

Get overall metrics about attacks:

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc": "2.0",
    "method": "byzantine_getAttackMetrics",
    "params": [],
    "id": 1
  }' \
  http://localhost:8545
```

## Updated Individual Attack Methods

All individual attack methods now support sequence ranges:

### Silent Message Attack with Range

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc": "2.0",
    "method": "byzantine_silentMessage",
    "params": [{
      "seq_s": 100,
      "seq_e": 200,
      "round": 0,
      "code": 1,
      "direction": 1,
      "targets": []
    }],
    "id": 1
  }' \
  http://localhost:8545
```

### Backward Compatibility

The old single sequence format still works:

```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{
    "jsonrpc": "2.0",
    "method": "byzantine_silentMessage",
    "params": [{
      "sequence": 100,
      "round": 0,
      "code": 1,
      "direction": 1,
      "targets": []
    }],
    "id": 1
  }' \
  http://localhost:8545
```

## Config File vs API Comparison

### Config File Format
```yaml
attacks:
  - name: "silent_attack"
    type: "silent"
    enabled: true
    seq_s: 100
    seq_e: 200
    round: 0
    parameters:
      code: 1
      direction: 1
```

### Equivalent API Call
```json
{
  "attacks": [{
    "name": "silent_attack",
    "type": "policy",
    "enabled": true,
    "seq_s": 100,
    "seq_e": 200,
    "round": 0,
    "parameters": {
    "code": 1,
      "fields": [
        {
          "target": "policy.direction",
          "value": 1
        }
      ],
      "targets": []
    }
  }]
}
```

## Attack Type Mappings

- `"silent"` → Silent Message Attack
- `"tamper"` → Tampered Message Attack
- `"fake"` → Fake Message Attack
- `"omit"` → Omit Message Attack
- `"roleSpoof"` → Role Spoofing Attack
- `"replay"` → Replay Attack
- `"store"` → Store Message Attack

## Notes

1. The API now accepts both `float64` and `uint64` for numeric parameters, improving compatibility with JSON-RPC.
2. Sequence ranges are supported via `seq_s` and `seq_e` parameters.
3. Single sequences can still be specified using the `sequence` parameter for backward compatibility.
4. The `enabled` field defaults to `true` if not specified.
5. Attack UIDs are automatically generated based on type, sequence range, and round.