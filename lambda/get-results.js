'use strict';

console.log('Loading function');
var nvlib = require('netvotelib');

var getResults = function(enrollmentId, decisionId, callback, errorCallback){
    nvlib.queryChaincode("get_results", {Id: decisionId}, enrollmentId, callback, errorCallback);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var decisionId = event.pathParameters.decisionId;

    nvlib.chainInit(event, context, function(chaincodeUser){
        getResults(chaincodeUser.enrollment_id, decisionId, function(results){
            var resultObj = JSON.parse(results.result.message);
            if(resultObj.Id != ""){
                nvlib.handleSuccess(resultObj, callback);
            }else{
                nvlib.handleNotFound(callback);
            }
        }, function(e){
            nvlib.handleError(e, callback);
        });

    }, function(e) {
        nvlib.handleError(e, callback)
    });
};