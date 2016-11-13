//TODO: figure out how this should be invoked.  It needs to be somewhere close to chaincode

var hfc = require("hfc");
var uuid = require('node-uuid');
var AWS = require('aws-sdk');


console.log(" **** registering account  ****");

// get the addresses from the docker-compose environment
var MEMBERSRVC_ADDRESS   = process.env.MEMBERSRVC_ADDRESS;
var chain;
chain = hfc.newChain("voter-client");
chain.setKeyValStore( hfc.newFileKeyValStore('/tmp/keyValStore') );
console.log("member services address ="+MEMBERSRVC_ADDRESS);
chain.setMemberServicesUrl("grpc://"+MEMBERSRVC_ADDRESS);


var registerUser = function(accountId, role, callback){
    var enrollmentId = role+"-"+accountId;
    var registrationRequest = {
        enrollmentID: enrollmentId,
        affiliation: "bank_a",
        attributes: [
            { name: "account_id", value: accountId },
            { name: "role", value: role}
        ]
    };

    chain.register(registrationRequest, function(error, secret) {
        if (error) throw Error(" Failed to register and enroll " + secret + ": " + error);
        console.log("Registered %s successfully", enrollmentId);
        callback(enrollmentId, secret);
    });
};

var provisionAPI = function(accountId, enrollmentId, secret){
    var apigateway = new AWS.APIGateway();
    var params = {
        description: 'api key for '+enrollmentId,
        enabled: true,
        generateDistinctId: true,
        name: enrollmentId
    };
    apigateway.createApiKey(params, function(err, data) {
        if (err){
            console.log(err, err.stack)
        }
        else{
            console.log(data);
            //TODO: good way to get usage plan id?
            var params = {
                keyId: key.id, /* required */
                keyType: 'API_KEY', /* required */
                usagePlanId: "458o87" /* required */
            };
            apigateway.createUsagePlanKey(params, function(err, data) {
                if (err) console.log(err, err.stack); // an error occurred
                else     console.log(data);           // successful response
            });
        }           // successful response
    });
};

//TODO: see if I can avoid hardcoding this password
chain.enroll("admin", "Xurw3yU9zI0l", function(err, admin) {
    if (err) {
        console.log("ERROR: failed to register admin: %s",err);
        process.exit(1);
    }
    // Set this user as the chain's registrar which is authorized to register other users.
    chain.setRegistrar(admin);

    var accountId = uuid.v1();

    registerUser(accountId, "voter", function(enrollmentId, secret){

        //TODO: store dynamoDB entry

        provisionAPI(accountId, enrollmentId, secret);

        registerUser(accountId, "admin", function(enrollmentId, secret){
            //do nothing for now
            console.log("Complete");
        })
    });

});

