package applogic

import (
	"context"
	"errors"
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
)

func TestExerciseReplayRequiresSameEconomicRequest(t *testing.T) {
	item := &models.TOptionExercise{
		Id: 9, ExerciseNo: "EX-9", AccountId: 3, PositionId: 7, ContractId: 11,
		ExerciseQty: decimal.RequireFromString("2.5"),
	}
	request := &option.ExerciseReq{
		AccountId: 3, PositionId: 7, ContractId: 11, ExerciseQty: "2.5",
		ClientExerciseId: "client-1",
	}
	if !sameExerciseRequest(item, request, decimal.RequireFromString("2.5")) {
		t.Fatal("same request was not recognized as an idempotent replay")
	}
	request.PositionId = 8
	if sameExerciseRequest(item, request, decimal.RequireFromString("2.5")) {
		t.Fatal("same idempotency key accepted a different position")
	}
}

func TestExerciseReplayResponseReturnsOriginalIdentity(t *testing.T) {
	item := &models.TOptionExercise{
		Id: 9, ExerciseNo: "EX-9", AccountId: 3, PositionId: 7, ContractId: 11,
		ExerciseQty: decimal.RequireFromString("2.5"),
	}
	request := &option.ExerciseReq{AccountId: 3, PositionId: 7, ContractId: 11}
	resp := exerciseReplayResponse(context.Background(), item, request, decimal.RequireFromString("2.5"))
	if resp.GetBase().GetCode() != 200 || resp.GetData().GetExerciseId() != 9 ||
		resp.GetData().GetExerciseNo() != "EX-9" {
		t.Fatalf("unexpected replay response: %+v", resp)
	}
}

func TestIsExerciseDuplicateKeyError(t *testing.T) {
	if !isExerciseDuplicateKeyError(&mysql.MySQLError{Number: 1062}) {
		t.Fatal("duplicate key error not recognized")
	}
	if !isExerciseDuplicateKeyError(
		errors.Join(errors.New("insert exercise"), &mysql.MySQLError{Number: 1062}),
	) {
		t.Fatal("wrapped duplicate key error not recognized")
	}
	if isExerciseDuplicateKeyError(&mysql.MySQLError{Number: 1213}) {
		t.Fatal("deadlock incorrectly recognized as duplicate key")
	}
}

func TestExerciseInstructionReplayRequiresSameRequest(t *testing.T) {
	item := &models.TOptionExerciseInstruction{
		Id: 5, AccountId: 3, PositionId: 7, ContractId: 11,
		InstructionType: int64(
			option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
		),
	}
	request := &option.SetExerciseInstructionReq{
		AccountId: 3, PositionId: 7, ContractId: 11,
		InstructionType: option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
	}
	if !sameExerciseInstruction(item, request) {
		t.Fatal("same expiry instruction was not recognized as a replay")
	}
	request.InstructionType = option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE
	if sameExerciseInstruction(item, request) {
		t.Fatal("same idempotency key accepted a different instruction")
	}
}

func TestValidExerciseInstructionType(t *testing.T) {
	for _, value := range []option.ExerciseInstructionType{
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE,
	} {
		if !validExerciseInstructionType(value) {
			t.Fatalf("valid instruction rejected: %s", value)
		}
	}
	if validExerciseInstructionType(
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_UNKNOWN,
	) {
		t.Fatal("unknown instruction accepted")
	}
}

func TestAllowsExerciseSubmissionOnlyForListedLifecycle(t *testing.T) {
	for _, status := range []option.ContractStatus{
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		option.ContractStatus_CONTRACT_STATUS_PAUSED,
	} {
		if !allowsExerciseSubmission(int64(status)) {
			t.Fatalf("listed lifecycle status rejected: %s", status)
		}
	}
	for _, status := range []option.ContractStatus{
		option.ContractStatus_CONTRACT_STATUS_UNKNOWN,
		option.ContractStatus_CONTRACT_STATUS_PENDING,
		option.ContractStatus_CONTRACT_STATUS_EXPIRED,
		option.ContractStatus_CONTRACT_STATUS_SETTLED,
		option.ContractStatus_CONTRACT_STATUS_OFFLINE,
	} {
		if allowsExerciseSubmission(int64(status)) {
			t.Fatalf("non-listed lifecycle status accepted: %s", status)
		}
	}
}
