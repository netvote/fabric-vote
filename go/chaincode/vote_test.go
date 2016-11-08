package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"os"
	"strconv"
)

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

func checkInit(t *testing.T, stub *shim.MockStub, args []string) {
	_, err := stub.MockInit("1", "init", args)
	if err != nil {
		fmt.Println("Init failed", err)
		t.FailNow()
	}
}

func checkInvokeWithResponse(t *testing.T, stub *shim.MockStub, function string, txId string, args []string, value string) {
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
	_, err := stub.MockInvoke("1", function, args)
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

	checkInvokeError(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","Options":["a","b"]}`}, "unauthorized")
}

func TestVoteChaincode_Invoke_AddDecision(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","Options":["a","b"]}`})

	checkState(t, stub, "test/DECISION/test-id", `{"Id":"test-id","Name":"What is your decision?","BallotId":"","Options":["a","b"],"ResponsesRequired":1,"VoteDelayMS":0,"Repeatable":false}`)
}

func TestVoteChaincode_Invoke_AddBallotWithDecisions(t *testing.T){
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-ballot")

	checkInvokeWithResponse(t, stub, "add_ballot", "transaction-id",
		[]string{`{"Ballot":{"Name":"Nov 8, 2016"}, "Decisions":[{"Id":"test-id","Name":"What is your decision?","Options":["a","b"],"ResponsesRequired":1}]}`},
		`{"Id":"transaction-id","Name":"Nov 8, 2016","Decisions":["test-id"],"Private":false}`)

	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"Nov 8, 2016","Decisions":["test-id"],"Private":false}`)
	checkState(t, stub, "test/DECISION/test-id", `{"Id":"test-id","Name":"What is your decision?","BallotId":"transaction-id","Options":["a","b"],"ResponsesRequired":1,"VoteDelayMS":0,"Repeatable":false}`)

}

func TestVoteChaincode_Invoke_AddDecisionWithBallot(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","BallotId":"123-213412-34123-41234","Options":["a","b"]}`})

	checkState(t, stub, "test/DECISION/test-id", `{"Id":"test-id","Name":"What is your decision?","BallotId":"123-213412-34123-41234","Options":["a","b"],"ResponsesRequired":1,"VoteDelayMS":0,"Repeatable":false}`)
	checkState(t, stub, "test/BALLOT/123-213412-34123-41234", `{"Id":"123-213412-34123-41234","Name":"","Decisions":["test-id"],"Private":false}`)

	checkInvoke(t, stub, "init_voter", []string{`{"Id":"slanders"}`})

	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0}`)

	checkState(t, stub, "test/ACCOUNT_BALLOTS/test", `{"Id":"test","PublicBallotIds":{"123-213412-34123-41234":true},"PrivateBallotIds":{}}`)
}

func TestVoteChaincode_Invoke_TestMultipleAllocates(t *testing.T) {
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)
	stub.MockTransactionStart("test-invoke-add-decision")

	//setup
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","BallotId":"123-213412-34123-41234","Options":["a","b"]}`})
	checkInvoke(t, stub, "init_voter", []string{`{"Id":"slanders"}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1},"LastVoteTimestampNS":0}`)

	//cast votes
	mockTime(100)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0},"LastVoteTimestampNS":100}`)

	//try to re-allocate votes, votes should remain at 0 for this decision
	checkInvoke(t, stub, "init_voter", []string{`{"VoterId":"slanders"}`})
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0},"LastVoteTimestampNS":100}`)
	resetTime()
}

