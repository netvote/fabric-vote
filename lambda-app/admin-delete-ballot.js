'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");


var deleteBallot = function(account, ballotId, callback, errorCallback){
    nvlib.invokeChaincode("delete_ballot", { Id: ballotId }, account.enrollment_id, function(){
        nvlib.removeDynamoItem("ballots", "id", account.account_id+":"+ballotId, errorCallback, callback);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.chainInit(event, context, function(account) {

        nvlib.getDynamoItem("ballots","id", account.account_id+":"+ballotId, function(err){
            nvlib.handleError(err, callback)
        }, function(data){
            if(data == undefined || data.Item == undefined) {
                nvlib.handleNotFound(callback);
            }else if (data.Item.owner != account.user){
                nvlib.handleUnauthorized(callback);
            }else{
                deleteBallot(account, ballotId, function () {
                    nvlib.handleSuccess({"result": ballotId+" deleted"}, callback);
                }, function(err){
                    nvlib.handleError(err, callback)
                });
            }
        });
    });
};