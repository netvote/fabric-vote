'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');


//TODO: fix this hack
var getBallot = function(voterId, enrollmentId, callback, errorCallback){
    nvlib.invokeChaincode("init_voter", {Id: voterId}, enrollmentId, function(){
        setTimeout(function() {
            nvlib.queryChaincode("get_ballot", {Id: voterId}, enrollmentId, function(ballot){
                setTimeout(function() {
                    nvlib.queryChaincode("get_ballot", {Id: voterId}, enrollmentId, callback, errorCallback);
                }, 500);
            }, errorCallback);
        }, 500);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var voterId = event.pathParameters.voterId;

    nvlib.chainInit(event, context, function(chaincodeUser){

        getBallot(voterId, chaincodeUser.enrollment_id, function(ballot){
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