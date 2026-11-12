# Croissant Hard Fork

Croissant brings the EVM-level part of Ethereum's Shanghai upgrade to WEMIX.
go-wemix runs it before the chain switches to WBFT, and go-wbft must produce the
same results when it imports those blocks. This document describes what
changes at the fork block, how the fork is ordered with Dalgona, and where
go-wbft differs from go-ethereum.

Fork names are desserts in alphabetical order of activation: Applepie,
Brioche, Croissant, Dalgona. The WBFT consensus switch was planned as
Croissant. It now comes after this fork and is called Dalgona.

## Scope

Ethereum's Shanghai bundles five EIPs. Croissant adopts the three that change
EVM execution and leaves out the two that do not apply to WEMIX.

| EIP | Title | Status |
|-----|-------|--------|
| [EIP-3855](https://eips.ethereum.org/EIPS/eip-3855) | PUSH0 instruction | Adopted |
| [EIP-3860](https://eips.ethereum.org/EIPS/eip-3860) | Limit and meter initcode | Adopted |
| [EIP-3651](https://eips.ethereum.org/EIPS/eip-3651) | Warm COINBASE | Adopted |
| [EIP-4895](https://eips.ethereum.org/EIPS/eip-4895) | Beacon chain push withdrawals | Not adopted |
| [EIP-6049](https://eips.ethereum.org/EIPS/eip-6049) | Deprecate SELFDESTRUCT | No change needed |

Why only part of Shanghai:

- **EIP-4895** pushes withdrawals from the beacon chain into the execution
  layer. WEMIX has no beacon chain, so there is nothing to withdraw.
- **EIP-6049** only announces that SELFDESTRUCT will change later. It does not
  change any behaviour.
- The three adopted EIPs are what Solidity 0.8.20 and later target by default
  (`evmVersion: shanghai`).

## Activation

Croissant is gated by block number, like the other WEMIX forks. Ethereum
activates Shanghai by timestamp instead.

The fork block is `ChainConfig.CroissantBlock`, `croissantBlock` in a genesis
`config`:

- `nil` (absent): the fork never activates.
- `0`: the chain starts on Croissant.
- `N`: blocks `N` and later run the Croissant rules.

Mainnet and testnet carry placeholder blocks that keep both forks inactive:
`999,999,999` for Croissant and `1,000,000,000` for Dalgona, the same values as
in go-wemix. They must be replaced with the scheduled blocks before a release
that activates them. The two repositories must always carry the same blocks.

### Order with Dalgona

Dalgona keeps every Croissant rule. `CheckConfigForkOrder` requires
`croissantBlock` to be set whenever `dalgonaBlock` is set, at or before it.
`gwemix init` rejects a genesis that breaks this, and so does a node started
with such a genesis or on mainnet or testnet. A chain that runs WBFT from
genesis sets both to `0`:

```json
"croissantBlock": 0,
"dalgonaBlock": 0,
```

Before this fork existed, Dalgona alone enabled `PUSH0` and the `CREATE` and
`CREATE2` part of EIP-3860. The transaction-level part of EIP-3860 and EIP-3651
were gated by Shanghai, which WEMIX never schedules, so they were off. A new
chain that sets both forks gets all three EIPs in full from Dalgona on.

A private network that already runs Dalgona without `croissantBlock` keeps
working unchanged. On restart without a genesis file, a node loads the chain
config stored in its database, and that path does not run the fork order
check. `IsCroissant` stays false, so such a network keeps the old Dalgona
rules. It cannot add Croissant later: Croissant would have to be at or before
Dalgona, and moving a fork the chain has already passed is rejected as an
incompatible change. Running `gwemix init` again with its old genesis file
fails the order check.

### Renamed genesis keys

The WBFT switch used the keys `croissantBlock` and `croissant`. They are now
`dalgonaBlock` and `dalgona`, and `croissantBlock` schedules this fork. Unknown
keys in a genesis `config` are ignored, so rename them in genesis files written
for the old meaning.

The governance contract `GovImp.sol` hardcodes the WBFT switch block for member
removal. Its constant is now `DALGONA_BLOCK` and holds the Dalgona placeholder.
It must be changed with `dalgonaBlock` when the fork is scheduled.

## Changes at the fork block

### EIP-3855: PUSH0

The opcode `0x5f` pushes the constant zero onto the stack for 2 gas. Before the
fork it is an invalid opcode.

### EIP-3860: Limit and meter initcode

The initcode limit is twice the maximum contract size, as in the EIP. WEMIX
raises `MaxCodeSize` to 253,952 bytes, so the limit is larger than on Ethereum:

| | Ethereum | WEMIX |
|-|----------|-------|
| `MaxCodeSize` | 24,576 | 253,952 |
| `MaxInitCodeSize` (`2 * MaxCodeSize`) | 49,152 | 507,904 |

go-wemix uses the same value. A different limit would split the chain on a
`CREATE` whose initcode falls between the two values.

From the fork block on:

- A creation transaction pays 2 gas per 32-byte word of initcode on top of its
  intrinsic gas.
- A creation transaction whose initcode exceeds `MaxInitCodeSize` is invalid
  and is rejected with `max initcode size exceeded`.
- `CREATE` and `CREATE2` pay 2 gas per initcode word and fail when the initcode
  exceeds `MaxInitCodeSize`.

A creation transaction is still bounded by the pool's transaction size limit
(256 KiB). With the 2x limit it can carry a contract of the maximum size
together with its constructor.

### EIP-3651: Warm COINBASE

The block's coinbase is added to the access list at the start of every
transaction. The first access to the coinbase then costs the warm price
(100 gas) instead of the cold price (2,600 gas).

## Code paths

Every change is selected by `Rules.IsCroissant`, or by `Rules.IsShanghai` as in
go-ethereum.

| EIP | Where | From the fork block on |
|-----|-------|------------------------|
| EIP-3855, EIP-3860 | `core/vm/interpreter.go` | `croissantInstructionSet`: london plus `PUSH0` and the EIP-3860 `CREATE` gas functions |
| EIP-3860 | `StateTransition.TransitionDb` | initcode word gas and `ErrMaxInitCodeSizeExceeded` |
| EIP-3860 | `txpool.ValidateTransaction` | initcode size check and initcode word gas |
| EIP-3651 | `StateDB.Prepare` | coinbase added to the access list |

Dalgona keeps its own instruction set, which already contains `PUSH0` and the
EIP-3860 gas functions.

## Differences from go-ethereum

**Instruction set base.** go-ethereum builds the Shanghai set on the merge set,
where `0x44` is `PREVRANDAO`. go-wemix never went through the merge and its
PoA headers carry difficulty 1. The Croissant set is built on the london set so
that `0x44` stays `DIFFICULTY` and both clients return the same value. From
Dalgona on, the Dalgona set provides `PREVRANDAO`.

**Block-number gating.** Shanghai is gated by timestamp and only after the
merge. Croissant is gated by block number and does not depend on the merge.

## Testing

| Package | What is checked |
|---------|-----------------|
| `params` | block-number gating, `Rules` propagation, compatibility check, order with Dalgona, mainnet and testnet order, `MaxInitCodeSize` derivation |
| `core/vm` | PUSH0 and the initcode gas functions in the Croissant set, `0x44` staying `DIFFICULTY` |
| `core/vm/runtime` | PUSH0 execution, `CREATE`/`CREATE2` size limit, `DIFFICULTY` execution |
| `core/txpool` | oversized creation transaction rejected from Croissant on |
| `core/state` | coinbase warmed under Croissant and Shanghai |
| `core` | initcode word gas, rejection of oversized creation transactions, 2,500 gas saved on a coinbase `BALANCE` |

Run them with:

```sh
go test -run 'Croissant|EIP3855|EIP3860|EIP3651|MaxInitCode|InitCodeLimit|WarmCoinbase' \
  ./params/ ./core/ ./core/state/ ./core/txpool/ ./core/vm/...
```
