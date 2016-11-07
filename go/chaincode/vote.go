package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"os"
	"time"
	"strconv"
)

//TODO: if blockchains are multi-elections, will need scoping by 'election'
//TODO: add time windows for ballots/decisions? to allow valid voting periods

//object prefixes
const TYPE_VOTER = "VOTER"
const TYPE_DECISION = "DECISION"
const TYPE_RESULTS = "RESULTS"
const TYPE_BALLOT = "BALLOT"

// voter partition (defaults)
const PARTITION_ALL = "ALL"

const ATTRIBUTE_ROLE = "role"
const ATTRIBUTE_VOTER_ID = "voter_id"
const ATTRIBUTE_ACCOUNT_ID = "account_id"

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
	VoteDelayMS int64
	Repeatable bool
}

type Ballot struct{
	Id string
	Name string
	Decisions []string
	Private bool
}

type BallotDecisions struct{
	Ballot Ballot
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
	LastVoteTimestampNS int64
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

func getKey(stub shim.ChaincodeStubInterface, objectType string, objectId string) (string){
	return getAccountId(stub)+"/"+objectType+"/"+objectId
}

func getAccountId(stub shim.ChaincodeStubInterface)(string){
	//testing hack because it's tricky to mock ReadCertAttribute - hardcoded to limit risk
	if(os.Getenv("TEST_ENV") != ""){
		return "test"
	}

	account_id_bytes, err := stub.ReadCertAttribute(ATTRIBUTE_ACCOUNT_ID)
	if(nil != err || string(account_id_bytes) == ""){
		panic("INVALID account ID")
	}
	return string(account_id_bytes)
}

func validate(stub shim.ChaincodeStubInterface, vote Vote){
	var voter = getVoter(stub, vote.VoterId)
	for _, decision := range vote.Decisions {
		d := getDecision(stub, decision.DecisionId)
		if(voter.DecisionIdToVoteCount == nil) {
			panic("This voter has no votes")
		}
		if(d.ResponsesRequired != len(decision.Selections)){
			panic("All selections must be made")
		}
		if(d.Repeatable){
			if(voter.LastVoteTimestampNS > 0 && (voter.LastVoteTimestampNS > (getNow()-d.VoteDelayMS))){
				panic("Already voted this period")
			}
		}
		var total int= 0
		for _, sel := range decision.Selections{
			total += sel
		}
		if(total != voter.DecisionIdToVoteCount[decision.DecisionId]){
			panic("All votes must be cast")
		}

		for k,_ := range decision.Selections {
			if(!stringInSlice(k, d.Options)){
				panic("Invalid option: "+k)
			}
		}
	}
}

// GET

func getState(stub shim.ChaincodeStubInterface, objectType string, id string, value interface{}){
	config, err := stub.GetState(getKey(stub, objectType, id))
	if(err != nil){
		panic("error getting "+objectType+" id:"+id)
	}
	json.Unmarshal(config, &value)
}

func getDecision(stub shim.ChaincodeStubInterface, decisionId string) (Decision){
	var d Decision
	getState(stub, TYPE_DECISION, decisionId, &d)
	return d
}

func getDecisionResults(stub shim.ChaincodeStubInterface, decisionId string) (DecisionResults){
	var d DecisionResults
	getState(stub, TYPE_RESULTS, decisionId, &d)
	return d
}

func getVoter(stub shim.ChaincodeStubInterface, voterId string) (Voter) {
	var v Voter
	getState(stub, TYPE_VOTER, voterId, &v)
	return v
}

func getBallot(stub shim.ChaincodeStubInterface, ballotId string)(Ballot){
	var b Ballot
	getState(stub, TYPE_BALLOT, ballotId, &b)
	return b
}

// SAVE

func saveState(stub shim.ChaincodeStubInterface, objectType string, id string, object interface{}){
	var json_bytes, err = json.Marshal(object)
	if err != nil {
		panic("Invalid JSON while saving results")
	}
	put_err := stub.PutState(getKey(stub, objectType, id), json_bytes)
	if(put_err != nil){
		panic("Error while putting type:"+objectType+", id:"+id)
	}
}

func saveDecisionResults(stub shim.ChaincodeStubInterface, decision DecisionResults){
	saveState(stub, TYPE_RESULTS, decision.DecisionId, decision)
}

func saveBallot(stub shim.ChaincodeStubInterface, ballot Ballot){
	saveState(stub, TYPE_BALLOT, ballot.Id, ballot)
}

func saveVoter(stub shim.ChaincodeStubInterface, v Voter){
	saveState(stub, TYPE_VOTER, v.Id, v)
}

func saveDecision(stub shim.ChaincodeStubInterface, decision Decision){
	saveState(stub, TYPE_DECISION, decision.Id, decision)
}

func AllocateVotes(stub shim.ChaincodeStubInterface, voterId string, ballotId string) {

	ballot := getBallot(stub, ballotId)
	if ballot.Private {
		panic("unauthorized")
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
	saveVoter(stub, voter)

}

func AddBallot(stub shim.ChaincodeStubInterface, ballotDecisions BallotDecisions) (Ballot){
	ballot := ballotDecisions.Ballot
	ballot.Id = stub.GetTxID()
	ballot.Decisions = []string{}

	for _, decision := range ballotDecisions.Decisions {
		decision.BallotId = ballot.Id
		addDecisionToChain(stub, decision)
		ballot.Decisions = append(ballot.Decisions, decision.Id)
	}

	saveBallot(stub, ballot)
	return ballot
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

func CastVote(stub shim.ChaincodeStubInterface, vote Vote){
	validate(stub, vote)

	voter := getVoter(stub, vote.VoterId)
	results_array := make([]DecisionResults, 0)

	for _, voter_decision := range vote.Decisions {

		decisionResults := getDecisionResults(stub, voter_decision.DecisionId)
		decision := getDecision(stub,voter_decision.DecisionId)

		for selection, vote_count := range voter_decision.Selections {
			if(nil == decisionResults.Results[PARTITION_ALL]){
				decisionResults.Results[PARTITION_ALL] = map[string]int{selection: 0}
			}

			//cast vote for this decision
			decisionResults.Results[PARTITION_ALL][selection] += vote_count
			//remove votes from voter
			if(!decision.Repeatable){
				voter.DecisionIdToVoteCount[voter_decision.DecisionId] -= vote_count
			}

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
	voter.LastVoteTimestampNS = getNow()
	saveVoter(stub, voter)
}

func getNow() (int64){
	if(os.Getenv("TEST_TIME") != ""){
		i, _ := strconv.ParseInt(os.Getenv("TEST_TIME"), 10, 64)
		return i
	}
	return time.Now().UnixNano()
}

func getVoterId(stub shim.ChaincodeStubInterface) (string){
	//testing hack because it's tricky to mock ReadCertAttribute - hardcoded to limit risk
	if(os.Getenv("TEST_ENV") != ""){
		return "slanders"
	}
	voter_id_bytes, err := stub.ReadCertAttribute(ATTRIBUTE_VOTER_ID)
	if(nil != err){
		panic("invalid voter_id")
	}
	return string(voter_id_bytes)
}

func hasRole(stub shim.ChaincodeStubInterface, role string) (bool){
	if(os.Getenv("TEST_ENV") != ""){
		return true
	}
	result, _ := stub.VerifyAttribute(ATTRIBUTE_ROLE, []byte(role))
	if(!result){
		panic("unauthorized")
	}
	return result
}

func addDecisionToChain(stub shim.ChaincodeStubInterface, decision Decision) ([]byte){
	if(decision.ResponsesRequired == 0) {
		decision.ResponsesRequired = 1
	}
	results := DecisionResults { DecisionId: decision.Id, Results: make(map[string]map[string]int)}
	saveDecision(stub, decision)
	saveDecisionResults(stub, results)
	return nil
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

	stub.PutState(getKey(stub, TYPE_VOTER, voter.Id), voter_json)
	return nil, nil
}


// SimpleChaincode example simple Chaincode implementation
type VoteChaincode struct {
}

func parseArg(arg string, value interface{}){
	var arg_bytes = []byte(arg)
	if err := json.Unmarshal(arg_bytes, &value); err != nil {
		panic("error parsing arg json")
	}
}

func handleInvoke(stub shim.ChaincodeStubInterface, function string, args []string) (result []byte, err error){
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			fmt.Printf("error: %v\n",err)
		}
	}()

	if function == FUNC_ADD_DECISION {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decision Decision
			parseArg(args[0], &decision)
			result, err = AddDecision(stub, decision)
		}
	} else if function == FUNC_ADD_BALLOT {
		if(hasRole(stub, ROLE_ADMIN)) {
			var ballotDecisions BallotDecisions
			parseArg(args[0], &ballotDecisions)
			ballot := AddBallot(stub, ballotDecisions)
			result, err =  json.Marshal(ballot)
		}
	}else if function == FUNC_ADD_VOTER { //TODO: bulk voter adding
		if(hasRole(stub, ROLE_ADMIN)) {
			var voter Voter
			parseArg(args[0], &voter)
			result, err =  AddVoter(stub, voter)
		}
	} else if function == FUNC_ALLOCATE_BALLOT_VOTES {
		if(hasRole(stub, ROLE_VOTER)) {
			voter_id := getVoterId(stub)
			var ballot Ballot
			parseArg(args[0], &ballot)
			AllocateVotes(stub, voter_id, ballot.Id)
		}
	} else if function == FUNC_CAST_VOTES {
		if(hasRole(stub, ROLE_VOTER)) {
			var vote Vote
			parseArg(args[0], &vote)
			CastVote(stub, vote)
		}
	} else{
		err = errors.New("Invalid Function: "+function)
	}
	return result, err
}

func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return handleInvoke(stub, function, args)
}

// Init chain code
func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error)  {
	return nil, nil
}

func handleQuery(stub shim.ChaincodeStubInterface, function string, args []string) (result []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			fmt.Printf("error: %v\n",err)
		}
	}()
	if function == QUERY_GET_RESULTS {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decisionResults DecisionResults
			parseArg(args[0], &decisionResults)
			result, err = json.Marshal(getDecisionResults(stub, decisionResults.DecisionId))
		}
	} else if function == QUERY_GET_BALLOT {
		if(hasRole(stub, ROLE_VOTER)) {
			voter_id := getVoterId(stub)
			voter := getVoter(stub, voter_id)
			ballot := make([]Decision, 0)
			for k, _ := range voter.DecisionIdToVoteCount {
				if (voter.DecisionIdToVoteCount[k] > 0) {
					ballot = append(ballot, getDecision(stub, k))
				}
			}
			result, err = json.Marshal(ballot)
		}
	}
	return
}

// Query callback representing the query of a chaincode
func (t *VoteChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return handleQuery(stub, function, args)
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
