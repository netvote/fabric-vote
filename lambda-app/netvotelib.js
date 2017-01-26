var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');
var crypto = require('crypto');
var twilio = require('twilio');

var CHAINCODE_ID = "";
var CHAIN_HOSTNAME = "";
var CHAIN_PORT = 80;


module.exports = {

    nvInit: function(event, context, callback, errorCallback){
        initNetvote(event, callback, errorCallback);
    },

    chainInit: function(event, context, callback, errorCallback){
        initNetvote(event, function(account){
                enroll(account.enrollment_id, account.enrollment_secret, function () {
                    callback(account);
                }, errorCallback);
        }, errorCallback);
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

    queryDynamoItems: function(table, key, value, errorCallback, callback){
        queryDynamoDocs(table, key, value, errorCallback, callback);
    },

    saveDynamoItem: function(table, obj, errorCallback, callback){
        insertDynamoDoc(table, obj, errorCallback, callback);
    },

    removeDynamoItem: function(table, key, value, errorCallback, callback){
        deleteDynamoItem(table, key, value, errorCallback, callback);
    },

    sendSms: function(phoneNumbers, message, callback, errorCallback){
        this.getDynamoItem("config","id","twilio",function(err){
            errorCallback(err)
        }, function(data) {
            var sid = data.Item.sid;
            var token = data.Item.token;
            var client = twilio(sid, token);
            var fromPhone = data.Item.phone;

            for(var i=0; i<phoneNumbers.length; i++) {
                client.messages.create({
                    body: message,
                    to: phoneNumbers[i],
                    from: fromPhone
                }, function (err) {
                    if (err) {
                        errorCallback(err)
                    }
                });
            }
            callback({"message": "sms sent"});
        });
    },

    handleSuccess: function(obj, callback){
        var respObj = {
            "statusCode": 200,
            "headers": {
                "Access-Control-Allow-Origin":"*"
            },
            "body": JSON.stringify(obj)
        };
        callback(null, respObj);
    },

    handleUnauthorized: function(callback){
        var respObj = {
            "statusCode": 401,
            "headers": {
                "Access-Control-Allow-Origin":"*"
            },
            "body": JSON.stringify({"status":"unauthorized"})
        };
        callback(null, respObj);
    },

    handleNotFound : function(callback){
        var respObj = {
            "statusCode": 404,
            "headers": {
                "Access-Control-Allow-Origin":"*"
            },
            "body": JSON.stringify({"error": "not found"})
        };
        callback(null, respObj);
    },

    handleError : function(e, callback){
        var respObj = {
            "statusCode": 500,
            "headers": {
                "Access-Control-Allow-Origin":"*"
            },
            "body": JSON.stringify({"error":e})
        };
        callback(null, respObj);
    }
};

var initNetvote = function(event, callback, errorCallback) {
        var apiKey = event.requestContext.identity.apiKey;

        getDyanmoDoc("config", "id", "chaincode", errorCallback, function (data) {
            CHAINCODE_ID = data.Item.version;
            CHAIN_HOSTNAME = data.Item.hostname;
            CHAIN_PORT = data.Item.port;
            getApiCredentials(apiKey, errorCallback, function (data) {
                if (data.Item == undefined) {
                    errorCallback("apiKey not found")
                } else {
                    var account = data.Item;
                    if(event.requestContext.authorizer != undefined) {
                        account["user"] = event.requestContext.authorizer.claims.sub;
                    }else{
                        account["user"] = "UNKNOWN";
                    }
                    callback(data.Item);
                }
            });
        });
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

var deleteDynamoItem = function(table, key, value, errorCallback, callback){
    var params = {
        TableName: table,
        Key:{}
    };
    params.Key[key] = value;

    dynamo.deleteItem(params, function(err) {
        if (err) {
            console.log("delete error!");
            errorCallback(err);
        } else {
            console.log("delete success!");
            callback();
        }
    });

    console.log("dynamo payload: "+JSON.stringify(params));
};


var enroll = function(enrollmentId, enrollmentSecret, callback, errorCallback){
    console.log("Enrolling..."+enrollmentId);
    var loginBody  = {
        "enrollId": enrollmentId,
        "enrollSecret": enrollmentSecret
    };
    var postData = JSON.stringify(loginBody);
    postRequest("/registrar", postData, callback, errorCallback);
};


var queryDynamoDocs = function(table, key, value, errorCallback, callback){
    var params = {
        TableName: table,
        IndexName: key+"-index",
        KeyConditions: [ dynamo.Condition(key, "EQ", value) ]
    };

    dynamo.query(params, function(err, data) {
        if (err) {
            console.error("Unable to read item. Error JSON:", JSON.stringify(err, null, 2));
            errorCallback(err);
        } else{
            callback(data);
        }
    });
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
    getDyanmoDoc("accounts", "api_key", apiKey, errorCallback, callback);
};