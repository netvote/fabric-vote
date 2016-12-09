'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');

var getResults = function(enrollmentId, decisionId, callback, errorCallback){
    nvlib.queryChaincode("query", "get_results", {Id: decisionId}, enrollmentId, callback, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var decisionId = event.pathParameters.decisionId;

    nvlib.chainInit(event, context, function(chaincodeUser){
        getResults(chaincodeUser.enrollment_id, decisionId, function(results){
            nvlib.handleSuccess(JSON.parse(results.result.message), callback);
        }, function(e){
            handleError(e, callback);
        });

    }, function(e) {
        nvlib.handleError(e, callback)
    });
};