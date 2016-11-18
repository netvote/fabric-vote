'use strict';
var CHAINCODE_ID = "91afb32cd2189691357b6a7d25d4edc6b15b3079b4e43be135c2cf09d13f37d8";
console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');

var getChaincodeId = function(){
    return CHAINCODE_ID;
};

var postRequest = function(urlPath, postData, callback, errorCallback){
    var options = {
        hostname: 'peer.stevenlanders.net',
        port: 80,
        path: urlPath,
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData)
        }
    };

    var req = http.request(options, function(res){
        var body = '';
        res.setEncoding('utf8');

        res.on('data', function (chunk) {
            body += chunk;
        });

        res.on('end', function(){
            callback(JSON.parse(body));
        });
    });

    req.on('error', function(e){
        console.log("error response:"+e.message);
        errorCallback(e);
    });

    // write data to request body
    req.write(postData);
    req.end();
};

var enroll = function(enrollmentId, enrollmentSecret, callback, errorCallback){
    console.log("Enrolling...")
    var loginBody  = {
        "enrollId": enrollmentId,
        "enrollSecret": enrollmentSecret
    };
    var postData = JSON.stringify(loginBody);
    postRequest("/registrar", postData, callback, errorCallback);
};


var getResults = function(enrollmentId, decisionId, callback, errorCallback){
    invokeChaincode("query", "get_results", {Id: decisionId}, enrollmentId, callback, errorCallback);
};

var invokeChaincode = function(method, operation, payload, secureContext, callback, errorCallback){

    var timeMs = new Date().getTime();
    var randomNumber = Math.floor(Math.random()*100000);
    var correlationId = parseInt(""+timeMs+""+randomNumber);

    var postData = JSON.stringify({
        "jsonrpc": "2.0",
        "method":method,
        "params": {
            "chaincodeID": {
                "name" : getChaincodeId()
            },
            "ctorMsg": {
                "args":[operation, JSON.stringify(payload)]
            },
            "attributes": ["role","account_id"],
            "secureContext": secureContext
        },
        "id": correlationId
    });

    console.log("postData: "+postData);
    postRequest("/chaincode", postData, callback, errorCallback);
};

var handleError = function(e, callback){
    var respObj = {
        "statusCode": 500,
        "headers": {},
        "body": JSON.stringify({"error":e})
    };
    callback(null, respObj);
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var apiKey = event.requestContext.identity.apiKey;

    var params = {
        TableName: "accounts",
        Key:{
            "api_key": apiKey
        }
    };

    dynamo.getItem(params, function(err, data) {
        var respObj = {};
        if (err) {
            console.error("Unable to read item. Error JSON:", JSON.stringify(err, null, 2));
            handleError(err, callback);
        } else {
            console.log("GetItem succeeded:", JSON.stringify(data, null, 2));

            var enrollmentId = data.Item.enrollment_id;
            var enrollmentSecret = data.Item.enrollment_secret;
            var decisionId = event.pathParameters.decisionId;

            enroll(enrollmentId, enrollmentSecret, function(){
                console.log("enroll success");
                getResults(enrollmentId, decisionId, function(results){
                    console.log("getResults success: "+JSON.stringify(results));

                    respObj = {
                        "statusCode": 200,
                        "headers": {},
                        "body": results.result.message
                    };

                    callback(null, respObj);
                }, function(e){
                    handleError(e, callback);
                });

            }, function(e){
                handleError(e, callback);
            });
        }
    });

};
