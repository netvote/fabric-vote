
if(process.argv.length < 4){
    console.error("Parameters required: username, secret");
    process.exit(1);
}

userName = process.argv[2];
secret = process.argv[3];


// TEST DATA
var DECISION = {
    Id: "us-president-2016",
    Name: "President of the United States",
    Options: ["Taft", "Bryan"]
};

var VOTER = {
    Id: "slanders",
    Name: "Steven Landers",
    Partitions: ["us", "ga", "district-123"],
    DecisionIdToVoteCount: {
        "us-president-2016": 1
    }
};

var VOTE = {
    VoterId: "slanders",
    Decisions: [
        {
            DecisionId: "us-president-2016",
            Selections: {
                "Taft": 1
            }
        }
    ]
};


var hfc = require("hfc");

console.log(" **** starting NetVote ****");

// get the addresses from the docker-compose environment
var PEER_ADDRESS         = "192.168.99.101:7051";
var MEMBERSRVC_ADDRESS   = "192.168.99.101:7054";


var chaincodeID = "netvote";
var chain = hfc.newChain("voter-client");
chain.setKeyValStore( hfc.newFileKeyValStore('/tmp/keyValStore') );
console.log("member services address ="+MEMBERSRVC_ADDRESS);
console.log("peer address ="+PEER_ADDRESS);
chain.setMemberServicesUrl("grpc://"+MEMBERSRVC_ADDRESS);
chain.addPeer("grpc://"+PEER_ADDRESS);

chain.enroll(userName, secret, function(err, user) {
    if (err) {
        console.log("ERROR: failed to enroll user: %s",err);
        process.exit(1);
    }

    castVote(user, VOTE, function(r){
        getResults(user, DECISION);
    });
});

function toString(obj){
    return JSON.stringify(obj);
}


function castVote(user, vote, callback){
    invoke_chaincode(user, "cast_votes", [ toString(vote) ], callback);
}

function createVoter(user, voter, callback){
    console.log("creating voter: "+voter.Id);
    invoke_chaincode(user, "add_voter", [ toString(voter) ], callback);
}

function createDecision(user, decision, callback){
    console.log("creating decision: "+decision.Id);
    invoke_chaincode(user, "add_decision", [ toString(decision) ], callback);
}

function getResults(user, decision){
    console.log("getting results: "+decision.Id);
    query_chaincode(user, "get_decision", [toString(decision)])
}

// Query chaincode
function query(user) {
    query_chaincode(user, "query", ["a"])
}

function invoke(user){
    invoke_chaincode(user, "invoke", ["a", "b", "1"])
}

function query_chaincode(user, func, args) {
    console.log("querying chaincode: "+func+" "+args);
    // Construct a query request
    var queryRequest = {
        chaincodeID: chaincodeID,
        fcn: func,
        args: args,
        attrs: ["group"]
    };
    // Issue the query request and listen for events
    var tx = user.query(queryRequest);
    tx.on('complete', function (results) {
        console.log("Election results: "+results.result);
        process.exit(0);
    });
    tx.on('error', function (error) {
        console.log("Failed to query chaincode: request=%j, error=%k",queryRequest,error);
        process.exit(1);
    });
}

//Invoke chaincode
function invoke_chaincode(user, func, args, callback) {
    console.log("invoke chaincode: "+func+" "+args);
    // Construct a query request
    var invokeRequest = {
        chaincodeID: chaincodeID,
        fcn: func,
        args: args,
        attrs: ["group"]
    };
    // Issue the invoke request and listen for events
    var tx = user.invoke(invokeRequest);
    tx.on('submitted', function (results) {
        console.log("invoke submitted successfully; results=%j",results);
    });
    tx.on('complete', function (results) {
        console.log("invoke completed successfully; results=%j",results);
        callback(results);
    });
    tx.on('error', function (error) {
        console.log("Failed to invoke chaincode: request=%j, error=%k",invokeRequest,error);
        process.exit(1);
    });
}
