package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"os"
	"strconv"
)

const CREATE_DECISION_JSON = `{"Id":"test-id","Name":"What is your decision?","BallotId":"transaction-id","Options":[{"Id":"a","Name":"A","Props":{"image":"/url"}}],"Props":{"Key":"Value"}}`
const TEST_DECISION_JSON = `{"Id":"test-id","Name":"What is your decision?","BallotId":"transaction-id","Options":[{"Id":"a","Name":"A","Props":{"image":"/url"}}],"Props":{"Key":"Value"},"ResponsesRequired":1,"RepeatVoteDelayNS":0,"Repeatable":false}`


const CREATE_DECISION_JSON_REQUIRED_2 = `{"Id":"test-id","Name":"What is your decision?","BallotId":"transaction-id","Options":[{"Id":"a","Name":"A","Props":{"image":"/url"}}],"Props":{"Key":"Value"},"ResponsesRequired":2}`

const CREATE_REPEATABLE_DECISION_JSON = `{"Id":"test-id","Name":"What is your decision?","BallotId":"transaction-id","Options":[{"Id":"a","Name":"A","Props":{"image":"/url"}}],"Props":{"Key":"Value"},"ResponsesRequired":1,"RepeatVoteDelayNS":100,"Repeatable":true}`


func mockEnv(){
	//makes certificate test/VOTER/id return slanders
	os.Setenv("TEST_ENV","1")
}

func unmockEnv(){
	os.Unsetenv("TEST_ENV")
}

func mockTime(timeNS int64){
	os.Setenv("TEST_TIME", strconv.FormatInt(timeNS, 10))
}

func resetTime(){
	os.Unsetenv("TEST_TIME")
}

func to_byte_array(function string, args []string)([][]byte){
	result := [][]byte{}

	result = append(result, []byte(function))

	for _,it := range args {
		result = append(result, []byte(it))
	}
	return result
}

func checkInvokeWithResponse(t *testing.T, stub *shim.MockStub, function string, txId string, args []string, value string) {
	//b_args := to_byte_array(function, args)
	bytes, err := stub.MockInvoke(txId, function, args)
	if err != nil {
		fmt.Println("Invoke", args, "failed", err)
		t.FailNow()
	}
	if bytes == nil {
		fmt.Println("Response is nil")
		t.FailNow()
	}
	if string(bytes) != value {
		fmt.Println(string(bytes))
		fmt.Println("State value", string(bytes), "was not", value, "as expected")
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, function string, args []string) {
	checkInvokeTX(t, stub, "1", function, args)
}

func checkInvokeTX(t *testing.T, stub *shim.MockStub, transactionId string, function string, args []string) {
	fmt.Println(args)
	_, err := stub.MockInvoke(transactionId, function, args)
	if err != nil {
		fmt.Println("Invoke", args, "failed", err)
		t.FailNow()
	}
}

func checkInvokeError(t *testing.T, stub *shim.MockStub, function string, args []string, error string) {
	_, err := stub.MockInvoke("1", function, args)
	if err == nil {
		fmt.Println("No error was found, but error was expected: "+error)
		t.FailNow()
	}
	if err.Error() != error {
		fmt.Println("Expected: "+error+", Found: "+err.Error())
		t.FailNow()
	}
}

func checkGone(t *testing.T, stub *shim.MockStub, name string){
	bytes := stub.State[name]
	if bytes != nil {
		fmt.Println("Expected state", name, "to be gone")
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, name string, value string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != value {
		fmt.Println(string(bytes))
		fmt.Println("State value", name, "was not", value, "as expected")
		t.FailNow()
	}
}

func checkQuery(t *testing.T, stub *shim.MockStub, function string, args []string, value string) {
	bytes, err := stub.MockQuery(function, args)
	if err != nil {
		fmt.Println("Query", args, "failed", err)
		t.FailNow()
	}
	if bytes == nil {
		fmt.Println("Query", args, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != value {
		fmt.Println(string(bytes))
		fmt.Println("Query value", args, "was not", value, "as expected")
		t.FailNow()
	}
}

func TestVoteChaincode_Invoke_AddDecision_Error(t *testing.T) {
	unmockEnv()
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvokeError(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","Options":["a","b"]}`}, "unauthorized: role=")
}

func TestVoteChaincode_Invoke_AddDecision(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{CREATE_DECISION_JSON})

	checkState(t, stub, "test/DECISION/test-id", TEST_DECISION_JSON)
}

func TestVoteChaincode_Invoke_AddBallotWithDecisions(t *testing.T){
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-ballot")

	checkInvokeTX(t, stub,  "transaction-id", "add_ballot",
		[]string{`{"Ballot":{"Name":"Nov 8, 2016"}, "Decisions":[`+CREATE_DECISION_JSON+`]}`})

	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"Nov 8, 2016","Decisions":["test-id"],"Private":false}`)
	checkState(t, stub, "test/DECISION/test-id", TEST_DECISION_JSON)

}

func TestVoteChaincode_Invoke_AddDecisionWithBallot(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{CREATE_DECISION_JSON})

	checkState(t, stub, "test/DECISION/test-id", TEST_DECISION_JSON)
	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"","Decisions":["test-id"],"Private":false}`)

	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})

	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0,"Props":null}`)

	checkState(t, stub, "test/ACCOUNT_BALLOTS/test", `{"Id":"test","PublicBallotIds":{"transaction-id":true},"PrivateBallotIds":{}}`)
}

func TestVoteChaincode_Invoke_TestMultipleAllocates(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)
	stub.MockTransactionStart("test-invoke-add-decision")

	//setup
	checkInvoke(t, stub, "add_decision", []string{CREATE_DECISION_JSON})

	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})

	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0,"Props":null}`)

	//cast votes
	mockTime(100)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0},"LastVoteTimestampNS":100,"Props":null}`)

	//try to re-allocate votes, votes should remain at 0 for this decision
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0},"LastVoteTimestampNS":100,"Props":null}`)
	resetTime()
}

