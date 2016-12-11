'use strict';

console.log('Loading function');

exports.handler = (event, context, callback) => {
    //console.log('Received event:', JSON.stringify(event, null, 2));
    event.Records.forEach((record) => {
        // Kinesis data is base64 encoded so decode here
        const payload = new Buffer(record.kinesis.data, 'base64').toString('ascii');
        var voteEvent = JSON.parse(payload)
        console.log('Decoded payload:', JSON.stringify(voteEvent));
    });
    callback(null, `Successfully processed ${event.Records.length} records.`);
};
