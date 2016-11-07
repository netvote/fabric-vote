
var username, role;
if(process.argv.length < 3){
    console.error("Parameter required: username")
    process.exit(1);
}

username = process.argv[2];
role = process.argv[3];

var hfc = require("hfc");

console.log(" **** starting NetVote ****");

// get the addresses from the docker-compose environment
var MEMBERSRVC_ADDRESS   = process.env.MEMBERSRVC_ADDRESS;
var chain;
chain = hfc.newChain("voter-client");
chain.setKeyValStore( hfc.newFileKeyValStore('/tmp/keyValStore') );
console.log("member services address ="+MEMBERSRVC_ADDRESS);
chain.setMemberServicesUrl("grpc://"+MEMBERSRVC_ADDRESS);


chain.enroll("admin", "Xurw3yU9zI0l", function(err, admin) {
    if (err) {
        console.log("ERROR: failed to register admin: %s",err);
        process.exit(1);
    }
    // Set this user as the chain's registrar which is authorized to register other users.
    chain.setRegistrar(admin);

    var userName = username;
    // registrationRequest
    var registrationRequest = {
        enrollmentID: userName,
        affiliation: "company_a",
        attributes: [
	    { name: "account_id", value: "acct-id" },
            { name: "voter_id", value: userName },
            { name: "role", value: role}
        ]
    };

    chain.register(registrationRequest, function(error, user) {
        if (error) throw Error(" Failed to register and enroll " + userName + ": " + error);
        console.log("Registered %s successfully\n", userName);
        console.log(user.toString()+"\n");

        //TODO: somehow this code needs to get to the user

    });

});
