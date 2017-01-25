'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");


var deleteBallot = function(enrollmentId, ballotId, callback, errorCallback){
    nvlib.invokeChaincode("delete_ballot", { Id: ballotId }, enrollmentId, function(){
        nvlib.removeDynamoItem("ballots", "id", ballotId, errorCallback, callback);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.chainInit(event, context, function(chaincodeUser) {

        deleteBallot(chaincodeUser.enrollment_id, ballotId, function () {
            nvlib.handleSuccess({"result": "success"}, callback);
        }, function(err){
            nvlib.handleError(err, callback)
        });

    });
};