func TestVoteChaincode_Invoke_AddPrivateBallot(t *testing.T){
	mockEnv()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-ballot")

	checkInvokeWithResponse(t, stub, "add_ballot", "transaction-id",
		[]string{`{"Ballot":{"Name":"Nov 8, 2016", "Private":true}, "Decisions":[{"Id":"test-id","Name":"What is your decision?","Options":["a","b"],"ResponsesRequired":1}]}`},
		`{"Id":"transaction-id","Name":"Nov 8, 2016","Decisions":["test-id"],"Private":true}`)

	checkState(t, stub, "test/BALLOT/transaction-id", `{"Id":"transaction-id","Name":"Nov 8, 2016","Decisions":["test-id"],"Private":true}`)
	checkState(t, stub, "test/DECISION/test-id", `{"Id":"test-id","Name":"What is your decision?","BallotId":"transaction-id","Options":["a","b"],"ResponsesRequired":1,"VoteDelayMS":0,"Repeatable":false}`)
	checkState(t, stub, "test/ACCOUNT_BALLOTS/test", `{"Id":"test","PublicBallotIds":{},"PrivateBallotIds":{"transaction-id":true}}`)

	checkInvokeWithResponse(t, stub, "init_voter", "test", []string{`{"Id":"slanders"}`}, "[]")
}

func TestVoteChaincode_Invoke_AddVoter(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-voter")

	checkInvoke(t, stub, "add_voter", []string{`{"Id":"voter-id","Partitions":["us","ga","123"],"DecisionIdToVoteCount":{"d1":2,"d2":1}}`})

	checkState(t, stub, "test/VOTER/voter-id", `{"Id":"voter-id","Partitions":["us","ga","123"],"DecisionIdToVoteCount":{"d1":2,"d2":1},"LastVoteTimestampNS":0}`)

}

func TestVoteChaincode_Invoke_InvalidFunction(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-bad-function")

	checkInvokeError(t, stub, "not_real", []string{``}, "Invalid Function: not_real")
}

func TestVoteChaincode_Invoke_CastVote(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-us-president","Name":"president","Options":["Taft","Bryan"]}`})
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-ga-governor","Name":"governor","Options":["Mark","Sarah"]}`})
	
	checkInvoke(t, stub, "add_voter", []string{`{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-us-president":1,"1912-ga-governor":1}}`})
	checkInvoke(t, stub, "add_voter", []string{`{"Id":"jsmith","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-us-president":1,"1912-ga-governor":1}}`})
	checkInvoke(t, stub, "add_voter", []string{`{"Id":"acooper","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-us-president":1,"1912-ga-governor":1}}`})


	checkState(t, stub, "test/VOTER/slanders", 	`{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-ga-governor":1,"1912-us-president":1},"LastVoteTimestampNS":0}`)
	checkState(t, stub, "test/VOTER/jsmith", 	`{"Id":"jsmith","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":1,"1912-us-president":1},"LastVoteTimestampNS":0}`)
	checkState(t, stub, "test/VOTER/acooper", 	`{"Id":"acooper","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":1,"1912-us-president":1},"LastVoteTimestampNS":0}`)

	checkState(t, stub, "test/DECISION/1912-us-president", `{"Id":"1912-us-president","Name":"president","BallotId":"","Options":["Taft","Bryan"],"ResponsesRequired":1,"VoteDelayMS":0,"Repeatable":false}`)
	checkState(t, stub, "test/DECISION/1912-ga-governor", `{"Id":"1912-ga-governor","Name":"governor","BallotId":"","Options":["Mark","Sarah"],"ResponsesRequired":1,"VoteDelayMS":0,"Repeatable":false}`)

	mockTime(100)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Taft":1}}, {"DecisionId":"1912-ga-governor", "Selections": {"Sarah":1}}]}`})
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"jsmith", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Bryan":1}}, {"DecisionId":"1912-ga-governor", "Selections": {"Mark":1}}]}`})
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"acooper", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Taft":1}}, {"DecisionId":"1912-ga-governor", "Selections": {"Mark":1}}]}`})
	resetTime()

	//VERIFY SIDE EFFECTS
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-ga-governor":0,"1912-us-president":0},"LastVoteTimestampNS":100}`)
	checkState(t, stub, "test/VOTER/jsmith", `{"Id":"jsmith","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":0,"1912-us-president":0},"LastVoteTimestampNS":100}`)
	checkState(t, stub, "test/VOTER/acooper", `{"Id":"acooper","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":0,"1912-us-president":0},"LastVoteTimestampNS":100}`)

	checkState(t, stub, "test/RESULTS/1912-us-president", `{"Id":"1912-us-president","Results":{"ALL":{"Bryan":1,"Taft":2},"district-123":{"Taft":1},"district-124":{"Bryan":1,"Taft":1},"ga":{"Bryan":1,"Taft":2},"us":{"Bryan":1,"Taft":2}}}`)
	checkState(t, stub, "test/RESULTS/1912-ga-governor", `{"Id":"1912-ga-governor","Results":{"ALL":{"Mark":2,"Sarah":1},"district-123":{"Sarah":1},"district-124":{"Mark":2},"ga":{"Mark":2,"Sarah":1},"us":{"Mark":2,"Sarah":1}}}`)
}