func TestVoteChaincode_Invoke_AddPrivateBallot(t *testing.T){
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-ballot")

	checkInvokeTX(t, stub,  "transaction-id", "add_ballot",
		[]string{`{"Ballot":{"Name":"Nov 8, 2016","Private":true}, "Decisions":[`+CREATE_DECISION_JSON+`]}`})

	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"Nov 8, 2016","Decisions":["test-id"],"Private":true}`)
	checkState(t, stub, "test/DECISION/test-id", TEST_DECISION_JSON)
	checkState(t, stub, "test/ACCOUNT_BALLOTS/test", `{"Id":"test","PublicBallotIds":{},"PrivateBallotIds":{"transaction-id":true}}`)

	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{},"LastVoteTimestampNS":0,"Props":null}`)

}

func TestVoteChaincode_Invoke_AddVoter(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-voter")

	checkInvoke(t, stub, "add_voter", []string{`{"Id":"voter-id","Partitions":["us","ga","123"],"DecisionIdToVoteCount":{"d1":2,"d2":1}}`})

	checkState(t, stub, "test/VOTER/voter-id", `{"Id":"voter-id","Partitions":["us","ga","123"],"DecisionIdToVoteCount":{"d1":2,"d2":1},"LastVoteTimestampNS":0,"Props":null}`)

}

func TestVoteChaincode_Invoke_InvalidFunction(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-bad-function")

	checkInvokeError(t, stub, "not_real", []string{``}, "Invalid Function: not_real")
}

func TestVoteChaincode_Invoke_CastVote(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{CREATE_DECISION_JSON})

	checkState(t, stub, "test/DECISION/test-id", TEST_DECISION_JSON)
	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"","Decisions":["test-id"],"Private":false}`)
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0,"Props":null}`)
	checkState(t, stub, "test/ACCOUNT_BALLOTS/test", `{"Id":"test","PublicBallotIds":{"transaction-id":true},"PrivateBallotIds":{}}`)
	mockTime(500)

	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`})

	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0},"LastVoteTimestampNS":500,"Props":null}`)
	checkState(t, stub, "test/RESULTS/test-id", `{"Id":"test-id","Results":{"ALL":{"a":1}}}`)
}

func TestVoteChaincode_Invoke_CastRepeatableVote(t *testing.T){
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{CREATE_REPEATABLE_DECISION_JSON})

	checkState(t, stub, "test/DECISION/test-id", CREATE_REPEATABLE_DECISION_JSON)
	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"","Decisions":["test-id"],"Private":false}`)
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0,"Props":null}`)
	checkState(t, stub, "test/ACCOUNT_BALLOTS/test", `{"Id":"test","PublicBallotIds":{"transaction-id":true},"PrivateBallotIds":{}}`)
	mockTime(500)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`})

	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":500,"Props":null}`)
	checkState(t, stub, "test/RESULTS/test-id", `{"Id":"test-id","Results":{"ALL":{"a":1}}}`)

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`}, "Already voted this period")
	mockTime(1500)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`})

}

func TestVoteChaincode_Query_Decision(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{TEST_DECISION_JSON})

	checkQuery(t, stub, "get_results", []string{`{"Id":"test-id"}`}, `{"Id":"test-id","Results":{}}`)
}

func TestVoteChaincode_Invoke_ValidateCastTooMany(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{TEST_DECISION_JSON})
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":2}}]}`}, "Values must add up to exactly ResponsesRequired")
}

func TestVoteChaincode_Invoke_ValidateInvalidOption(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{TEST_DECISION_JSON})
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"c":1}}]}`}, "Invalid option: c")
}

func TestVoteChaincode_Invoke_ValidateCastTooFew(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{CREATE_DECISION_JSON_REQUIRED_2})
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders"}`})

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`}, "All selections must be made")
}

func TestVoteChaincode_Invoke_InitVoterWithProps(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{CREATE_DECISION_JSON})
	checkInvokeTX(t, stub, "transaction-id", "init_voter", []string{`{"Id":"slanders","Props":{"key":"val"}}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0,"Props":{"key":"val"}}`)
}