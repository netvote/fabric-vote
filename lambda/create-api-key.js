'use strict';

console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');
var aws = require('aws-sdk');

var ADMIN_USAGE_PLAN = "a2i576";

var provisionAPI = function(enrollmentId, secret, usagePlanId, callback, errorCallback){
    var apigateway = new aws.APIGateway();
    var params = {
        description: 'api key for '+enrollmentId,
        enabled: true,
        generateDistinctId: true,
        name: enrollmentId
    };
    apigateway.createApiKey(params, function(err, data) {
        if (err){
            errorCallback(err);
        }
        else{
            console.log(data);
            var params = {
                keyId: data.id, /* required */
                keyType: 'API_KEY', /* required */
                usagePlanId: usagePlanId /* required */
            };
            var api_key = data.value;
            apigateway.createUsagePlanKey(params, function(err, data) {
                if (err){
                    errorCallback(err)
                } else{
                    callback(api_key)
                }
            });
        }
    });
};

var createAccount = function(callback, errorCallback){
    var options = {
        hostname: 'accounts.stevenlanders.net',
        path: "/"
    };

    var req_callback = function(response) {
        var str = '';

        //another chunk of data has been recieved, so append it to `str`
        response.on('data', function (chunk) {
            str += chunk;
        });

        //the whole response has been recieved, so we just print it out here
        response.on('end', function () {
            var obj = JSON.parse(str);
            console.log("create account callback with: "+JSON.stringify(obj));
            callback(obj);
        });
    };

    var req = http.request(options, req_callback);

    req.on('error', function(e){
        errorCallback(e);
    });

    req.end()
};


var handleError = function(e, callback){
    var respObj = {
        "statusCode": 500,
        "headers": {},
        "body": JSON.stringify({"error":e})
    };
    callback(null, respObj);
};

var insertApiRecord = function(cred, api_key, callback, errorCallback){

    var params = {
        TableName: "accounts",
        Item:{
            "enrollment_id": cred.enrollId,
            "enrollment_secret": cred.secret,
            "api_key": api_key,
            "account_id": cred.enrollId.replace("voter_","").replace("admin_",""),
            "role": cred.enrollId.substring(0, cred.enrollId.indexOf("_"))
        }
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

var createApiKey = function(cred, usagePlanId, callback, errorCallback){
    // actually create api key
    provisionAPI(cred.enrollId, cred.secret, usagePlanId, function(api_key){
        console.log("successfully added API KEY for "+cred.enrollId);
        insertApiRecord(cred, api_key, callback, errorCallback);
    }, errorCallback);
};


exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));


    createAccount(function(creds){
        console.log("create api account success: "+JSON.stringify(creds));

        createApiKey(creds.admin, ADMIN_USAGE_PLAN, function(){
            console.log("create api admin success");
            callback(null, {
                "statusCode": 200,
                "headers": {},
                "body": "success"
            });
        }, function(err){
            handleError(err, callback);
        });

    }, function(err){
        handleError(err, callback);
    });

};
