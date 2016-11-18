# fabric-vote

This evolved from the [Hyperledger Starter Kit](https://hyperledger-fabric.readthedocs.io/en/latest/starter/fabric-starter-kit/#fabric-starter-kit).  

### Admin APIs
#### Create Ballot

`POST /ballot`

Creates a set of decisions.  A ballot is just a payload of decisions with some metadata around privacy.

Payload:
```
{
   "Ballot": {
      "Name": "Test Election",
      "Private": false
   },
   "Decisions": [{
      "Id": "favorite-color",
      "Name": "What is your favorite color?",
      "Options": ["Red", "Blue","Green"]
   }, {
      "Id": "favorite-beer",
      "Name": "Pick your two favorite beers",
      "Options": ["IPA", "Amber Ale","Stout","Pilsner"],
      "ResponsesRequired": 2,
      "Repeatable": true,
      "RepeatVoteDelayNS": 86400000000
   }]
}
````
##### Fields
- **Ballot.Name**: Name for ballot (not displayed today)
- **Ballot.Private**: (default=true) If true, only assigned voters will see the decisions.  If false, any voter for this account will see these decisions.  
- **Decision.Id**: Key for this decision (must be unique)
- **Decision.Name**: Displayable name for this decision
- **Decision.Options**: List of options for this decision
- **Decision.RepeatVoteDelayNS**: Wait period before a repeat-vote is allowed
- **Decision.ResponsesRequired**: Number of vote units that must be spent in a decision.

#### Get Results 

`GET /decision/{decision-id}`

Gets current results for a decision

Response:
```
{
   "Id": "favorite-color",
   "Results": {
      "ALL" : {
         "Red": 1,
         "Blue": 2
      }
   }
}
````
##### Fields
- **Id**: Unique identifier for this decision
- **Results.{key}**: Values like ALL above are a partition of voters.  By default, only ALL is a partition.  
- **Results[key] map**: The results are in the form OPTION:SCORE

### Voter APIs
#### Get Ballot for Voter

`GET /ballot/{voter-id}`

Returns the list of decisions this voter is eligible to make.  These might be across ballots within an account.  A UX can decide whether these are in different views.  (e.g., voting for favorite cake and voting for company board might not be on same page)

This calls an `init_voter` followed by a `get_ballot` chaincode transaction.

Response:
```
[{
   "Id": "favorite-color",
   "Name": "What is your favorite color?",
   "BallotId": "ba0d6eee-6f45-4a0c-b3f7-2f8659b72c2b",
   "Options": [
      "red",
      "blue",
      "green"
   ],
   "Repeatable": false,
   "RepeatVoteDelayNS": 0,
   "ResponsesRequired": 1
},
{
   "Id": "favorite-beer",
   "Name": "What is your favorite beer?",
   "BallotId": "47db9c36-af07-4383-baaf-0e143c4cb232",
   "Options": [
      "ipa",
      "amber ale",
      "pilsner",
      "stout"
   ],
   "Repeatable": false,
   "RepeatVoteDelayNS": 0,
   "ResponsesRequired": 1
}]
```
##### Fields
- **Id**: Unique identifier for this decision
- **Name**: Displayable name for this decision
- **BallotId**: Which ballot this decision was created for
- **Options**: List of selections
- **Repeatable**: Whether a user can vote more than once
- **RepeatVoteDelayNS**: Wait period before a repeat-vote is allowed
- **ResponsesRequired**: Number of vote units that must be spent in a decision.

#### Cast Vote

`POST /vote/{voter-id}`
Casts a vote for a voter

Payload:
```
[{
   "DecisionId": "favorite-color",
   "Selections": {
     "red": 1
   }
 },
 {
   "DecisionId": "favorite-beer",
   "Selections": {
     "ipa": 1
   }
}]
```
##### Fields
- **DecisionId**: Unique identifier for this decision
- **Selections**: Map of selection to number of votes to allocate (must add up to ResponsesRequired)

### Chaincode (golang):  

This contains blockchian transactions for creating decisions, voters, and casting votes. 

#### Admin Invoke Transactions
- `add_decision`: (admin) create a decision configuration 
- `add_ballot`: (admin) creates a ballot with list of decision objects, returns ballot with ID
- `add_voter`: (admin) creates a voter on blockchain and allocates votes *may not be needed*

#### Voter Invoke Transactions
- `init_voter`: (voter) lazy-creates a voter and allocates votes for all 'public ballots' in same account
- `cast_votes`: (voter) spends votes on decisions, which updates results, removes votes from voter

#### Query Transactions
- `get_results`: (admin) retrieves current results of a given decision
- `get_ballot`: (voter) retrieves ballot and vote units for the current user (using certificate attribute)

### membersrvc

For now, this is a hardcoded yaml with the initial admin.  All users are created at runtime via the node apps.  The steps follow the registration/enrollment process defined here: http://hyperledger-fabric.readthedocs.io/en/latest/protocol-spec/#421-userclient-enrollment-process

### node clients

These are basic clients that interact with a docker peer and can `Invoke` or `Query` chaincode transactions.
 
The Hyperledger rest API is available on port `7050` on the peer node.
