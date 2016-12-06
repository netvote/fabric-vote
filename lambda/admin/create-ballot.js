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

var insertDynamoDoc = function(table, obj, errorCallback, callback){
    var params = {
        TableName: table,
        Item: obj
    };
    console.log("dynamo payload: "+JSON.stringify(params));
    dynamo.putItem(params, function(err) {
        if (err) {
            console.log("insert error!");
            errorCallback(err);
        } else {
            console.log("insert success!");
            callback();
        }
    });
};

var saveBallotToDynamo = function(ballot, errorCallback, callback){
    var obj = {
        "id": ballot.Ballot.Id,
        "payload": new Buffer(JSON.stringify(ballot)).toString("base64"),
        "requires2FA": (true === ballot.Ballot["Requires2FA"])
    };
    insertDynamoDoc("ballots", obj, errorCallback, callback);
};

var enroll = function(enrollmentId, enrollmentSecret, callback, errorCallback){
    var loginBody  = {
        "enrollId": enrollmentId,
        "enrollSecret": enrollmentSecret
    };
    var postData = JSON.stringify(loginBody);
    postRequest("/registrar", postData, callback, errorCallback);
};


var createBallot = function(enrollmentId, ballot, callback, errorCallback){
    invokeChaincode("invoke", "add_ballot", ballot, enrollmentId, callback, errorCallback);
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
    var errorHandler = function(err){
        handleError(err, callback);
    }

    getDynamoItem("config","id","chaincode",errorHandler,
        function(data) {
            CHAINCODE_ID = data.Item.version;
            CHAIN_HOSTNAME = data.Item.hostname;
            CHAIN_PORT = data.Item.port;

            getApiCredentials(apiKey,  errorHandler,
                function (data) {
                    var enrollmentId = data.Item.enrollment_id;
                    var enrollmentSecret = data.Item.enrollment_secret;
                    var ballot = JSON.parse(event.body);

                    enroll(enrollmentId, enrollmentSecret, function () {
                        console.log("enroll success");
                        createBallot(enrollmentId, ballot, function (result) {
                            saveBallotToDynamo(ballot, errorHandler,
                                function(){
                                    console.log("createBallot success: " + JSON.stringify(result));

                                    var respObj = {
                                        "statusCode": 200,
                                        "headers": {},
                                        "body": JSON.stringify({"result": "success"})
                                    };

                                    callback(null, respObj);
                                });
                        }, errorHandler);

                    }, errorHandler);
                });
        });
};