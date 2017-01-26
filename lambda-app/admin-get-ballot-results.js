'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');

var getResults = function(enrollmentId, ballotId, callback, errorCallback){
    nvlib.queryChaincode("get_ballot_results", {Id: ballotId}, enrollmentId, callback, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.chainInit(event, context, function(account){

        nvlib.getDynamoItem("ballots","id", account.account_id+":"+ballotId, function(err){
            nvlib.handleError(err, callback)
        }, function(data){
            if(data == undefined || data.Item == undefined) {
                nvlib.handleNotFound(callback);
            }else if (data.Item.owner != account.user){
                nvlib.handleUnauthorized(callback);
            }else{
                getResults(account.enrollment_id, ballotId, function(results){
                    var resultObj = JSON.parse(results.result.message);
                    if(resultObj.Id != ""){
                        nvlib.handleSuccess(resultObj, callback);
                    }else{
                        nvlib.handleNotFound(callback);
                    }
                }, function(e){
                    nvlib.handleError(e, callback);
                });
            }
        });
    }, function(e) {
        nvlib.handleError(e, callback)
    });
};