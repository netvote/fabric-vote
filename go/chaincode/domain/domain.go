package domain

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"os"
)

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

type AccountBallots struct{
	AccountId string
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

func (t *StateDAO) getAccountBallots()(AccountBallots){
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
	accountBallots := t.getAccountBallots()
	account_id := t.getAccountId()
	if(accountBallots.AccountId != account_id){
		accountBallots = AccountBallots{AccountId: account_id, PrivateBallotIds: make(map[string]bool), PublicBallotIds: make(map[string]bool)}
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
	t.saveState(TYPE_RESULTS, decision.DecisionId, decision)
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