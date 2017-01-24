'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");
var uuidV4 = require('uuid/v4');


var createBallot = function(account, ballot, callback, errorCallback){
    nvlib.invokeChaincode("add_ballot", ballot, account.enrollment_id, function(){
        var obj = {
            "id": account.account_id+":"+ballot.Ballot.Id,
            "payload": new Buffer(JSON.stringify(ballot)).toString("base64"),
            "requires2FA": (true === ballot.Ballot["Requires2FA"]),
            "tenantId": account.account_id,
            "owner": account.user
        };
        nvlib.saveDynamoItem("ballots", obj, errorCallback, callback);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballot = JSON.parse(event.body);

    //generate ID
    ballot.Ballot["Id"] = uuidV4();

    nvlib.chainInit(event, context, function(account) {
        createBallot(account, ballot, function () {
            nvlib.handleSuccess({"ballotId": ballot.Ballot.Id}, callback);
        }, function(err){
            nvlib.handleError(err, callback)
        });
    });
};