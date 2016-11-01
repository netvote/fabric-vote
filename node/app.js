/**
 * This code is all over the place - and is mostly a scratch spce to get things working
 */


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
var PEER_ADDRESS         = process.env.CORE_PEER_ADDRESS;
var MEMBERSRVC_ADDRESS   = process.env.MEMBERSRVC_ADDRESS;

var chain, chaincodeID;

chain = hfc.newChain("voter-client");

chain.setKeyValStore( hfc.newFileKeyValStore('/tmp/keyValStore') );
console.log("member services address ="+MEMBERSRVC_ADDRESS);
console.log("peer address ="+PEER_ADDRESS);
chain.setMemberServicesUrl("grpc://"+MEMBERSRVC_ADDRESS);
chain.addPeer("grpc://"+PEER_ADDRESS);

var mode =  process.env['DEPLOY_MODE'];
console.log("DEPLOY_MODE=" + mode);
if (mode === 'dev') {
    chain.setDevMode(true);
    //Deploy will not take long as the chain should already be running
    chain.setDeployWaitTime(10);
} else {
    chain.setDevMode(false);
    //Deploy will take much longer in network mode
    chain.setDeployWaitTime(120);
}


chain.setInvokeWaitTime(10);

// Begin by enrolling the user
enroll();

// Enroll a user.
function enroll() {
    console.log("enrolling user admin ...");
    // Enroll "admin" which is preregistered in the membersrvc.yaml
    chain.enroll("admin", "Xurw3yU9zI0l", function(err, admin) {
        if (err) {
            console.log("ERROR: failed to register admin: %s",err);
            process.exit(1);
        }
        // Set this user as the chain's registrar which is authorized to register other users.
        chain.setRegistrar(admin);

        var userName = "slanders";
        // registrationRequest
        var registrationRequest = {
            enrollmentID: userName,
            affiliation: "company_a",
            attributes: [
                { name: "group", value: "sales" }
            ]
        };
        chain.registerAndEnroll(registrationRequest, function(error, user) {
            if (error) throw Error(" Failed to register and enroll " + userName + ": " + error);
            console.log("Enrolled %s successfully\n", userName);
            console.log(user);
            deploy(user);
        });
    });
}

// Deploy chaincode
function deploy(user) {
    console.log("deploying chaincode; please wait ...");
    // Construct the deploy request
    var deployRequest = {
        chaincodeName: process.env.CORE_CHAINCODE_ID_NAME,
        fcn: "init",
        args: []
    };

    var tx = user.deploy(deployRequest);
    tx.on('complete', function(results) {
        console.log("deploy complete; results: %j",results);
        chaincodeID = results.chaincodeID;
        doVote(user);
    });
    tx.on('error', function(error) {
        console.log("Failed to deploy chaincode: request=%j, error=%k",deployRequest,error);
        process.exit(1);
    });
}

function toString(obj){
    return JSON.stringify(obj);
}


function doVote(user){
    createDecision(user, DECISION, function(r){
        createVoter(user, VOTER, function(r){
            castVote(user, VOTE, function(r){
                getResults(user, DECISION);
            });
        });
    });
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
        attrs: ["role"]
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
