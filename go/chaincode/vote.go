package main

import (
	"errors"
	"fmt"
	"netvote/go/chaincode/domain"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"os"
	"time"
	"strconv"
)

//TODO: if blockchains are multi-elections, will need scoping by 'election'
//TODO: add time windows for ballots/decisions? to allow valid voting periods

// voter partition (defaults)
const PARTITION_ALL = "ALL"
const ATTRIBUTE_ROLE = "role"

const ROLE_ADMIN = "admin"
const ROLE_VOTER = "voter"

// function names
const FUNC_ADD_DECISION = "add_decision"
const FUNC_ADD_VOTER = "add_voter"
const FUNC_ADD_BALLOT = "add_ballot"
const FUNC_CAST_VOTES = "cast_votes"
const FUNC_INIT_VOTER = "init_voter"
const FUNC_ALLOCATE_BALLOT_VOTES = "allocate_ballot_votes"
const QUERY_GET_RESULTS = "get_results"
const QUERY_GET_BALLOT = "get_ballot"

type VoteChaincode struct {
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

func validate(stateDao domain.StateDAO, vote Vote){
	var voter = stateDao.GetVoter(vote.VoterId)
	for _, decision := range vote.Decisions {
		d := stateDao.GetDecision(decision.DecisionId)
		if(voter.DecisionIdToVoteCount == nil) {
			panic("This voter has no votes")
		}
		if(d.ResponsesRequired != len(decision.Selections)){
			panic("All selections must be made")
		}
		if(d.Repeatable){
			if(alreadyVoted(voter, d)){
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

func alreadyVoted(voter domain.Voter, decision domain.Decision)(bool){
	return (voter.LastVoteTimestampNS > 0 && (voter.LastVoteTimestampNS > (getNow()-decision.RepeatVoteDelayNS)))
}

func addBallotDecisionsToVoter(stateDao domain.StateDAO, ballot domain.Ballot, voter *domain.Voter, save bool){
	for _, decisionId := range ballot.Decisions {
		decision := stateDao.GetDecision(decisionId)
		addDecisionToVoter(voter, decision)
	}
	if(save) {
		stateDao.SaveVoter(*voter)
	}
}

func addDecisionToVoter(voter *domain.Voter, decision domain.Decision){
	if _, exists := voter.DecisionIdToVoteCount[decision.Id]; exists {
		//already allocated for this, skip
	}else{
		if(voter.DecisionIdToVoteCount == nil){
			voter.DecisionIdToVoteCount = make(map[string]int)
		}
		voter.DecisionIdToVoteCount[decision.Id] = decision.ResponsesRequired
	}
}

func addBallot(stateDao domain.StateDAO, ballotDecisions domain.BallotDecisions) (domain.Ballot){
	ballot := ballotDecisions.Ballot
	ballot.Id = stateDao.Stub.GetTxID()
	ballot.Decisions = []string{}

	for _, decision := range ballotDecisions.Decisions {
		decision.BallotId = ballot.Id
		addDecisionToChain(stateDao, decision)
		ballot.Decisions = append(ballot.Decisions, decision.Id)
	}

	stateDao.SaveBallot(ballot)
	return ballot
}

func addDecisionToBallot(stateDao domain.StateDAO, ballotId string, decisionId string){
	ballot := stateDao.GetBallot(ballotId)
	if(ballot.Id == ""){
		ballot = domain.Ballot{Id: ballotId, Decisions: []string{decisionId}}
		stateDao.SaveBallot(ballot)
	}
}

func log(message string){
	fmt.Printf("NETVOTE LOG: %s\n", message)
}

func castVote(stateDao domain.StateDAO, vote Vote){
	validate(stateDao, vote)
	voter := stateDao.GetVoter(vote.VoterId)
	results_array := make([]domain.DecisionResults, 0)

	for _, voter_decision := range vote.Decisions {

		decisionResults := stateDao.GetDecisionResults(voter_decision.DecisionId)
		decision := stateDao.GetDecision(voter_decision.DecisionId)

		for selection, vote_count := range voter_decision.Selections {
			if(nil == decisionResults.Results[PARTITION_ALL]){
				decisionResults.Results[PARTITION_ALL] = map[string]int{selection: 0}
			}

			//cast vote for this decision
			decisionResults.Results[PARTITION_ALL][selection] += vote_count
			//if not repeatable, remove votes from voter
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
		stateDao.SaveDecisionResults(d)
	}
	voter.LastVoteTimestampNS = getNow()
	stateDao.SaveVoter(voter)
}

func getNow() (int64){
	if(os.Getenv("TEST_TIME") != ""){
		i, _ := strconv.ParseInt(os.Getenv("TEST_TIME"), 10, 64)
		return i
	}
	return time.Now().UnixNano()
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

func addDecisionToChain(stateDao domain.StateDAO, decision domain.Decision) ([]byte){
	if(decision.ResponsesRequired == 0) {
		decision.ResponsesRequired = 1
	}
	results := domain.DecisionResults { Id: decision.Id, Results: make(map[string]map[string]int)}
	stateDao.SaveDecision(decision)
	stateDao.SaveDecisionResults(results)
	return nil
}

func addDecision(stateDao domain.StateDAO, decision domain.Decision){
	addDecisionToChain(stateDao, decision)
	if(decision.BallotId != ""){
		addDecisionToBallot(stateDao, decision.BallotId, decision.Id)
	}
}

func addVoter(stateDao domain.StateDAO, voter domain.Voter){
	if(voter.DecisionIdToVoteCount == nil){
		voter.DecisionIdToVoteCount = make(map[string]int)
	}
	if(voter.Partitions == nil){
		voter.Partitions = []string{}
	}
	stateDao.SaveVoter(voter)
}

func parseArg(arg string, value interface{}){
	var arg_bytes = []byte(arg)
	if err := json.Unmarshal(arg_bytes, &value); err != nil {
		panic("error parsing arg json")
	}
}

func lazyInitVoter(stateDao domain.StateDAO, voter domain.Voter)(domain.Voter){
	v := stateDao.GetVoter(voter.Id)
	if(v.Id != voter.Id){
		addVoter(stateDao, voter)
		v = stateDao.GetVoter(voter.Id)
	}
	return v
}

func allocateVotesToVoter(stateDao domain.StateDAO, voter domain.Voter)([]domain.Decision){
	accountBallots := stateDao.GetAccountBallots()
	var result = make([]domain.Decision, 0)
	for ballotId := range accountBallots.PublicBallotIds {
		ballot := stateDao.GetBallot(ballotId)
		log("ballot:")
		printJson(ballot)
		addBallotDecisionsToVoter(stateDao, ballot, &voter, false)
	}
	stateDao.SaveVoter(voter)
	//TODO: allocate private ballots if criteira is met
	return result
}

func printJson(value interface{}){
	result, _:=  json.Marshal(value)
	log(string(result))
}

func handleInvoke(stub shim.ChaincodeStubInterface, function string, args []string) (result []byte, err error){
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			fmt.Printf("error: %v\n",err)
		}
	}()

	stateDao := domain.StateDAO{Stub: stub}

	if function == FUNC_ADD_DECISION {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decision domain.Decision
			parseArg(args[0], &decision)
			addDecision(stateDao, decision)
		}
	} else if function == FUNC_ADD_BALLOT {
		if(hasRole(stub, ROLE_ADMIN)) {
			var ballotDecisions domain.BallotDecisions
			parseArg(args[0], &ballotDecisions)
			ballot := addBallot(stateDao, ballotDecisions)
			result, err =  json.Marshal(ballot)
			if(err != nil){
				panic("error marshalling result")
			}
		}
	}else if function == FUNC_ADD_VOTER { //TODO: bulk voter adding
		if(hasRole(stub, ROLE_ADMIN)) {
			var voter domain.Voter
			parseArg(args[0], &voter)
			addVoter(stateDao, voter)
		}
	} else if function == FUNC_INIT_VOTER {
		if(hasRole(stub, ROLE_VOTER)) {
			var voter domain.Voter
			parseArg(args[0], &voter)
			printJson(voter)
			voter = lazyInitVoter(stateDao, voter)
			printJson(voter)
			allocateVotesToVoter(stateDao, voter)
			result, err = json.Marshal(getActiveDecisions(stateDao, voter))
		}
	} else if function == FUNC_CAST_VOTES {
		if(hasRole(stub, ROLE_VOTER)) {
			var vote Vote
			parseArg(args[0], &vote)
			castVote(stateDao, vote)
		}
	} else{
		err = errors.New("Invalid Function: "+function)
	}
	return result, err
}

func getActiveDecisions(stateDao domain.StateDAO, voter domain.Voter)([]domain.Decision){
	result := make([]domain.Decision, 0)
	for k,_ := range voter.DecisionIdToVoteCount{
		if(voter.DecisionIdToVoteCount[k] > 0){
			decision := stateDao.GetDecision(k)
			if(!decision.Repeatable || !alreadyVoted(voter, decision)){
				result = append(result, decision)
			}
		}
	}
	return result
}

func handleQuery(stub shim.ChaincodeStubInterface, function string, args []string) (result []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			fmt.Printf("error: %v\n",err)
		}
	}()

	stateDao := domain.StateDAO{Stub: stub}
	if function == QUERY_GET_RESULTS {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decisionResults domain.DecisionResults
			parseArg(args[0], &decisionResults)
			result, err = json.Marshal(stateDao.GetDecisionResults(decisionResults.Id))
		}
	} else if function == QUERY_GET_BALLOT {
		if(hasRole(stub, ROLE_VOTER)) {
			var voter_obj domain.Voter
			parseArg(args[0], &voter_obj)
			voter := stateDao.GetVoter(voter_obj.Id)
			result, err = json.Marshal(getActiveDecisions(stateDao, voter))
		}
	}
	return
}

// CHAINCODE INTERFACE METHODS

func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return handleInvoke(stub, function, args)
}

func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error)  {
	return nil, nil
}

func (t *VoteChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return handleQuery(stub, function, args)
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
