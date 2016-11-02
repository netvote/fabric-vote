package main

import (
	"fmt"
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"os"
)

func mockCert(){
	//makes certificate voter_id return slanders
	os.Setenv("TEST_ENV","1")
}


func checkInit(t *testing.T, stub *shim.MockStub, args []string) {
	_, err := stub.MockInit("1", "init", args)
	if err != nil {
		fmt.Println("Init failed", err)
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

func TestVoteChaincode_Invoke_AddDecision(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","Options":["a","b"]}`})

	checkState(t, stub, "DECISION_test-id", `{"Id":"test-id","Name":"What is your decision?","BallotId":"","Options":["a","b"],"ResponsesRequired":1}`)
}

func TestVoteChaincode_Invoke_AddDecisionWithBallot(t *testing.T) {
	mockCert()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-decision")

	checkInvoke(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","BallotId":"123-213412-34123-41234","Options":["a","b"]}`})

	checkState(t, stub, "DECISION_test-id", `{"Id":"test-id","Name":"What is your decision?","BallotId":"123-213412-34123-41234","Options":["a","b"],"ResponsesRequired":1}`)
	checkState(t, stub, "BALLOT_123-213412-34123-41234", `{"Id":"123-213412-34123-41234","Name":"","Decisions":["test-id"]}`)

	checkInvoke(t, stub, "allocate_ballot_votes", []string{`{"Id":"123-213412-34123-41234"}`})

	checkState(t, stub, "VOTER_slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1}}`)
}

func TestVoteChaincode_Invoke_TestMultipleAllocates(t *testing.T) {
	mockCert()
	scc := new(VoteChaincode)

	stub := shim.NewMockStub("vote", scc)
	stub.MockTransactionStart("test-invoke-add-decision")

	//setup
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"test-id","Name":"What is your decision?","BallotId":"123-213412-34123-41234","Options":["a","b"]}`})
	checkInvoke(t, stub, "allocate_ballot_votes", []string{`{"Id":"123-213412-34123-41234"}`})
	checkState(t, stub, "VOTER_slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":1}}`)

	//cast votes
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"test-id", "Selections": {"a":1}}]}`})
	checkState(t, stub, "VOTER_slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0}}`)

	//try to re-allocate votes, votes should remain at 0 for this decision
	checkInvoke(t, stub, "allocate_ballot_votes", []string{`{"Id":"123-213412-34123-41234"}`})
	checkState(t, stub, "VOTER_slanders", `{"Id":"slanders","Partitions":[],"DecisionIdToVoteCount":{"test-id":0}}`)
}

func TestVoteChaincode_Invoke_AddVoter(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-add-voter")

	checkInvoke(t, stub, "add_voter", []string{`{"Id":"voter-id","Partitions":["us","ga","123"],"DecisionIdToVoteCount":{"d1":2,"d2":1}}`})

	checkState(t, stub, "VOTER_voter-id", `{"Id":"voter-id","Partitions":["us","ga","123"],"DecisionIdToVoteCount":{"d1":2,"d2":1}}`)

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


	checkState(t, stub, "VOTER_slanders", 	`{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-ga-governor":1,"1912-us-president":1}}`)
	checkState(t, stub, "VOTER_jsmith", 	`{"Id":"jsmith","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":1,"1912-us-president":1}}`)
	checkState(t, stub, "VOTER_acooper", 	`{"Id":"acooper","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":1,"1912-us-president":1}}`)

	checkState(t, stub, "DECISION_1912-us-president", `{"Id":"1912-us-president","Name":"president","BallotId":"","Options":["Taft","Bryan"],"ResponsesRequired":1}`)
	checkState(t, stub, "DECISION_1912-ga-governor", `{"Id":"1912-ga-governor","Name":"governor","BallotId":"","Options":["Mark","Sarah"],"ResponsesRequired":1}`)

	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"slanders", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Taft":1}}, {"DecisionId":"1912-ga-governor", "Selections": {"Sarah":1}}]}`})
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"jsmith", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Bryan":1}}, {"DecisionId":"1912-ga-governor", "Selections": {"Mark":1}}]}`})
	checkInvoke(t, stub, "cast_votes", []string{`{"VoterId":"acooper", "Decisions":[{"DecisionId":"1912-us-president", "Selections": {"Taft":1}}, {"DecisionId":"1912-ga-governor", "Selections": {"Mark":1}}]}`})

	//VERIFY SIDE EFFECTS
	checkState(t, stub, "VOTER_slanders", `{"Id":"slanders","Partitions":["us","ga","district-123"],"DecisionIdToVoteCount":{"1912-ga-governor":0,"1912-us-president":0}}`)
	checkState(t, stub, "VOTER_jsmith", `{"Id":"jsmith","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":0,"1912-us-president":0}}`)
	checkState(t, stub, "VOTER_acooper", `{"Id":"acooper","Partitions":["us","ga","district-124"],"DecisionIdToVoteCount":{"1912-ga-governor":0,"1912-us-president":0}}`)

	checkState(t, stub, "RESULTS_1912-us-president", `{"DecisionId":"1912-us-president","Results":{"ALL":{"Bryan":1,"Taft":2},"district-123":{"Taft":1},"district-124":{"Bryan":1,"Taft":1},"ga":{"Bryan":1,"Taft":2},"us":{"Bryan":1,"Taft":2}}}`)
	checkState(t, stub, "RESULTS_1912-ga-governor", `{"DecisionId":"1912-ga-governor","Results":{"ALL":{"Mark":2,"Sarah":1},"district-123":{"Sarah":1},"district-124":{"Mark":2},"ga":{"Mark":2,"Sarah":1},"us":{"Mark":2,"Sarah":1}}}`)
}

func TestVoteChaincode_Query_Decision(t *testing.T) {
	scc := new(VoteChaincode)
	stub := shim.NewMockStub("vote", scc)

	stub.MockTransactionStart("test-invoke-cast-vote")
	checkInvoke(t, stub, "add_decision", []string{`{"Id":"1912-us-president","Name":"president","Options":["Taft","Bryan"]}`})

	checkQuery(t, stub, "get_results", []string{`{"DecisionId":"1912-us-president"}`}, `{"DecisionId":"1912-us-president","Results":{}}`)
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