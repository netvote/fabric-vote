'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    nvlib.nvInit(event, context, function(account) {

        nvlib.queryDynamoDocs("ballots", "owner", account.user,
            function(e){
                nvlib.handleError(e, callback)
            },
            function(data){
                if(data == undefined || data.Items == undefined){
                    nvlib.handleNotFound(callback);
                }else {
                    var ballots = data.Items;
                    var result = [];

                    for(var i=0; i<ballots.length; i++){
                        result.push(JSON.parse(new Buffer(ballots[i].payload, 'base64').toString("ascii")))
                    }

                    nvlib.handleSuccess(result, callback);
                }
            }
        );

    });
};