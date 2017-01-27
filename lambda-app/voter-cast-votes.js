'use strict';

console.log('Loading function');

var nvlib = require("netvotelib");

var castVotes = function(enrollmentId, votes, callback, errorCallback){
    nvlib.invokeChaincode("cast_votes", votes, enrollmentId, callback, errorCallback);
};

var verifyTwoFactor = function(voterBallot, voterId, accountId, twoFactorCode, errorCallback, callback){
    nvlib.getDynamoItem("ballots", "id", accountId+":"+voterBallot.BallotId, errorCallback, function(data){
        if(data.Item.requires2FA){

            var hashKey = nvlib.hash256(accountId+":"+voterId);

            nvlib.getDynamoItem("ballot_sms_codes", "id", hashKey, errorCallback, function(data){
                var currentTime = new Date().getTime();
                console.log(JSON.stringify(data));
                console.log("code="+twoFactorCode+", date="+currentTime+", lookup="+accountId+":"+voterId);
                if(data.Item == undefined){
                    callback("fail - no code");
                } else if(data.Item.expiration < currentTime){
                    callback("expired");
                } else if(twoFactorCode && data.Item.code.toString() == twoFactorCode){
                    callback("success");
                } else{
                    callback("fail - not valid");
                }
            });
        } else{
            callback("success");
        }
    });
};


exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;
    var twoFactorCode = event.headers["nv-two-factor-code"];
    var decisions = JSON.parse(event.body);

    nvlib.chainInit(event, context, function(account) {

        var enrollmentId = account.enrollment_id;
        var accountId = account.account_id;
        var voterId = account.user;

        var voterBallot = {
            "VoterId": voterId,
            "BallotId": ballotId,
            "Decisions": decisions
        };

        verifyTwoFactor(voterBallot, voterId, accountId, twoFactorCode,
            function (err) {
                nvlib.handleError(err, callback)
            },
            function (result) {
                if (result == "success") {
                    castVotes(enrollmentId, voterBallot, function (result) {
                        nvlib.handleSuccess({"result": "success"}, callback)
                    }, function (e) {
                        nvlib.handleError(e, callback);
                    });
                } else {
                    console.log("unauthorized = "+result);
                    nvlib.handleUnauthorized(callback);
                }
            }
        );
    });

};