# Byzantine Module Diagrams

## 1. Module Architecture Overview

```mermaid
graph TB
    subgraph "Geth Node"
        Main[cmd/geth/main.go]
        Config[cmd/geth/config.go]
        
        subgraph "Byzantine Module"
            Register[byzantine/register.go]
            Service[ByzantineService]
            API[Byzantine API]
            
            subgraph "Core Components"
                AM[AttackManager]
                MS[MessageStorage]
                HS[HistoryStorage]
                EP[EventPublisher]
                CH[ConsensusHook]
            end
            
            subgraph "Attack Implementations"
                SA[Silent Attack]
                DV[Double Vote]
                TA[Tamper Attack]
                FA[Fake Attack]
                OA[Omit Attack]
                RS[Role Spoof]
                RA[Replay Attack]
                MF[Message Flooding]
            end
        end
        
        subgraph "Consensus Layer"
            Backend[WBFT Backend]
            Handler[Message Handler]
            Core[QBFT Core]
        end
    end
    
    Main --> |RegisterFlags| Register
    Main --> |RegisterCommands| Register
    Config --> |Register Service| Register
    Register --> |Create| Service
    Register --> |SetByzantineHook| Backend
    
    Service --> AM
    Service --> MS
    Service --> HS
    Service --> EP
    Service --> CH
    Service --> API
    
    AM --> SA
    AM --> DV
    AM --> TA
    AM --> FA
    AM --> OA
    AM --> RS
    AM --> RA
    AM --> MF
    
    CH --> |Hook| Backend
    CH --> |Hook| Handler
    Backend --> Core
    Handler --> Core
```

## 2. Message Flow Diagram

```mermaid
sequenceDiagram
    participant Node as Geth Node
    participant WBFT as WBFT Backend
    participant Hook as ConsensusHook
    participant AM as AttackManager
    participant Attack as Attack Instance
    participant Network as P2P Network
    
    Note over Node,Network: Outbound Message Flow
    Node->>WBFT: Broadcast(message)
    WBFT->>Hook: BeforeBroadcast(code, seq, round)
    Hook->>AM: EvaluateAndExecuteAttacks(event)
    AM->>Attack: CheckExecuteCondition(event)
    Attack-->>AM: true/false
    AM->>Attack: Execute(event)
    Attack-->>AM: AttackResult
    AM-->>Hook: AttackDecision
    Hook-->>WBFT: true (send) / false (drop)
    alt Message Allowed
        WBFT->>Network: Send Message
    else Message Blocked
        Note over WBFT: Message Dropped
    end
    
    Note over Node,Network: Inbound Message Flow
    Network->>WBFT: Receive Message
    WBFT->>Hook: BeforeProcessMessage(code, seq, round)
    Hook->>AM: EvaluateAndExecuteAttacks(event)
    AM->>Attack: CheckExecuteCondition(event)
    Attack-->>AM: true/false
    AM->>Attack: Execute(event)
    Attack-->>AM: AttackResult
    AM-->>Hook: AttackDecision
    Hook-->>WBFT: true (process) / false (drop)
    alt Message Allowed
        WBFT->>Node: Process Message
    else Message Blocked
        Note over WBFT: Message Dropped
    end
```

## 3. Attack Lifecycle State Diagram

```mermaid
stateDiagram-v2
    [*] --> Pending: RegisterAttack
    Pending --> Active: Conditions Met
    Pending --> Cancelled: CancelAttack
    
    Active --> Executing: Event Triggered
    Active --> Cancelled: CancelAttack
    
    Executing --> Executed: Execute Success
    Executing --> Failed: Execute Error
    
    Executed --> Completed: All Rounds Done
    Executed --> Active: More Rounds
    
    Failed --> [*]
    Cancelled --> [*]
    Completed --> [*]
```

## 4. Component Interaction Diagram

```mermaid
graph LR
    subgraph "External Interfaces"
        CLI[CLI Commands]
        RPC[RPC API]
        P2P[P2P Network]
    end
    
    subgraph "Byzantine Service"
        SVC[ByzantineService]
        CFG[ConfigLoader]
        
        subgraph "Management"
            AM[AttackManager]
            REG[AttackRegistry]
        end
        
        subgraph "Storage"
            MS[MessageStorage]
            HS[HistoryStorage]
        end
        
        subgraph "Integration"
            HOOK[ConsensusHook]
            ADAPT[HookAdapter]
            EVENT[EventObserver]
        end
    end
    
    subgraph "Consensus"
        WBFT[WBFT Backend]
        HANDLER[Message Handler]
    end
    
    CLI --> SVC
    RPC --> SVC
    
    SVC --> CFG
    SVC --> AM
    SVC --> MS
    SVC --> HS
    SVC --> HOOK
    
    AM --> REG
    AM --> EVENT
    
    HOOK --> ADAPT
    ADAPT --> EVENT
    
    HOOK <--> WBFT
    HOOK <--> HANDLER
    
    WBFT <--> P2P
    HANDLER <--> P2P
```

## 5. Attack Type Hierarchy

