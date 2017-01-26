'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");

var updateBallot = function(account, ballot, callback, errorCallback){
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
    var ballotId = event.pathParameters.ballotId;

    nvlib.chainInit(event, context, function(account) {
        nvlib.getDynamoItem("ballot","id", account.account_id+":"+ballotId, function(err){
            nvlib.handleError(err, callback)
        }, function(data){
            if(data == undefined || data.Item == undefined) {
                nvlib.handleNotFound(callback);
            }else if (data.Item.owner != account.user){
                nvlib.handleUnauthorized(callback);
            }else{
                updateBallot(account, ballot, function () {
                    nvlib.handleSuccess({"ballotId": ballot.Ballot.Id}, callback);
                }, function(err){
                    nvlib.handleError(err, callback)
                });
            }
        });
    });
};