'use strict';

var nvlib = require('netvotelib');

var getCode = function(accountId, voterId, errorCallback, callback){
    var code = Math.floor(Math.random() * (999999 - 100000) + 100000);
    var now = new Date();
    var expiration = new Date(now.getTime()+(15*60*1000)).getTime();

    var sms_code = {
        "id": nvlib.hash256(accountId+":"+voterId),
        "code": code,
        "expiration": expiration
    };

    nvlib.saveDynamoItem("ballot_sms_codes", sms_code, errorCallback, function(){
        callback(code);
    });
};


exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    nvlib.nvInit(event, context, function(account){
        var accountId =account.account_id;
        var voterId = account.user;
        var body = JSON.parse(event.body);
        var toPhone = body.phone;


        getCode(accountId, voterId, function(err){nvlib.handleError(err, callback);},
            function(code){
                nvlib.sendSms([toPhone], "Ballot Code: "+code, function (result) {
                    nvlib.handleSuccess({"message": "sms sent"}, callback);
                }, function(e){nvlib.handleError(e, callback)});
            }
        );

    }, function(e){nvlib.handleError(e, callback)});
};
