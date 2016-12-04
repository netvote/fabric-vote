'use strict';

var CHAINCODE_ID = "";
var CHAIN_HOSTNAME = "";
var CHAIN_PORT = 80;

console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');

var postRequest = function(urlPath, postData, callback, errorCallback){
    var options = {
        hostname: CHAIN_HOSTNAME,
        port: CHAIN_PORT,
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
                "name" : CHAINCODE_ID
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


var getDynamoItem = function(table, key, value, errorCallback, callback){
    var params = {
        TableName: table,
        Key:{}
    };
    params.Key[key] = value;

    dynamo.getItem(params, function(err, data) {
        if (err) {
            console.error("Unable to read item. Error JSON:", JSON.stringify(err, null, 2));
            errorCallback(err);
        } else{
            callback(data);
        }
    });
};


var getApiCredentials = function(apiKey, errorCallback, callback){
    getDynamoItem("accounts", "api_key", apiKey, errorCallback, callback);
};


exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var apiKey = event.requestContext.identity.apiKey;

    getDynamoItem("config","id","chaincode",function(err){
        handleError(err, callback);
    }, function(data){
        CHAINCODE_ID = data.Item.version;
        CHAIN_HOSTNAME = data.Item.hostname;
        CHAIN_PORT = data.Item.port;

        getApiCredentials(apiKey, function(err){
            handleError(err, callback);
        }, function(data){

            var enrollmentId = data.Item.enrollment_id;
            var enrollmentSecret = data.Item.enrollment_secret;
            var decisionId = event.pathParameters.decisionId;

            enroll(enrollmentId, enrollmentSecret, function(){
                console.log("enroll success");
                getResults(enrollmentId, decisionId, function(results){
                    console.log("getResults success: "+JSON.stringify(results));

                    var respObj = {
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
        });
    });
};
