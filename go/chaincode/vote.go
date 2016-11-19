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
	Reasons map[string]map[string]string
	Props map[string]string
}

func stringInSlice(a string, list []Option) bool {
	for _, b := range list {
		if b.Id == a {
			return true
		}
	}
	return false
}

func validate(stateDao StateDAO, vote Vote){
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
			panic("Values must add up to exactly ResponsesRequired")
		}

		for k,_ := range decision.Selections {
			if(!stringInSlice(k, d.Options)){
				panic("Invalid option: "+k)
			}
		}
	}
}

func alreadyVoted(voter Voter, decision Decision)(bool){
	return (voter.LastVoteTimestampNS > 0 && (voter.LastVoteTimestampNS > (getNow()-decision.RepeatVoteDelayNS)))
}

func addBallotDecisionsToVoter(stateDao StateDAO, ballot Ballot, voter *Voter, save bool){
	for _, decisionId := range ballot.Decisions {
		decision := stateDao.GetDecision(decisionId)
		addDecisionToVoter(voter, decision)
	}
	if(save) {
		stateDao.SaveVoter(*voter)
	}
}

func addDecisionToVoter(voter *Voter, decision Decision){
	if _, exists := voter.DecisionIdToVoteCount[decision.Id]; exists {
		//already allocated for this, skip
	}else{
		if(voter.DecisionIdToVoteCount == nil){
			voter.DecisionIdToVoteCount = make(map[string]int)
		}
		voter.DecisionIdToVoteCount[decision.Id] = decision.ResponsesRequired
	}
}

