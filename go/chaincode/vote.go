package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"os"
	"time"
)

//TODO: if blockchains are multi-elections, will need scoping by 'election'
//TODO: add time windows for ballots/decisions? to allow valid voting periods
//TODO: add repeatable votes
//TODO: add tenant ID to keys to form prefix for multi-tenancy (e.g., /{TENANT_ID}/{OBJECT_TYPE}/{ID}

//object prefixes
const VOTER_PREFIX = "VOTER_"
const DECISION_PREFIX = "DECISION_"
const RESULTS_PREFIX = "RESULTS_"
const BALLOT_PREFIX = "BALLOT_"

// voter partition (defaults)
const PARTITION_ALL = "ALL"

const ATTRIBUTE_ROLE = "role"
const ATTRIBUTE_VOTER_ID = "voter_id"

const ROLE_ADMIN = "admin"
const ROLE_VOTER = "voter"

// function names
const FUNC_ADD_DECISION = "add_decision"
const FUNC_ADD_VOTER = "add_voter"
const FUNC_ADD_BALLOT = "add_ballot"
const FUNC_CAST_VOTES = "cast_votes"
const FUNC_ALLOCATE_BALLOT_VOTES = "allocate_ballot_votes"
const QUERY_GET_RESULTS = "get_results"
const QUERY_GET_BALLOT = "get_ballot"

type Decision struct {
	Id      string
	Name    string
	BallotId string
	Options []string
	ResponsesRequired int
	VoteRate time.Duration
	Repeatable bool
}

type Ballot struct{
	Id string
	Name string
	Decisions []string
	Private bool
}

type BallotDecisions struct{
	Name string
	Decisions []Decision
}

type DecisionResults struct{
	DecisionId string
	Results map[string]map[string]int
}

type Voter struct {
	Id string
	Partitions []string
	DecisionIdToVoteCount map[string]int
	last_vote_time time.Time
}

type Vote struct {
	VoterId string
	Decisions []VoterDecision
}

