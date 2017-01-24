'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');


//TODO: fix this hack...init ideally has been done prior to this (but when?)
var getBallot = function(voterId, ballotId, enrollmentId, callback, errorCallback){
    var operation = (ballotId == undefined) ? "get_decisions" : "get_ballot";
    nvlib.invokeChaincode("init_voter", {Id: voterId}, enrollmentId, function(){
        setTimeout(function() {
            nvlib.queryChaincode(operation, {BallotId: ballotId, VoterId: voterId}, enrollmentId, function(ballot){
                setTimeout(function() {
                    nvlib.queryChaincode(operation, {BallotId: ballotId, VoterId: voterId}, enrollmentId, callback, errorCallback);
                }, 500);
            }, errorCallback);
        }, 500);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var voterId = event.pathParameters.voterId;
    var ballotId = event.pathParameters.ballotId;

    nvlib.chainInit(event, context, function(account){

        getBallot(voterId, ballotId, account.enrollment_id, function(ballot){
            //result.message is returned as string.  Parsing so handleSuccess can stringify without quotes
            nvlib.handleSuccess(JSON.parse(ballot.result.message), callback);
        }, function(e){
            nvlib.handleError(e, callback);
        });

    },
    function(e){
        nvlib.handleError(e, callback);
    });
};