```mermaid
graph TD
    Attack[Attack Interface]
    
    Attack --> |implements| BaseAttack[BaseAttack]
    
    BaseAttack --> SA[SilentMessageAttack]
    BaseAttack --> DV[DoubleVoteAttack]
    BaseAttack --> TA[TamperAttack]
    BaseAttack --> FA[FakeAttack]
    BaseAttack --> OA[OmitAttack]
    BaseAttack --> RS[RoleSpoofAttack]
    BaseAttack --> RA[ReplayAttack]
    BaseAttack --> MF[MessageFloodingAttack]
    
    subgraph "Attack Categories"
        Safety[Safety Attacks]
        Liveness[Liveness Attacks]
        Integrity[Integrity Attacks]
        Role[Role Attacks]
        Replay[Replay Attacks]
        Network[Network Attacks]
    end
    
    DV --> Safety
    SA --> Liveness
    TA --> Integrity
    FA --> Integrity
    OA --> Integrity
    RS --> Role
    RA --> Replay
    MF --> Network
```

## 6. Event Flow Diagram

```mermaid
graph TD
    subgraph "Event Sources"
        CONS[Consensus Events]
        API[API Events]
        TIMER[Timer Events]
    end
    
    subgraph "Event System"
        PUB[EventPublisher]
        QUEUE[Event Queue]
        
        subgraph "Event Types"
            MSG_RECV[MessageReceived]
            MSG_SENT[MessageSent]
            ROUND_CHG[RoundChange]
            PROP_CREATE[ProposalCreated]
            BLK_COMMIT[BlockCommitted]
        end
    end
    
    subgraph "Event Handlers"
        AM[AttackManager]
        STORAGE[Storage Handler]
        MONITOR[Monitor Handler]
    end
    
    CONS --> PUB
    API --> PUB
    TIMER --> PUB
    
    PUB --> QUEUE
    
    QUEUE --> MSG_RECV
    QUEUE --> MSG_SENT
    QUEUE --> ROUND_CHG
    QUEUE --> PROP_CREATE
    QUEUE --> BLK_COMMIT
    
    MSG_RECV --> AM
    MSG_SENT --> AM
    ROUND_CHG --> AM
    PROP_CREATE --> AM
    BLK_COMMIT --> AM
    
    MSG_RECV --> STORAGE
    MSG_SENT --> STORAGE
    
    MSG_RECV --> MONITOR
    MSG_SENT --> MONITOR
```

## 7. API Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant RPC as RPC Server
    participant API as ByzantineAPI
    participant Service as ByzantineService
    participant AM as AttackManager
    participant Storage
    
    Client->>RPC: byzantine_registerAttack
    RPC->>API: RegisterAttack(params)
    API->>API: Validate Request
    API->>Service: RegisterAttack(config)
    Service->>Service: Generate UID
    Service->>AM: RegisterAttack(attack)
    AM->>Storage: SaveAttackConfig
    Storage-->>AM: Success
    AM-->>Service: UID
    Service-->>API: UID
    API-->>RPC: Response
    RPC-->>Client: {uid: 12345}
```

## 8. Configuration Loading Process

```mermaid
graph TD
    Start[Start]
    
    subgraph "Configuration Sources"
        CLI[CLI Flags]
        FILE[Config File]
        API[API Config]
    end
    
    subgraph "Configuration Loading"
        LOADER[ConfigLoader]
        PARSER[Config Parser]
        VALIDATOR[Config Validator]
    end
    
    subgraph "Component Configuration"
        SVC_CFG[Service Config]
        ATK_CFG[Attack Configs]
        STG_CFG[Storage Config]
        MON_CFG[Monitor Config]
    end
    
    Start --> CLI
    Start --> FILE
    
    CLI --> LOADER
    FILE --> LOADER
    API --> LOADER
    
    LOADER --> PARSER
    PARSER --> VALIDATOR
    
    VALIDATOR --> SVC_CFG
    VALIDATOR --> ATK_CFG
    VALIDATOR --> STG_CFG
    VALIDATOR --> MON_CFG
    
    SVC_CFG --> Service[ByzantineService]
    ATK_CFG --> AM[AttackManager]
    STG_CFG --> Storage[Storage]
    MON_CFG --> Monitor[Monitor]
```

## 9. Attack Decision Flow

```mermaid
flowchart TD
    Event[Consensus Event]
    
    Event --> Check{Check Active Attacks}
    
    Check -->|No Active Attacks| Allow[Allow Message]
    Check -->|Has Active Attacks| Evaluate
    
    subgraph "Evaluation Process"
        Evaluate[Evaluate Each Attack]
        Condition{Check Condition}
        Execute[Execute Attack]
        Result[Get Result]
    end
    
    Evaluate --> Condition
    Condition -->|Not Met| NextAttack{More Attacks?}
    Condition -->|Met| Execute
    Execute --> Result
    
    Result --> Decision{Should Block?}
    
    Decision -->|Yes| Block[Block Message]
    Decision -->|No| NextAttack
    
    NextAttack -->|Yes| Evaluate
    NextAttack -->|No| Allow
    
    Allow --> Continue[Continue Processing]
    Block --> Drop[Drop Message]
```