type VoterDecision struct {
	DecisionId string
	Selections map[string]int
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func validate(stub shim.ChaincodeStubInterface, vote Vote) (error){
	var voter = getVoter(stub, vote.VoterId)
	for _, decision := range vote.Decisions {
		d := getDecision(stub, decision.DecisionId)
		if(voter.DecisionIdToVoteCount == nil) {
			return errors.New("This voter has no votes")
		}
		if(d.ResponsesRequired != len(decision.Selections)){
			return errors.New("All selections must be made")
		}
		var total int= 0
		for _, sel := range decision.Selections{
			total += sel
		}
		if(total != voter.DecisionIdToVoteCount[decision.DecisionId]){
			return errors.New("All votes must be cast")
		}

		for k,_ := range decision.Selections {
			if(!stringInSlice(k, d.Options)){
				return errors.New("Invalid option: "+k)
			}
		}
	}
	return nil
}

func getDecision(stub shim.ChaincodeStubInterface, decisionId string) (Decision){
	var d Decision
	var decisionConfig([]byte)
	decisionConfig, _ = stub.GetState(DECISION_PREFIX+decisionId)

	json.Unmarshal(decisionConfig, &d)
	return d
}

func getDecisionResults(stub shim.ChaincodeStubInterface, decisionId string) (DecisionResults){
	var d DecisionResults
	var config([]byte)
	config, _ = stub.GetState(RESULTS_PREFIX+decisionId)
	json.Unmarshal(config, &d)
	return d
}

func saveDecisionResults(stub shim.ChaincodeStubInterface, decision DecisionResults) (error){
	var decision_json, err = json.Marshal(decision)
	if err != nil {
		return errors.New("Invalid JSON!")
	}
	stub.PutState(RESULTS_PREFIX+decision.DecisionId, decision_json)
	return nil
}

func saveVoter(stub shim.ChaincodeStubInterface, v Voter) (error){
	var voter_json, err = json.Marshal(v)
	if err != nil {
		return errors.New("Invalid JSON!")
	}
	stub.PutState(VOTER_PREFIX+v.Id, voter_json)
	return nil
}

func getVoter(stub shim.ChaincodeStubInterface, voterId string) (Voter) {
	var v Voter
	var v_json([]byte)
	v_json, _ = stub.GetState(VOTER_PREFIX+voterId)

	json.Unmarshal(v_json, &v)
	return v
}


func saveDecision(stub shim.ChaincodeStubInterface, decision Decision) (error){
	var decision_json, err = json.Marshal(decision)
	if err != nil {
		return errors.New("Invalid JSON!")
	}
	stub.PutState(DECISION_PREFIX+decision.Id, decision_json)
	return nil
}

func allocateVotes(stub shim.ChaincodeStubInterface, voterId string, ballotId string) (error) {

	ballot := getBallot(stub, ballotId)
	if ballot.Private {
		return errors.New("Unauthorized")
	}

	voter := getVoter(stub, voterId)
	if(voter.Id == ""){
		voter.Id = voterId
		AddVoter(stub, voter)
		voter = getVoter(stub, voterId)
	}


	for _, decisionId := range ballot.Decisions {
		decision := getDecision(stub, decisionId)

		if _, exists := voter.DecisionIdToVoteCount[decisionId]; exists {
			//already allocated for this, skip
		}else{
			voter.DecisionIdToVoteCount[decisionId] = decision.ResponsesRequired
		}
	}
	return saveVoter(stub, voter)

}

func saveBallotDecisions(stub shim.ChaincodeStubInterface, ballotDecisions BallotDecisions) (Ballot){
	ballot := Ballot{Id: stub.GetTxID(), Name: ballotDecisions.Name, Decisions: []string{}}

	for _, decision := range ballotDecisions.Decisions {
		decision.BallotId = ballot.Id
		addDecisionToChain(stub, decision)
		ballot.Decisions = append(ballot.Decisions, decision.Id)
	}

	saveBallot(stub, ballot)
	return ballot
}


func saveBallot(stub shim.ChaincodeStubInterface, ballot Ballot)(error){
	var ballot_json, err = json.Marshal(ballot)
	if err != nil {
		return errors.New("Invalid JSON!")
	}
	stub.PutState(BALLOT_PREFIX+ballot.Id, ballot_json)
	return nil
}

func getBallot(stub shim.ChaincodeStubInterface, ballotId string)(Ballot){
	var b Ballot
	var config([]byte)
	config, _ = stub.GetState(BALLOT_PREFIX+ballotId)
	if(config == nil){
		return b
	}
	json.Unmarshal(config, &b)
	return b
}

func addDecisionToBallot(stub shim.ChaincodeStubInterface, ballotId string, decisionId string){
	ballot := getBallot(stub, ballotId)
	if(ballot.Id == ""){
		ballot = Ballot{Id: ballotId, Decisions: []string{decisionId}}
		saveBallot(stub, ballot)
	}
}

func log(message string){
	fmt.Printf("NETVOTE LOG: %s\n", message)
}

func CastVote(stub shim.ChaincodeStubInterface, vote Vote) ([]byte, error){
	var validation_errors = validate(stub, vote)
	if validation_errors != nil {
		return nil, validation_errors
	}

	voter := getVoter(stub, vote.VoterId)
	results_array := make([]DecisionResults, 0)

	for _, voter_decision := range vote.Decisions {

		decisionResults := getDecisionResults(stub, voter_decision.DecisionId)

		for selection, vote_count := range voter_decision.Selections {
			if(nil == decisionResults.Results[PARTITION_ALL]){
				decisionResults.Results[PARTITION_ALL] = map[string]int{selection: 0}
			}

			//cast vote for this decision
			decisionResults.Results[PARTITION_ALL][selection] += vote_count
			//remove votes from voter
			voter.DecisionIdToVoteCount[voter_decision.DecisionId] -= vote_count

			for _, partition := range voter.Partitions {
				if(nil == decisionResults.Results[partition]){
					decisionResults.Results[partition] = map[string]int{selection: 0}
				}
				decisionResults.Results[partition][selection] += vote_count
			}
		}
		results_array = append(results_array, decisionResults)

	}
	for _, d := range results_array {
		saveDecisionResults(stub, d)
	}
	voter.last_vote_time = time.Now()
	saveVoter(stub, voter)

	return nil, nil
}

func getVoterId(stub shim.ChaincodeStubInterface) (string, error){
	//testing hack because it's tricky to mock ReadCertAttribute - hardcoded to limit risk
	if(os.Getenv("TEST_ENV") != ""){
		return "slanders", nil
	}

	voter_id_bytes, err := stub.ReadCertAttribute(ATTRIBUTE_VOTER_ID)
	if(nil != err){
		return "", err
	}
	return string(voter_id_bytes), nil
}

func hasRole(stub shim.ChaincodeStubInterface, role string) (bool){
	if(os.Getenv("TEST_ENV") != ""){
		return true
	}
	result, _ := stub.VerifyAttribute(ATTRIBUTE_ROLE, []byte(role))
	return result
}

func addDecisionToChain(stub shim.ChaincodeStubInterface, decision Decision) ([]byte, error){
	if(decision.ResponsesRequired == 0) {
		decision.ResponsesRequired = 1
	}
	results := DecisionResults { DecisionId: decision.Id, Results: make(map[string]map[string]int)}
	saveDecision(stub, decision)
	saveDecisionResults(stub, results)
	return nil, nil
}

func AddDecision(stub shim.ChaincodeStubInterface, decision Decision) ([]byte, error){
	addDecisionToChain(stub, decision)
	if(decision.BallotId != ""){
		addDecisionToBallot(stub, decision.BallotId, decision.Id)
	}
	return nil, nil
}

func AddVoter(stub shim.ChaincodeStubInterface, voter Voter) ([]byte, error){
	if(voter.DecisionIdToVoteCount == nil){
		voter.DecisionIdToVoteCount = make(map[string]int)
	}
	if(voter.Partitions == nil){
		voter.Partitions = []string{}
	}
	var voter_json, err = json.Marshal(voter)
	if err != nil {
		return nil, err
	}

	stub.PutState(VOTER_PREFIX+voter.Id, voter_json)
	return nil, nil
}


// SimpleChaincode example simple Chaincode implementation
type VoteChaincode struct {
}


func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == FUNC_ADD_DECISION {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decision Decision
			var decision_bytes = []byte(args[0])
			if err := json.Unmarshal(decision_bytes, &decision); err != nil {
				return nil, err
			}
			return AddDecision(stub, decision)
		}

	} else if function == FUNC_ADD_BALLOT {
		if(hasRole(stub, ROLE_ADMIN)) {
			var ballotDecisions BallotDecisions
			var decisions_bytes = []byte(args[0])
			if err := json.Unmarshal(decisions_bytes, &ballotDecisions); err != nil {
				return nil, err
			}
			ballot := saveBallotDecisions(stub, ballotDecisions)
			return json.Marshal(ballot)
		}
	}else if function == FUNC_ADD_VOTER { //TODO: bulk voter adding
		if(hasRole(stub, ROLE_ADMIN)) {
			var voter Voter
			var voter_bytes = []byte(args[0])
			if err := json.Unmarshal(voter_bytes, &voter); err != nil {
				return nil, err
			}
			return AddVoter(stub, voter)
		}
	} else if function == FUNC_ALLOCATE_BALLOT_VOTES {
		if(hasRole(stub, ROLE_VOTER)) {
			voter_id, err := getVoterId(stub)
			if (nil != err) {
				return nil, err
			}
			var ballot Ballot
			var ballot_bytes = []byte(args[0])
			if err := json.Unmarshal(ballot_bytes, &ballot); err != nil {
				return nil, err
			}
			err = allocateVotes(stub, voter_id, ballot.Id)
			if(err != nil){
				return nil, err
			}
		}
	} else if function == FUNC_CAST_VOTES {
		if(hasRole(stub, ROLE_VOTER)) {
			var vote Vote
			var vote_bytes = []byte(args[0])
			if err := json.Unmarshal(vote_bytes, &vote); err != nil {
				return nil, err
			}
			return CastVote(stub, vote)
		}
	} else{
		return nil, errors.New("Invalid Function: "+function)
	}

	return nil, nil
}



// Init chain code
func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error)  {
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *VoteChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == QUERY_GET_RESULTS {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decision DecisionResults
			var decision_bytes = []byte(args[0])
			if err := json.Unmarshal(decision_bytes, &decision); err != nil {
				return nil, err
			}
			return json.Marshal(getDecisionResults(stub, decision.DecisionId))
		}
	} else if function == QUERY_GET_BALLOT {
		if(hasRole(stub, ROLE_VOTER)) {
			voter_id, err := getVoterId(stub)
			if (nil != err) {
				return nil, err
			}
			voter := getVoter(stub, voter_id)
			ballot := make([]Decision, 0)
			for k, _ := range voter.DecisionIdToVoteCount {
				if (voter.DecisionIdToVoteCount[k] > 0) {
					ballot = append(ballot, getDecision(stub, k))
				}
			}
			return json.Marshal(ballot)
		}
	}
	return nil, nil
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
