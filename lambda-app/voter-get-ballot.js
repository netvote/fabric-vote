'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');


//TODO: fix this hack...init ideally has been done prior to this (but when?)
var getBallot = function(voterId, ballotId, enrollmentId, callback, errorCallback){
    nvlib.invokeChaincode("assign_ballot", {BallotId: ballotId, Voter: {Id:voterId}}, enrollmentId, function(){
        setTimeout(function() {
            nvlib.queryChaincode("get_ballot", {BallotId: ballotId, VoterId: voterId}, enrollmentId, function(ballot){
                setTimeout(function() {
                    nvlib.queryChaincode("get_ballot", {BallotId: ballotId, VoterId: voterId}, enrollmentId, callback, errorCallback);
                }, 500);
            }, errorCallback);
        }, 500);
    }, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.chainInit(event, context, function(account){

        nvlib.getDynamoItem("ballots", "id", account.account_id+":"+ballotId,
            function(e){
                nvlib.handleError(e, callback)
            },
            function(data){
                if(data == undefined || data.Item == undefined) {
                    nvlib.handleNotFound(callback);
                }else {
                    var ballotObj = JSON.parse(new Buffer(data.Item.payload, 'base64').toString("ascii"));
                    getBallot(account.user, ballotId, account.enrollment_id, function(ballot){
                        var decisions = JSON.parse(ballot.result.message);
                        var result = {
                            Ballot: ballotObj.Ballot,
                            Decisions: decisions
                        };
                        nvlib.handleSuccess(result, callback);
                    }, function(e){
                        nvlib.handleError(e, callback);
                    });
                }
            }
        );



    },
    function(e){
        nvlib.handleError(e, callback);
    });
};