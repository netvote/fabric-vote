'use strict';

console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');

var postRequest = function(urlPath, postData, callback, errorCallback){
    var options = {
        hostname: 'peer.stevenlanders.net',
        port: 7050,
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
        errorCallback(e);
    });

    // write data to request body
    req.write(postData);
    req.end();
};

var enroll = function(enrollmentId, enrollmentSecret, callback, errorCallback){
    var loginBody  = {
        "enrollId": enrollmentId,
        "enrollSecret": enrollmentSecret
    };
    var postData = JSON.stringify(loginBody);
    postRequest("/registrar", postData, callback, errorCallback);
};


var castVotes = function(enrollmentId, votes, callback, errorCallback){
    invokeChaincode("invoke", "cast_votes", votes, enrollmentId, callback, errorCallback);
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
                "name" :"659cb5dcc3063054e4c90908050eebf68eb2bd193cc1520f1f2d198f0ff42268"
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
            var voterId = event.pathParameters.voterid;

            var votes = JSON.parse(event.body);

            enroll(enrollmentId, enrollmentSecret, function(){
                console.log("enroll success");
                castVotes(enrollmentId, {"VoterId": voterId, "Decisions": votes}, function(result){
                    console.log("castVote success: "+JSON.stringify(result));

                    respObj = {
                        "statusCode": 200,
                        "headers": {},
                        "body": "success"
                    };

                    console.log("cast vote success");

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
