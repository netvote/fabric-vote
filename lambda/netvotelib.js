var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');
var crypto = require('crypto');

var CHAINCODE_ID = "";
var CHAIN_HOSTNAME = "";
var CHAIN_PORT = 80;


module.exports = {

    chainInit: function(event, callback, errorCallback){
        var apiKey = event.requestContext.identity.apiKey;

        getDyanmoDoc("config","id","chaincode",errorCallback, function(data){
            CHAINCODE_ID = data.Item.version;
            CHAIN_HOSTNAME = data.Item.hostname;
            CHAIN_PORT = data.Item.port;
            getApiCredentials(apiKey, errorCallback, function(data){
                if(data.Item == undefined){
                    errorCallback("apiKey not found")
                }else {
                    var enrollmentId = data.Item.enrollment_id;
                    var enrollmentSecret = data.Item.enrollment_secret;
                    enroll(enrollmentId, enrollmentSecret, function () {
                        callback(data.Item);
                    }, errorCallback);
                }
            });
        })
    },

    hash256: function(str){
        const keyhash = crypto.createHash('sha256');
        keyhash.update(str);
        return keyhash.digest('hex');
    },

    invokeChaincode: function (operation, payload, callback, enrollmentId, errorCallback) {
        callChaincode("invoke", operation, payload, callback, enrollmentId, errorCallback);
    },

    queryChaincode: function (operation, payload, callback, enrollmentId, errorCallback) {
        callChaincode("query", operation, payload, callback, enrollmentId, errorCallback);
    },

    getDynamoItem: function(table, key, value, errorCallback, callback){
        getDyanmoDoc(table, key, value, errorCallback, callback);
    },

    saveDynamoItem: function(table, obj, errorCallback, callback){
        insertDynamoDoc("ballots", obj, errorCallback, callback);
    },

    handleSuccess: function(obj, callback){
        var respObj = {
            "statusCode": 200,
            "headers": {},
            "body": JSON.stringify(obj)
        };
        callback(null, respObj);
    },

    handleUnauthorized: function(callback){
        var respObj = {
            "statusCode": 401,
            "headers": {},
            "body": JSON.stringify({"status":"unauthorized"})
        };
        callback(null, respObj);
    },

    handleError : function(e, callback){
        var respObj = {
            "statusCode": 500,
            "headers": {},
            "body": JSON.stringify({"error":e})
        };
        callback(null, respObj);
    }
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


var enroll = function(enrollmentId, enrollmentSecret, callback, errorCallback){
    console.log("Enrolling...")
    var loginBody  = {
        "enrollId": enrollmentId,
        "enrollSecret": enrollmentSecret
    };
    var postData = JSON.stringify(loginBody);
    postRequest("/registrar", postData, callback, errorCallback);
};



var getDyanmoDoc = function(table, key, value, errorCallback, callback){
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

var callChaincode = function(method, operation, payload, enrollmentId, callback, errorCallback){

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
            "secureContext": enrollmentId
        },
        "id": correlationId
    });

    console.log("postData: "+postData);
    postRequest("/chaincode", postData, callback, errorCallback);
};

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

var getApiCredentials = function(apiKey, errorCallback, callback){
    getItemFromDyanmoDB("accounts", "api_key", apiKey, errorCallback, callback);
};