func addBallot(stateDao StateDAO, ballotDecisions BallotDecisions) (Ballot){
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

func addDecisionToBallot(stateDao StateDAO, ballotId string, decisionId string){
	ballot := stateDao.GetBallot(ballotId)
	if(ballot.Id == ""){
		ballot = Ballot{Id: ballotId, Decisions: []string{decisionId}}
		stateDao.SaveBallot(ballot)
	}
}

func log(message string){
	fmt.Printf("NETVOTE LOG: %s\n", message)
}

func castVote(stateDao StateDAO, vote Vote){
	validate(stateDao, vote)
	voter := stateDao.GetVoter(vote.VoterId)
	results_array := make([]DecisionResults, 0)

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

func addDecisionToChain(stateDao StateDAO, decision Decision) ([]byte){
	if(decision.ResponsesRequired == 0) {
		decision.ResponsesRequired = 1
	}
	results := DecisionResults { Id: decision.Id, Results: make(map[string]map[string]int)}
	stateDao.SaveDecision(decision)
	stateDao.SaveDecisionResults(results)
	return nil
}

func addDecision(stateDao StateDAO, decision Decision){
	addDecisionToChain(stateDao, decision)
	if(decision.BallotId != ""){
		addDecisionToBallot(stateDao, decision.BallotId, decision.Id)
	}
}

func addVoter(stateDao StateDAO, voter Voter){
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
		panic(err)
	}
}

func lazyInitVoter(stateDao StateDAO, voter Voter)(Voter){
	v := stateDao.GetVoter(voter.Id)
	if(v.Id != voter.Id){
		addVoter(stateDao, voter)
		v = stateDao.GetVoter(voter.Id)
	}
	return v
}

func allocateVotesToVoter(stateDao StateDAO, voter Voter)([]Decision){
	accountBallots := stateDao.GetAccountBallots()
	var result = make([]Decision, 0)
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

	stateDao := StateDAO{Stub: stub}

	if function == FUNC_ADD_DECISION {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decision Decision
			parseArg(args[0], &decision)
			addDecision(stateDao, decision)
		}
	} else if function == FUNC_ADD_BALLOT {
		if(hasRole(stub, ROLE_ADMIN)) {
			var ballotDecisions BallotDecisions
			parseArg(args[0], &ballotDecisions)
			ballot := addBallot(stateDao, ballotDecisions)
			result, err =  json.Marshal(ballot)
			if(err != nil){
				panic("error marshalling result")
			}
		}
	}else if function == FUNC_ADD_VOTER { //TODO: bulk voter adding
		if(hasRole(stub, ROLE_ADMIN)) {
			var voter Voter
			parseArg(args[0], &voter)
			addVoter(stateDao, voter)
		}
	} else if function == FUNC_INIT_VOTER {
		if(hasRole(stub, ROLE_VOTER)) {
			var voter Voter
			parseArg(args[0], &voter)
			voter = lazyInitVoter(stateDao, voter)
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

func getActiveDecisions(stateDao StateDAO, voter Voter)([]Decision){
	result := make([]Decision, 0)
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

	stateDao := StateDAO{Stub: stub}
	if function == QUERY_GET_RESULTS {
		if(hasRole(stub, ROLE_ADMIN)) {
			var decisionResults DecisionResults
			parseArg(args[0], &decisionResults)
			result, err = json.Marshal(stateDao.GetDecisionResults(decisionResults.Id))
		}
	} else if function == QUERY_GET_BALLOT {
		if(hasRole(stub, ROLE_VOTER)) {
			var voter_obj Voter
			parseArg(args[0], &voter_obj)
			voter := stateDao.GetVoter(voter_obj.Id)
			result, err = json.Marshal(getActiveDecisions(stateDao, voter))
		}
	}
	return
}

// CHAINCODE INTERFACE METHODS

func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
	function, args := stub.GetFunctionAndParameters()
	return handleInvoke(stub, function, args)
}

func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface) ([]byte, error)  {
	return nil, nil
}

func (t *VoteChaincode) Query(stub shim.ChaincodeStubInterface) ([]byte, error) {
	function, args := stub.GetFunctionAndParameters()
	return handleQuery(stub, function, args)
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}


// This class contains the accessors for getting/putting state from the blockchian
type StateDAO struct{
	Stub shim.ChaincodeStubInterface
}

//object types
const TYPE_VOTER = "VOTER"
const TYPE_DECISION = "DECISION"
const TYPE_RESULTS = "RESULTS"
const TYPE_BALLOT = "BALLOT"
const TYPE_ACCOUNT_BALLOTS = "ACCOUNT_BALLOTS"
const ATTRIBUTE_ACCOUNT_ID = "account_id"

type Option struct {
	Id string
	Name string
	Props map[string]string
}

type Decision struct {
	Id                string
	Name              string
	BallotId          string
	Options           []Option
	Props map[string]string
	ResponsesRequired int
	RepeatVoteDelayNS int64
	Repeatable        bool
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
	Id string
	Results map[string]map[string]int
}

type Voter struct {
	Id string
	Partitions []string
	DecisionIdToVoteCount map[string]int
	LastVoteTimestampNS int64
}

type AccountBallots struct{
	Id string
	PublicBallotIds map[string]bool
	PrivateBallotIds map[string]bool
}

func (t *StateDAO) getKey(objectType string, objectId string) (string){
	return t.getAccountId()+"/"+objectType+"/"+objectId
}

func (t *StateDAO) getAccountId()(string){
	//testing hack because it's tricky to mock ReadCertAttribute - hardcoded to limit risk
	if(os.Getenv("TEST_ENV") != ""){
		return "test"
	}

	account_id_bytes, err := t.Stub.ReadCertAttribute(ATTRIBUTE_ACCOUNT_ID)
	if(nil != err || string(account_id_bytes) == ""){
		panic("INVALID account ID")
	}
	return string(account_id_bytes)
}

func (t *StateDAO) getState(objectType string, id string, value interface{}){
	config, err := t.Stub.GetState(t.getKey(objectType, id))
	if(err != nil){
		panic("error getting "+objectType+" id:"+id)
	}
	json.Unmarshal(config, &value)
}

func (t *StateDAO) GetDecision(decisionId string) (Decision){
	var d Decision
	t.getState(TYPE_DECISION, decisionId, &d)
	return d
}

func (t *StateDAO) GetDecisionResults(decisionId string) (DecisionResults){
	var d DecisionResults
	t.getState(TYPE_RESULTS, decisionId, &d)
	return d
}

func (t *StateDAO) GetVoter(voterId string) (Voter) {
	var v Voter
	t.getState(TYPE_VOTER, voterId, &v)
	return v
}

func (t *StateDAO) GetBallot(ballotId string)(Ballot){
	var b Ballot
	t.getState(TYPE_BALLOT, ballotId, &b)
	return b
}

func (t *StateDAO) GetAccountBallots()(AccountBallots){
	var accountBallots AccountBallots
	t.getState(TYPE_ACCOUNT_BALLOTS, t.getAccountId(), &accountBallots)
	return accountBallots
}

func (t *StateDAO) saveState(objectType string, id string, object interface{}){
	var json_bytes, err = json.Marshal(object)
	if err != nil {
		panic("Invalid JSON while saving results")
	}
	put_err := t.Stub.PutState(t.getKey(objectType, id), json_bytes)
	if(put_err != nil){
		panic("Error while putting type:"+objectType+", id:"+id)
	}
}

func (t *StateDAO) addToAccountBallots(ballot Ballot){
	accountBallots := t.GetAccountBallots()
	account_id := t.getAccountId()
	if(accountBallots.Id != account_id){
		accountBallots = AccountBallots{Id: account_id, PrivateBallotIds: make(map[string]bool), PublicBallotIds: make(map[string]bool)}
	}
	if(ballot.Private){
		accountBallots.PrivateBallotIds[ballot.Id] = true
		delete(accountBallots.PublicBallotIds, ballot.Id)
	}else{
		accountBallots.PublicBallotIds[ballot.Id] = true
		delete(accountBallots.PrivateBallotIds, ballot.Id)
	}
	t.saveState(TYPE_ACCOUNT_BALLOTS, account_id, accountBallots)
}


func (t *StateDAO) SaveDecisionResults(decision DecisionResults){
	t.saveState(TYPE_RESULTS, decision.Id, decision)
}

func (t *StateDAO) SaveBallot(ballot Ballot){
	t.saveState(TYPE_BALLOT, ballot.Id, ballot)
	t.addToAccountBallots(ballot)
}

func (t *StateDAO) SaveVoter(v Voter){
	t.saveState(TYPE_VOTER, v.Id, v)
}

func (t *StateDAO) SaveDecision(decision Decision){
	t.saveState(TYPE_DECISION, decision.Id, decision)
}