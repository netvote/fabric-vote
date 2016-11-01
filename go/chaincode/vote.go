package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)

const VOTER_PREFIX = "VOTER_"
const DECISION_PREFIX = "DECISION_"
const PARTITION_ALL = "ALL"

const FUNC_ADD_DECISION = "add_decision"
const FUNC_ADD_VOTER = "add_voter"
const FUNC_CAST_VOTES = "cast_votes"

const QUERY_GET_DECISION = "get_decision"

type Decision struct {
	Id      string
	Name    string
	Options []string
	Results map[string]map[string]int
	ResponsesRequired int
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


// SimpleChaincode example simple Chaincode implementation
type VoteChaincode struct {
}

func getVoter(stub shim.ChaincodeStubInterface, voterId string) (Voter) {
	var v Voter
	var v_json([]byte)
	v_json, _ = stub.GetState(VOTER_PREFIX+voterId)

	json.Unmarshal(v_json, &v)
	return v
}

func clearVoter(stub shim.ChaincodeStubInterface, voter Voter) (error){
	voter.DecisionIdToVoteCount = nil
	var voter_json, err = json.Marshal(voter)
	if err != nil {
		return errors.New("Invalid JSON!")
	}
	stub.PutState(VOTER_PREFIX+voter.Id, voter_json)
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
	fmt.Printf("NETVOTE LOG: %s", message)
}

func CastVote(stub shim.ChaincodeStubInterface, vote Vote) ([]byte, error){
	var validation_errors = validate(stub, vote)
	log("validated!")
	if validation_errors != nil {
		log("found errors!")
		return nil, validation_errors
	}

	log("getting voter!")
	voter := getVoter(stub, vote.VoterId)
	decisions := make([]Decision, 0)

	log("any decisions?")
	for _, voter_decision := range vote.Decisions {
		log("Yep!")
		var decision Decision = getDecision(stub, voter_decision.DecisionId)

		if(nil == decision.Results){
			decision.Results = make(map[string]map[string]int)
		}

		for selection, vote_count := range voter_decision.Selections {
			log("applying selections")
			if(nil == decision.Results[PARTITION_ALL]){
				decision.Results[PARTITION_ALL] = map[string]int{selection: 0}
			}
			decision.Results[PARTITION_ALL][selection] += vote_count

			for _, partition := range voter.Partitions {
				if(nil == decision.Results[partition]){
					decision.Results[partition] = map[string]int{selection: 0}
				}
				decision.Results[partition][selection] += vote_count
			}

		}
		decisions = append(decisions, decision)
	}
	for _, d := range decisions {
		saveDecision(stub, d)
	}
	clearVoter(stub, voter)

	return nil, nil
}

func AddDecision(stub shim.ChaincodeStubInterface, decision Decision) ([]byte, error){
	if(decision.ResponsesRequired == 0) {
		decision.ResponsesRequired = 1
	}

	var decision_json, err = json.Marshal(decision)
	if err != nil {
		return nil, err
	}

	stub.PutState(DECISION_PREFIX+decision.Id, decision_json)
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


func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	group, err := stub.ReadCertAttribute("group")
	fmt.Printf("Group => %v error %v \n", string(group), err)
	if function == FUNC_ADD_DECISION {
		var decision Decision
		var decision_bytes = []byte(args[0])
		if err := json.Unmarshal(decision_bytes, &decision); err != nil {
			return nil, err
		}
		return AddDecision(stub, decision)

	} else if function == FUNC_ADD_VOTER {
		var voter Voter
		var voter_bytes = []byte(args[0])
		if err := json.Unmarshal(voter_bytes, &voter); err != nil {
			return nil, err
		}
		return AddVoter(stub, voter)

	} else if function == FUNC_CAST_VOTES {
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
	if function == QUERY_GET_DECISION {
		var decision Decision
		var decision_bytes = []byte(args[0])
		if err := json.Unmarshal(decision_bytes, &decision); err != nil {
			return nil, err
		}
		return json.Marshal(getDecision(stub, decision.Id))
	}
	return nil, nil
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
