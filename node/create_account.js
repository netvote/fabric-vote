var hfc = require("hfc");
var express = require('express');
var uuid = require('node-uuid');
var app = express();

var MEMBERSRVC_ADDRESS   = process.env.MEMBERSRVC_ADDRESS;
var chain;
chain = hfc.newChain("voter-client");
chain.setKeyValStore( hfc.newFileKeyValStore('/tmp/keyValStore') );
console.log("member services address ="+MEMBERSRVC_ADDRESS);
chain.setMemberServicesUrl("grpc://"+MEMBERSRVC_ADDRESS);


var admin_user =  process.env['ADMIN_USER'];
var admin_pass =  process.env['ADMIN_PASS'];

var registerUser = function(accountId, role, callback){
    var enrollmentId = role+"_"+accountId;
    var registrationRequest = {
        enrollmentID: enrollmentId,
        affiliation: "netvote",
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


app.get('/', function (req, res) {

    chain.enroll(admin_user, admin_pass, function(err, admin) {
        if (err) {
            console.log("ERROR: failed to register admin: %s", err);
            process.exit(1);
        }
        // Set this user as the chain's registrar which is authorized to register other users.
        chain.setRegistrar(admin);

        var accountId = uuid.v4().replace(/-/g, "_");

        result = {};

        registerUser(accountId, "voter", function (voterEnrollId, voterSecret) {
            registerUser(accountId, "admin", function (adminEnrollId, adminSecret) {
                result = {
                    "voter": {
                        "enrollId": voterEnrollId,
                        "secret": voterSecret
                    },
                    "admin": {
                        "enrollId": adminEnrollId,
                        "secret": adminSecret
                    }
                };
                res.json(result);
            });
        });
    })
});

app.listen(8000, function () {
    console.log('Started server on port 8000')
});

process.on('uncaughtException', function (err) {
    console.log(err);
});