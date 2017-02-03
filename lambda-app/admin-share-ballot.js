'use strict';

console.log('Loading function');
var nvlib = require("netvotelib");
var bwipjs = require('bwip-js');
var AWS = require('aws-sdk');
var ses = new AWS.SES({apiVersion: '2010-12-01'});



exports.handler = function(event, context, callback){
    console.log('Received event:', JSON.stringify(event, null, 2));
    console.log('Received context:', JSON.stringify(context, null, 2));

    var ballotId = event.pathParameters.ballotId;

    nvlib.nvInit(event, context, function(account) {

        var shareOptions = JSON.parse(event.body);

        var phones = shareOptions.sms ? shareOptions.sms : [];
        var emails = shareOptions.email ? shareOptions.email : [];

        var ballotUrl = "netvote://ballot/"+ballotId;

        nvlib.getDynamoItem("ballots", "id", account.account_id+":"+ballotId,
            function(e){
                nvlib.handleError(e, callback)
            },
            function(data){
                if(data == undefined || data.Item == undefined) {
                    nvlib.handleNotFound(callback);
                }else if(data.Item.owner != account.user){
                    nvlib.handleUnauthorized(callback);
                }else {
                    var ballot = JSON.parse(new Buffer(data.Item.payload, 'base64').toString("ascii"));

                    nvlib.sendSms(phones, "Ballot: "+ballotUrl, function (result) {

                        if(emails.length > 0) {
                            bwipjs.toBuffer({
                                bcid: 'qrcode',       // Barcode type
                                text: ballotUrl,    // Text to encode
                                includetext: false,            // Show human-readable text
                            }, function (err, png) {
                                if (err) {
                                    // Decide how to handle the error
                                    // `err` may be a string or Error object
                                } else {


var raw_email = new Buffer(`From: "" <steven.landers@gmail.com>
To: "" <EMAIL>
Subject: Netvote
MIME-Version: 1.0
Content-Type: multipart/mixed;
    boundary="_003_97DCB304C5294779BEBCFC8357FCCAAA"

--_003_97DCB304C5294779BEBCFC8357FCCAAA
Content-Type: text/plain; charset="us-ascii"
Content-Transfer-Encoding: quoted-printable

Please scan attached QR code.

--_003_97DCB304C5294779BEBCFC8357FCCAAA
Content-Type: image/png; name="netvote.png"
Content-Description: netvote.png
Content-Disposition: attachment; filename="netvote.png"; size=IMG_SIZE;
    creation-date="DATE";
    modification-date="DATE"
Content-Transfer-Encoding: base64

IMAGE_BASE64

--_003_97DCB304C5294779BEBCFC8357FCCAAA
    
`.replace('IMAGE_BASE64', png.toString('base64')).replace('IMG_SIZE',""+png.length).replace('EMAIL', emails[0]).replace(/DATE/g, ""+new Date()));

console.log(raw_email);


                                    var params = {
                                        RawMessage: {
                                            Data: raw_email
                                        },
                                        Source: "steven.landers@gmail.com"
                                    };

                                    ses.sendRawEmail(params, function(err, data) {
                                        if (err) {
                                            console.log(err, err.stack);
                                            nvlib.handleError(err, callback)
                                        } else {
                                            console.log(data);
                                            nvlib.handleSuccess(result, callback);
                                        }
                                    });
                                }
                            });
                        }else{
                            nvlib.handleSuccess(result, callback);
                        }

                    }, function (err) {
                        nvlib.handleError(err, callback)
                    });
                }
            }
        );

    }, function(err){
        nvlib.handleError(err, callback)
    });
};