# Omit Attack Configuration Guide

## 개요
Omit Attack은 WBFT 합의 메시지에서 특정 필드를 누락시켜 전송하는 공격입니다. 이 가이드는 omit attack 설정 방법을 설명합니다.

## 설정 파일 구조

### 기본 구조
```json
{
  "enabled": true,
  "attacks": [
    {
      "name": "attack_name",
      "type": "omitMessage",
      "sequence_start": 10,
      "sequence_end": 20,
      "round": 0,
      "code": 1,
      "enabled": true,
      "max_executions": 5,
      "parameters": {
        "code": 1,
        "cmd": 1,
        "cnt": 0,
        "targets": []
      }
    }
  ]
}
```

## 파라미터 설명

### 공통 파라미터
- `name`: 공격 식별자
- `type`: "omitMessage" (고정값)
- `sequence_start`: 공격 시작 시퀀스
- `sequence_end`: 공격 종료 시퀀스
- `round`: 라운드 번호 (0은 모든 라운드)
- `enabled`: 공격 활성화 여부
- `max_executions`: 최대 실행 횟수

### code 파라미터 (메시지 타입)
- `1`: PrePrepare
- `2`: Prepare
- `4`: Commit
- `8`: RoundChange
- `16`: RoundChange-PrePrepare
- `32`: Propagation

### cmd 파라미터 (누락 대상)
#### PrePrepare (code=1)
- `cmd=1`: 이전 블록의 Prepare Seal 누락
- `cmd=2`: 이전 블록의 Commit Seal 누락

#### Propagation (code=32)
- `cmd=1`: 현재 블록의 Prepare Seal 누락
- `cmd=2`: 현재 블록의 Commit Seal 누락

#### RoundChange-PrePrepare (code=16)
- `cmd=1`: RoundChange 메시지 누락
- `cmd=2`: Prepare 메시지 누락

### cnt 파라미터 (누락 개수)
- `0`: 모든 항목 누락
- `n`: 처음 n개 항목만 누락

### targets 파라미터
- 빈 배열 `[]`: 모든 밸리데이터에게 전송
- 주소 배열: 지정된 밸리데이터에게만 전송

## 사용 예시

### 1. 기본 설정 (omit_attack_minimal.json)
```json
{
  "name": "simple_omit_test",
  "type": "omitMessage",
  "sequence_start": 5,
  "sequence_end": 10,
  "round": 0,
  "code": 1,
  "enabled": true,
  "max_executions": 3,
  "parameters": {
    "code": 1,
    "cmd": 1,
    "cnt": 0,
    "targets": []
  }
}
```
- 시퀀스 5-10에서 PrePrepare 메시지의 모든 이전 Prepare Seal 누락

### 2. 부분 누락 설정
```json
{
  "parameters": {
    "code": 1,
    "cmd": 2,
    "cnt": 3,
    "targets": []
  }
}
```
- 이전 Commit Seal 중 처음 3개만 누락

### 3. 타겟 지정 설정
```json
{
  "parameters": {
    "code": 32,
    "cmd": 2,
    "cnt": 1,
    "targets": [
      "0x1234567890123456789012345678901234567890",
      "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
    ]
  }
}
```
- 특정 2개 밸리데이터에게만 누락된 메시지 전송

## 테스트 시나리오

### 1. omit_attack_config.json
- 다양한 omit attack 설정 예시
- 각 메시지 타입별 테스트 케이스 포함

### 2. omit_attack_test_scenarios.json
- 실제 테스트 시나리오와 예상 결과
- 각 공격의 목적과 검증 포인트 설명

### 3. omit_attack_combined.json
- 여러 공격을 조합한 복합 시나리오
- Silent attack과 함께 사용하는 예시

## 실행 방법

1. 설정 파일을 byzantine 노드의 설정 경로에 복사
2. Byzantine 노드 시작 시 설정 파일 지정:
   ```bash
   geth --byzantine.config=omit_attack_config.json
   ```

3. 로그 확인:
   ```
   [byzantine] Omitting fields from PrePrepare
   [byzantine] Omitted prev prepare seals
   [byzantine] Sending omitted message
   ```

## 주의사항

1. **합의 영향**: Seal을 모두 누락시키면 합의가 실패할 수 있음
2. **검증 실패**: 누락된 필드로 인해 메시지가 거부될 수 있음
3. **테스트 환경**: 프로덕션 환경에서는 절대 사용하지 말 것

## 디버깅

로그에서 다음 패턴을 확인:
- `[byzantine] Omitting fields from [MessageType]`
- `[byzantine] Omitted [field_type]`
- `[byzantine] Sending omitted message`

합의 실패 시 확인할 로그:
- `WBFT: invalid PRE-PREPARE block proposal`
- `insufficient seals`
- `validation failed`