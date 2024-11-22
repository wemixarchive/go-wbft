package core

import "testing"

func TestToPriority(t *testing.T) {
	// addToExtraSeal 에서 view priority 에 맞게 queue 에 쌓이는지 테스트
}

func TestProcessExtraSeal(t *testing.T) {
	// ProcessExtraSeal 동작 테스트
	// 1. latestView 보다 이전 View 가진 메세지 버림 처리 테스트
	// 2. digest 안맞으면 버림 처리 테스트
	// 3. prepare/commit 메세지가 아닌 메세지가 들어올 경우 테스트
}

func TestAddingExtraSeals(t *testing.T) {
	// 모은 extraseal이 잘 들어가는지 테스트 ( wbft 플로우 타야해서 여기서 못할 수도..?)
	// 다양한 view, round 가정해서 테스트
}
