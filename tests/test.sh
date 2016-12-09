HOST=https://0i2c83b9q9.execute-api.us-east-1.amazonaws.com/netvote_dev
ADMIN_KEY=LqmkaqEpkH63JdmpDi4Ty2ArLRPv65WQ5rM7tAWW
VOTER_KEY=C1lBXZzwA61rvI1eLZmYh24M4s2Yga675E8woTt7

IDX=`date +%s`
VOTER_ID="slanders$IDX"
BALLOT_ID="beercolor-$IDX"

echo $IDX
sed -i.bak "s/IDX/$IDX/g" *.json

echo "ADMIN: CREATING BALLOT"
curl -X POST -H "x-api-key: $ADMIN_KEY" -H "Content-Type: application/json" --data @ballot.json $HOST/ballot
echo ""
sleep 1

echo "ADMIN: GETTING RESULTS"
curl -s -H "x-api-key: $ADMIN_KEY" $HOST/results/decision/favorite-color$IDX |jq
curl -s -H "x-api-key: $ADMIN_KEY" $HOST/results/decision/favorite-beer$IDX |jq
echo ""
sleep 1

echo "getting ballot"
curl -s -H "x-api-key: $VOTER_KEY" $HOST/voter/$VOTER_ID/ballot |jq
echo ""
sleep 1

echo "VOTER: REQUEST SMS CODE"
curl -X POST -H "x-api-key: $VOTER_KEY" -H "Content-Type: application/json" --data @smsrequest.json $HOST/security/code/sms
echo ""
echo "ENTER CODE:"
read code

echo "VOTER: CASTING VOTE"
curl -X POST -H "x-api-key: $VOTER_KEY" -H "nv-two-factor-code: $code" -H "Content-Type: application/json" --data @votes.json $HOST/voter/$VOTER_ID/ballot/$BALLOT_ID
echo ""
sleep 1

echo "ADMIN: GETTING RESULTS"
curl -s -H "x-api-key: $ADMIN_KEY" $HOST/results/decision/favorite-color$IDX |jq
curl -s -H "x-api-key: $ADMIN_KEY" $HOST/results/decision/favorite-beer$IDX |jq
echo ""
sleep 1

echo "ADMIN: DELETING BALLOT RESULTS"
curl -s  -X DELETE -H "x-api-key: $ADMIN_KEY" $HOST/ballot/$BALLOT_ID |jq
echo ""
sleep 1

echo "ADMIN: GETTING DELETED RESULTS"
curl -s -H "x-api-key: $ADMIN_KEY" $HOST/results/decision/favorite-color$IDX |jq
curl -s -H "x-api-key: $ADMIN_KEY" $HOST/results/decision/favorite-beer$IDX |jq
echo ""
sleep 1

mv ballot.json.bak ballot.json
mv votes.json.bak votes.json
mv token.json.bak token.json
mv smsrequest.json.bak smsrequest.json
