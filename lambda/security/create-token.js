process.env['PATH'] = process.env['PATH'] + "/" + process.env['LAMBDA_TASK_ROOT'];

var nJwt = require('njwt');
var uuid = require("uuid");
var doc = require('dynamodb-doc');
var dynamo = new doc.DynamoDB();

//TODO: encrypt and checkout decryption key from KMS
var signingKey = "c254rpMd4mc9dW0F5T2y4J9SnhBxr2Io";


//TODO: include Secure flag (once we have HTTPS)
var toCookieStr = function(cookieObj){
    var result = "";
    for(var key in cookieObj){
        var cookie = cookieObj[key];
        var expires = cookie.expires !== undefined ? "expires=" + cookie.expires : "";
        result += key+"="+cookie.value+"; HttpOnly; "+expires;
    }
    return result;
};

var generateToken = function(widgetId, voteId, accountId, callback){
    var claims = {
        widgetId: widgetId,  
        voteId: voteId,
        accountId: accountId
    };

    var jwt = nJwt.create(claims,signingKey);

    callback(jwt.compact());
};

var getDynamoItem = function(table, key, value, errorCallback, callback){
    var params = {
        TableName: table,
        Key:{
            key: value
        }
    };

    dynamo.getItem(params, function(err, data) {
        if (err) {
            console.error("Unable to read item. Error JSON:", JSON.stringify(err, null, 2));
            errorCallback(err);
        } else{
            callback(data);
        }
    });
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
    var widgetId = event.widgetId;
    var voterId = event.voterId;
    
    if(voterId === null){
        voterId = uuid.v4();
    }

    var now = new Date();
    var expiration = new Date(now.getTime()+(30*60*1000));


    getDynamoItem("widgets", "id", widgetId, function(err){
        handleError(err, callback);
    }, function(data){

        var accountId = data.item.accountId;

        generateToken(widgetId, voterId, accountId, function(tokenId){

            var voterIdCookieName = "voterId_"+accountId;
            var tokenIdCookieName = "token_"+widgetId;

            cookies = {};
            cookies[voterIdCookieName] = {
                value: voterId
            };
            cookies[tokenIdCookieName] = {
                value: tokenId,
                expires: expiration.toGMTString()
            };

            callback(null, {
                "statusCode": 200,
                "headers": {
                    "Cookies": toCookieStr(cookies)
                },
                "body": {}
            });
        });
    });


};
