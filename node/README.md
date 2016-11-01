# Node Clients

These are basic node clients that demonstrate interactions with the blockchain using hyperledgers node SDK.

The node modules must be installed while within a docker container.

1. Spin up docker in top dir:  `docker-compose up -d`
2. SSH to netvote container: `docker exec -it netvote /bin/bash`
3. Go into the vote directory: `cd /user/vote`
4. Install hyperledger node_modules `npm install /opt/gopath/src/github.com/hyperledger/fabric/sdk/node`

(The install must happen within linux...and my Mac doesn't have virtualization)