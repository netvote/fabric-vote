'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");


exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.nvInit(event, context, function(account) {

        var shareOptions = JSON.parse(event.body);

        var phones = shareOptions.sms ? shareOptions.sms : [];
        var email = shareOptions.email ? shareOptions.email : [];

        nvlib.getDynamoItem("ballots", "id", account.account_id+":"+ballotId,
            function(e){
                nvlib.handleError(e, callback)
            },
            function(data){
                if(data == undefined || data.Item == undefined) {
                    nvlib.handleNotFound(callback);
                }else if(data.Item.owner != account.user){
                    nvlib.handleUnauthorized(callback);
                }else {
                    var ballot = JSON.parse(new Buffer(data.Item.payload, 'base64').toString("ascii"));

                    nvlib.sendSms(phones, "Ballot: netvote://ballot/" + ballotId, function (result) {
                        nvlib.handleSuccess(result, callback);
                    }, function (err) {
                        nvlib.handleError(err, callback)
                    });
                }
            }
        );

    }, function(err){
        nvlib.handleError(err, callback)
    });
};