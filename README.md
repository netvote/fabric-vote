# fabric-vote

This evolved from the [Hyperledger Starter Kit](https://hyperledger-fabric.readthedocs.io/en/latest/starter/fabric-starter-kit/#fabric-starter-kit).  

The project consists of three components:

### chaincode (golang):  

This contains blockchian transactions for creating decisions, voters, and casting votes.


### membersrvc

For now, this is a hardcoded yaml with the initial admin.  All users are created at run-time via the node apps.


### node clients

These are basic clients that interact with a docker peer and can `Invoke` or `Query` chaincode transactions.
 
The Hyperledger rest API is available on port 7050 on the peer node.