package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)

//TODO: if blockchains are multi-elections, will need scoping by 'election'

//object prefixes
const VOTER_PREFIX = "VOTER_"
const DECISION_PREFIX = "DECISION_"
const RESULTS_PREFIX = "RESULTS_"

// voter partition (defaults)
const PARTITION_ALL = "ALL"

// function names
const FUNC_ADD_DECISION = "add_decision"
const FUNC_ADD_VOTER = "add_voter"
const FUNC_CAST_VOTES = "cast_votes"
const QUERY_GET_RESULTS = "get_results"
const QUERY_GET_BALLOT = "get_ballot"

type Decision struct {
	Id      string
	Name    string
	Options []string
	ResponsesRequired int
}

type DecisionResults struct{
	DecisionId string
	Results map[string]map[string]int
}

type Voter struct {
	Id string
	Partitions []string
	DecisionIdToVoteCount map[string]int
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
	log("saving results..."+decision.DecisionId)
	stub.PutState(RESULTS_PREFIX+decision.DecisionId, decision_json)
	return nil
}

func getVoter(stub shim.ChaincodeStubInterface, voterId string) (Voter) {
	var v Voter
	var v_json([]byte)
	v_json, _ = stub.GetState(VOTER_PREFIX+voterId)

	json.Unmarshal(v_json, &v)
	return v
}

func clearVoter(stub shim.ChaincodeStubInterface, voter Voter) (error){
	stub.DelState(VOTER_PREFIX+voter.Id)
	return nil
}

func saveDecision(stub shim.ChaincodeStubInterface, decision Decision) (error){
	var decision_json, err = json.Marshal(decision)
	if err != nil {
		return errors.New("Invalid JSON!")
	}
	stub.PutState(DECISION_PREFIX+decision.Id, decision_json)
	return nil
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
			decisionResults.Results[PARTITION_ALL][selection] += vote_count

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
	clearVoter(stub, voter)

	return nil, nil
}

func AddDecision(stub shim.ChaincodeStubInterface, decision Decision) ([]byte, error){
	if(decision.ResponsesRequired == 0) {
		decision.ResponsesRequired = 1
	}
	results := DecisionResults { DecisionId: decision.Id, Results: make(map[string]map[string]int)}
	saveDecision(stub, decision)
	saveDecisionResults(stub, results)
	return nil, nil
}

func AddVoter(stub shim.ChaincodeStubInterface, voter Voter) ([]byte, error){
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
		//TODO: authorize
		var decision Decision
		var decision_bytes = []byte(args[0])
		if err := json.Unmarshal(decision_bytes, &decision); err != nil {
			return nil, err
		}
		return AddDecision(stub, decision)

	} else if function == FUNC_ADD_VOTER {
		//TODO: authorize
		var voter Voter
		var voter_bytes = []byte(args[0])
		if err := json.Unmarshal(voter_bytes, &voter); err != nil {
			return nil, err
		}
		return AddVoter(stub, voter)

	} else if function == FUNC_CAST_VOTES {
		//TODO: authorize voter
		var vote Vote
		var vote_bytes = []byte(args[0])
		if err := json.Unmarshal(vote_bytes, &vote); err != nil {
			return nil, err
		}
		return CastVote(stub, vote)
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
		//TODO: authorize
		//TODO: validate valid decision Id
		var decision DecisionResults
		var decision_bytes = []byte(args[0])
		if err := json.Unmarshal(decision_bytes, &decision); err != nil {
			return nil, err
		}
		return json.Marshal(getDecisionResults(stub, decision.DecisionId))
	} else if function == QUERY_GET_BALLOT {
		//TODO: validate valid voter_id
		//TODO: only allow for non-voted entries
		//TODO: also return number of vote units for this voter (in map)
		voter_id_bytes, err := stub.ReadCertAttribute("voter_id")
		if(nil != err){
			return nil, err
		}
		voter := getVoter(stub, string(voter_id_bytes))
		ballot := make([]Decision, 0)
		for k, _ := range voter.DecisionIdToVoteCount {
			ballot = append(ballot, getDecision(stub, k))
		}
		return json.Marshal(ballot)
	}
	return nil, nil
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
