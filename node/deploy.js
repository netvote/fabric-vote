/**
 * This code is all over the place - and is mostly a scratch spce to get things working
 */

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
var deploy_user =  process.env['DEPLOY_USER'];
var deploy_pass =  process.env['DEPLOY_PASS'];

if (mode === 'dev') {
    chain.setDevMode(true);
    chain.setDeployWaitTime(10);
} else {
    chain.setDevMode(false);
    chain.setDeployWaitTime(120);
}

chain.setInvokeWaitTime(10);

// Begin by enrolling the user
enroll();

// Enroll a user.
function enroll() {
    console.log("enrolling user "+deploy_user+" ...");
    // Enroll "admin" which is preregistered in the membersrvc.yaml
    chain.enroll(deploy_user, deploy_pass, function(err, admin) {
        if (err) {
            console.log("ERROR: failed to register admin: %s",err);
            process.exit(1);
        }
        // Set this user as the chain's registrar which is authorized to register other users.
        chain.setRegistrar(admin);
	deploy(admin);
    });
}

// Deploy chaincode
function deploy(user) {
    console.log("deploying chaincode; please wait ...");
    // Construct the deploy request
    var deployRequest = {
        chaincodeName: "netvote",
        chaincodePath: "netvote/go/chaincode/",
        fcn: "init",
        args: [],
        confidential: false
    };

    var tx = user.deploy(deployRequest);
    tx.on('complete', function(results) {
        console.log(JSON.stringify(results));
    });
    tx.on('error', function(error) {
        console.log("Failed to deploy chaincode: request=%j, error=%k",deployRequest,error);
        process.exit(1);
    });
}

function toString(obj){
    return JSON.stringify(obj);
}


