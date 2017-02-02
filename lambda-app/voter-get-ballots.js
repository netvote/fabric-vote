'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');


var getBallots = function(voterId, enrollmentId, callback, errorCallback){
    var operation = "get_voter_ballots";
    nvlib.queryChaincode(operation, {Id: voterId}, enrollmentId, callback, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    nvlib.chainInit(event, context, function(account){

            getBallots(account.user, account.enrollment_id, function(ballots){
                //result.message is returned as string.  Parsing so handleSuccess can stringify without quotes
                var ballotObj = JSON.parse(ballots.result.message);
                var result = [];
                for(var i =0; i< ballotObj.length; i++){

                    if(ballotObj[i].Id){
                        result.push(ballotObj[i]);
                    }
                }
                nvlib.handleSuccess(result, callback);
            }, function(e){
                nvlib.handleError(e, callback);
            });

        },
        function(e){
            nvlib.handleError(e, callback);
        });
};