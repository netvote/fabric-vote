'use strict';

var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var twilio = require('twilio');
var crypto = require('crypto');

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

var getCode = function(accountId, voterId, errorCallback, callback){
    var code = Math.floor(Math.random() * (999999 - 100000) + 100000);

    var now = new Date();
    var expiration = new Date(now.getTime()+(15*60*1000)).getTime();

    const keyhash = crypto.createHash('sha256');
    keyhash.update(accountId+":"+voterId);

    var sms_code = {
        "id": keyhash.digest('hex'),
        "code": code,
        "expiration": expiration
    };

    insertDynamoDoc("ballot_sms_codes", sms_code, errorCallback, function(){
        callback(code);
    });
};

exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    getDynamoItem("config","id","twilio",function(err){
        handleError(err, callback);
    }, function(data) {
        var sid = data.Item.sid;
        var token = data.Item.token;
        var client = twilio(sid, token);

        var toPhone = event.phone;
        var fromPhone = data.Item.phone;
        var voterId = event.voterId;
        var accountId = event.accountId;

        getCode(accountId, voterId, function(err){
            handleError(err, callback);
        }, function(code){
            client.messages.create({
                body: "Ballot Code: "+code,
                to: toPhone,
                from: fromPhone
            }, function(err) {
                if (err) {
                    handleError(err, callback);
                } else {
                    console.log("create api admin success");
                    callback(null, {
                        "statusCode": 200,
                        "headers": {},
                        "body": {"message": "sms sent"}
                    });
                }
            });
        });
    });

};
