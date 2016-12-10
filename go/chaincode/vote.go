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

// voter dimension (defaults)
const DIMENSION_ALL = "ALL"
const ATTRIBUTE_ROLE = "role"

const ROLE_ADMIN = "admin"
const ROLE_VOTER = "voter"

// function names
const QUERY_GET_ADMIN_BALLOT = "get_admin_ballot";

const FUNC_ADD_DECISION = "add_decision"
const FUNC_ADD_VOTER = "add_voter"
const FUNC_ADD_BALLOT = "add_ballot"
const FUNC_DELETE_BALLOT = "delete_ballot"
const FUNC_CAST_VOTES = "cast_votes"
const FUNC_INIT_VOTER = "init_voter"

const QUERY_GET_RESULTS = "get_results"
const QUERY_GET_BALLOT = "get_ballot"
const QUERY_GET_DECISIONS = "get_decisions"


type VoteChaincode struct {
}

type Option struct {
	Id string
	Name string
	Attributes map[string]string
}

type Decision struct {
	Id                string
	Name              string
	BallotId          string
	Options           []Option
	Attributes map[string]string
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
	Dimensions []string
	DecisionIdToVoteCount map[string]int
	LastVoteTimestampNS int64
	Attributes map[string]string
}

type AccountBallots struct{
	Id string
	PublicBallotIds map[string]bool
	PrivateBallotIds map[string]bool
}

type Vote struct {
	BallotId string
	VoterId string
	Decisions []VoterDecision
	Dimensions []string
}

type VoterDecision struct {
	DecisionId string
	Selections map[string]int
	Reasons map[string]map[string]string
	Attributes map[string]string
}

type VoteEvent struct {
	Ballot Ballot
	Voter Voter
	Vote Vote
	AccountId string
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
			if(nil == decisionResults.Results[DIMENSION_ALL]){
				decisionResults.Results[DIMENSION_ALL] = map[string]int{selection: 0}
			}

			//cast vote for this decision
			decisionResults.Results[DIMENSION_ALL][selection] += vote_count
			//if not repeatable, remove votes from voter
			if(!decision.Repeatable){
				voter.DecisionIdToVoteCount[voter_decision.DecisionId] -= vote_count
			}

			for _, dimension := range voter.Dimensions {
				if(nil == decisionResults.Results[dimension]){
					decisionResults.Results[dimension] = map[string]int{selection: 0}
				}
				decisionResults.Results[dimension][selection] += vote_count
			}
		}
		results_array = append(results_array, decisionResults)

	}
	for _, d := range results_array {
		stateDao.SaveDecisionResults(d)
	}
	voter.LastVoteTimestampNS = getNow()
	stateDao.SaveVoter(voter)

	ballot := stateDao.GetBallot(vote.BallotId)
	voteEvent := VoteEvent{Ballot: ballot, Vote: vote, Voter: voter}
	stateDao.setVoteEvent(voteEvent)
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

	result_bytes, _ := stub.ReadCertAttribute(ATTRIBUTE_ROLE)

	result, _ := stub.VerifyAttribute(ATTRIBUTE_ROLE, []byte(role))
	if(!result){
		panic("unauthorized: role="+string(result_bytes))
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
	if(voter.Dimensions == nil){
		voter.Dimensions = []string{}
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
	if function == FUNC_ADD_DECISION { //TODO: may not actually be a thing
		if(hasRole(stub, ROLE_ADMIN)) {
			var decision Decision
			parseArg(args[0], &decision)
			addDecision(stateDao, decision)
		}
	} else if function == FUNC_ADD_BALLOT {
		//ADD OR UPDATE
		if (hasRole(stub, ROLE_ADMIN)) {
			var ballotDecisions BallotDecisions
			parseArg(args[0], &ballotDecisions)
			addBallot(stateDao, ballotDecisions)
		}
	}else if function == FUNC_DELETE_BALLOT {
		if(hasRole(stub, ROLE_ADMIN)) {
			var ballot_payload Ballot
			parseArg(args[0], &ballot_payload)

			ballot := stateDao.GetBallot(ballot_payload.Id)
			for _, decisionId := range ballot.Decisions{
				stateDao.DeleteDecision(decisionId)
			}
			stateDao.DeleteBallot(ballot.Id)
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
	} else if function == QUERY_GET_DECISIONS {  //GETS ALL Decisions across all ballots
		if(hasRole(stub, ROLE_VOTER)) {
			var vote_obj Vote
			parseArg(args[0], &vote_obj)
			voter := stateDao.GetVoter(vote_obj.VoterId)
			result, err = json.Marshal(getActiveDecisions(stateDao, voter))
		}
	} else if function == QUERY_GET_BALLOT {  //GETS decisions for a specific ballot
		if(hasRole(stub, ROLE_VOTER)) {
			var vote_obj Vote
			parseArg(args[0], &vote_obj)
			if(vote_obj.BallotId == "" || vote_obj.VoterId == ""){
				panic("VoterId and BallotId are required")
			}
			voter := stateDao.GetVoter(vote_obj.VoterId)
			decisions := make([]Decision,0)
			active_decisions := getActiveDecisions(stateDao, voter)

			for _, d := range active_decisions{
				if(d.BallotId == vote_obj.BallotId) {
					decisions = append(decisions, d)
				}
			}

			result, err = json.Marshal(decisions)
		}
	} else if function == QUERY_GET_ADMIN_BALLOT { //TODO: short circuited by dyanmodb currently...perhaps not needed?
		if(hasRole(stub, ROLE_ADMIN)) {
			var ballot_obj Ballot
			parseArg(args[0], &ballot_obj)

			ballot := stateDao.GetBallot(ballot_obj.Id)

			bDecisions := make([]Decision,0)
			for _, decisionId := range ballot.Decisions{
				d := stateDao.GetDecision(decisionId)
				bDecisions = append(bDecisions, d)
			}

			bd := BallotDecisions { Ballot: ballot, Decisions: bDecisions }
			result, err = json.Marshal(bd)
		}
	}
	return
}

// CHAINCODE INTERFACE METHODS

func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//function, args := stub.GetFunctionAndParameters()
	return handleInvoke(stub, function, args)
}

func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error)  {
	return nil, nil
}

func (t *VoteChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//function, args := stub.GetFunctionAndParameters()
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

func (t *StateDAO) setVoteEvent(voteEvent VoteEvent){
	voteEvent.AccountId = t.getAccountId()
	var json_bytes, err = json.Marshal(voteEvent)
	if err != nil {
		panic("Invalid JSON while setting event")
	}
	t.Stub.SetEvent("VOTE", json_bytes)
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

func (t *StateDAO) deleteState(objectType string, id string){
	err := t.Stub.DelState(t.getKey(objectType, id))
	if(err != nil){
		panic("error deleting "+objectType+" id:"+id)
	}
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


func (t *StateDAO) DeleteDecision(decisionId string){
	t.deleteState(TYPE_RESULTS, decisionId);
	t.deleteState(TYPE_DECISION, decisionId);
}

func (t *StateDAO) DeleteBallot(ballotId string){
	t.deleteState(TYPE_BALLOT, ballotId);
	t.removeBallotFromAccountBallots(ballotId)
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

func (t *StateDAO) removeBallotFromAccountBallots(ballotId string){
	accountBallots := t.GetAccountBallots()
	delete(accountBallots.PublicBallotIds, ballotId)
	delete(accountBallots.PrivateBallotIds, ballotId)
	t.saveState(TYPE_ACCOUNT_BALLOTS, accountBallots.Id, accountBallots)
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