'use strict';

console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');

var enrollMock = function(enrollmentId, enrollmentSecret, callback, errorCallback){
    console.log('enroll mock:'+enrollmentId+"/"+enrollmentSecret);
    callback();
};

var postRequest = function(urlPath, postData, callback, errorCallback){
    var options = {
        hostname: 'url.to.peer',
        port: 7050,
        path: urlPath,
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData)
        }
    };

    var req = http.request(options, function(res){
        res.setEncoding('utf8');
        res.on('end', function(){
            callback();
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

var getBallotMock = function(enrollmentId, voterId, callback, errorCallback){
    console.log('getBallot mock: context='+enrollmentId+", voter="+voterId);
    var mockObj = [{"Id":"1912-us-president","Name":"president","BallotId":"","Options":["Taft","Bryan"],"ResponsesRequired":1,"RepeatVoteDelayNS":0,"Repeatable":false},
        {"Id":"1912-ga-governor","Name":"governor","BallotId":"","Options":["Mark","Sarah"],"ResponsesRequired":1,"RepeatVoteDelayNS":0,"Repeatable":false}];
    callback(mockObj);
};

var getBallot = function(enrollmentId, voterId, callback, errorCallback){
    invokeChaincode("init_voter", {Id: voterId}, enrollmentId, callback, errorCallback);
};

var invokeChaincode = function(operation, payload, secureContext, callback, errorCallback){
    var postData = JSON.stringify({
        "jsonrpc": "2.0",
        "method":"invoke",
        "params": {
            "chaincodeID": {
                "name" :"netvote"
            },
            "ctorMsg": {
                "args":[operation, JSON.stringify(payload)]
            },
            "attributes": ["role","account_id"],
            "secureContext": secureContext
        },
        "id": 2
    });

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

            enrollMock(enrollmentId, enrollmentSecret, function(){
                //success
                getBallotMock(enrollmentId, voterId, function(ballot){
                    respObj = {
                        "statusCode": 200,
                        "headers": {},
                        "body": JSON.stringify(ballot)
                    };
                    callback(null, respObj);
                }, function(e){
                    handleError(e, callback);
                });

                callback(null, respObj);
            }, function(e){
                handleError(e, callback);
            });
        }
    });

};
