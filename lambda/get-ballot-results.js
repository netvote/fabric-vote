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

    nvlib.chainInit(event, context, function(chaincodeUser){
        getResults(chaincodeUser.enrollment_id, ballotId, function(results){
            var resultObj = JSON.parse(results.result.message);
            if(resultObj.Id != ""){
                nvlib.handleSuccess(resultObj, callback);
            }else{
                nvlib.handleNotFound(callback);
            }
        }, function(e){
            nvlib.handleError(e, callback);
        });

    }, function(e) {
        nvlib.handleError(e, callback)
    });
};