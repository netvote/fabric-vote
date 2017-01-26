'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");


exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.nvInit(event, context, function(account) {

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
                    var ballotObj = JSON.parse(new Buffer(data.Item.payload, 'base64').toString("ascii"));
                    nvlib.handleSuccess(ballotObj, callback);
                }
            }
        );

    });
};