func TestVoteChaincode_Invoke_CastRepeatableVote(t *testing.T){
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"2017-allstar","Name":"allstars","Options":["Freeman","Upton"],"Repeatable":true,"VoteDelayMS":1000}`})
	checkInvoke(t, stub, "add_voter", []string{`{"Id":"slanders","Partitions":["us"],"DecisionIdToVoteCount":{"2017-allstar":1}}`})

	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":["us"],"DecisionIdToVoteCount":{"2017-allstar":1},"LastVoteTimestampNS":0}`)
	checkState(t, stub, "test/DECISION/2017-allstar", `{"Id":"2017-allstar","Name":"allstars","BallotId":"","Options":["Freeman","Upton"],"ResponsesRequired":1,"VoteDelayMS":1000,"Repeatable":true}`)

	//FIRST VOTE ALLOWED
	mockTime(500)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"2017-allstar", "Selections": {"Freeman":1}}]}`})
	checkState(t, stub, "test/RESULTS/2017-allstar", `{"Id":"2017-allstar","Results":{"ALL":{"Freeman":1},"us":{"Freeman":1}}}`)
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":["us"],"DecisionIdToVoteCount":{"2017-allstar":1},"LastVoteTimestampNS":500}`)

	//SECOND VOTE ALLOWED
	mockTime(1500)
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"2017-allstar", "Selections": {"Freeman":1}}]}`})
	checkState(t, stub, "test/RESULTS/2017-allstar", `{"Id":"2017-allstar","Results":{"ALL":{"Freeman":2},"us":{"Freeman":2}}}`)
	checkState(t, stub, "test/VOTER/slanders", `{"Id":"slanders","Partitions":["us"],"DecisionIdToVoteCount":{"2017-allstar":1},"LastVoteTimestampNS":1500}`)

	//THIRD VOTE TOO SOON
	mockTime(2000)
	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"2017-allstar", "Selections": {"Freeman":1}}]}`},"Already voted this period")
	resetTime()
}

func TestVoteChaincode_Query_Decision(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-us-president","Name":"president","Options":["Taft","Bryan"]}`})

	checkQuery(t, stub, "get_results", []string{`{"Id":"1912-us-president"}`}, `{"Id":"1912-us-president","Results":{}}`)
}

func TestVoteChaincode_Invoke_ValidateCastMoreVotes(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-us-president","Name":"president","Options":["Taft","Bryan"]}`})

	checkInvoke(t, stub, "add_voter", []string{`{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-us-president":1}}`})

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Taft":2}}]}`}, "All votes must be cast")
}

func TestVoteChaincode_Invoke_ValidateInvalidOption(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-us-president","Name":"president","Options":["Taft","Bryan"]}`})

	checkInvoke(t, stub, "add_voter", []string{`{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-us-president":1}}`})

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Bush":1}}]}`}, "Invalid option: Bush")
}

func TestVoteChaincode_Invoke_ValidateInvalidSelections(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-us-president","Name":"president","Options":["Taft","Bryan"],"ResponsesRequired":2}`})

	checkInvoke(t, stub, "add_voter", []string{`{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-us-president":2}}`})

	checkInvokeError(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Taft":1}}]}`}, "All selections must be made")
}