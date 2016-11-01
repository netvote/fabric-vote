# fabric-vote

This evolved from the [Hyperledger Starter Kit](https://hyperledger-fabric.readthedocs.io/en/latest/starter/fabric-starter-kit/#fabric-starter-kit).  

The project consists of three components:

### chaincode (golang):  

This contains blockchian transactions for creating decisions, voters, and casting votes. 

#### Invoke Transactions
- `add_decision`: (admin) create a decision configuration
- `add_voter`: (admin) creates a voter on blockchain and allocates votes
- `cast_votes`: (voter) spends votes on decisions, updates results, removes voter

#### Query Transactions
- `get_results`: (admin) retrieves current results of a given decision
- `get_ballot`: (voter) retrieves ballot and vote units for the current user (using certificate attribute)

### membersrvc

For now, this is a hardcoded yaml with the initial admin.  All users are created at runtime via the node apps.  

### node clients

These are basic clients that interact with a docker peer and can `Invoke` or `Query` chaincode transactions.
 
The Hyperledger rest API is available on port `7050` on the peer node.
