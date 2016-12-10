'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");


var createBallot = function(accountId, enrollmentId, ballot, callback, errorCallback){
    nvlib.invokeChaincode("add_ballot", ballot, enrollmentId, function(){
        var obj = {
            "id": accountId+":"+ballot.Ballot.Id,
            "payload": new Buffer(JSON.stringify(ballot)).toString("base64"),
            "requires2FA": (true === ballot.Ballot["Requires2FA"])
        };
        nvlib.saveDynamoItem("ballots", obj, errorCallback, callback);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballot = JSON.parse(event.body);

    nvlib.chainInit(event, context, function(account) {
        createBallot(account.account_id, account.enrollment_id, ballot, function () {
            nvlib.handleSuccess({"result": "success"}, callback);
        }, function(err){
            nvlib.handleError(err, callback)
        });
    });
};