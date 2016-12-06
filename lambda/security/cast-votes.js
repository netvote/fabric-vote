'use strict';

var CHAINCODE_ID = "";
var CHAIN_HOSTNAME = "";
var CHAIN_PORT = 80;

console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');
var crypto = require('crypto');

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

var handleUnauthorized = function(message, callback){
    var respObj = {
        "statusCode": 401,
        "headers": {},
        "body": JSON.stringify({"error": message})
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


var verifyTwoFactor = function(voterBallot, voterId, accountId, twoFactorCode, errorCallback, callback){
    getDynamoItem("ballots", "id", voterBallot.Id, errorCallback, function(data){
        if(data.Item.requires2FA){
            const keyhash = crypto.createHash('sha256');
            keyhash.update(accountId+":"+voterId);

            getDynamoItem("ballot_sms_codes", "id", keyhash.digest('hex'), errorCallback, function(data){

                var currentTime = new Date().getTime();
                console.log(JSON.stringify(data));
                console.log("code="+twoFactorCode+", date="+currentTime);

                if(data.Item.expiration < currentTime){
                    callback("expired");
                } if(twoFactorCode && data.Item.code.toString() == twoFactorCode){
                    callback("success");
                } else{
                    callback("fail");
                }
            });
        } else{
            callback("success");
        }
    });
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var apiKey = event.requestContext.identity.apiKey;

    getDynamoItem("config","id","chaincode",function(err){
        handleError(err, callback);
    }, function(data) {
        CHAINCODE_ID = data.Item.version;
        CHAIN_HOSTNAME = data.Item.hostname;
        CHAIN_PORT = data.Item.port;

        getApiCredentials(apiKey, function(err){
            handleError(err, callback);
        }, function(data){
            console.log("GetItem succeeded:", JSON.stringify(data, null, 2));

            var enrollmentId = data.Item.enrollment_id;
            var enrollmentSecret = data.Item.enrollment_secret;
            var voterId = event.pathParameters.voterid;

            var voterballot = JSON.parse(event.body);
            var votes = voterballot.VoterBallot;
            var accountId = data.Item.account_id;

            var twoFactorCode = event.headers["nv-two-factor-code"];

            verifyTwoFactor(voterballot, voterId, accountId, twoFactorCode, function(){ handleError(err, callback)}, function(result){
                if(result == "success"){
                    enroll(enrollmentId, enrollmentSecret, function(){
                        console.log("enroll success");
                        castVotes(enrollmentId, {"VoterId": voterId, "Decisions": votes}, function(result){
                            console.log("castVote success: "+JSON.stringify(result));

                            var respObj = {
                                "statusCode": 200,
                                "headers": {},
                                "body": JSON.stringify({"result":"success"})
                            };

                            console.log("cast vote success");

                            callback(null, respObj);
                        }, function(e){
                            handleError(e, callback);
                        });

                    }, function(e){
                        handleError(e, callback);
                    });
                }else{
                    handleUnauthorized("two-factor failed: "+result, callback)
                }
            })
        });
    });
};
