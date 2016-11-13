'use strict';

console.log('Loading function');
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();
var http = require('http');


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

    for(var i=0; i< event.enrollments.length; i++){
        var enrollment = event.enrollments[i];

    }
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
