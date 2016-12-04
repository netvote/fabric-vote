ADMIN_KEY=eAwaldwcMe6KGLni59m7e3Z36ZGJQKoM1wjAMbfg
VOTER_KEY=5tesDCtQa07TkVeP4UQtz3FcKQPEzCab5ME9Ho95

IDX=`date +%s`
VOTER_ID="slanders$IDX"

echo $IDX

sed -i.bak "s/IDX/$IDX/g" *.json

echo "ADMIN: CREATING BALLOT"
./create_ballot.sh $ADMIN_KEY
sleep 2

echo "ADMIN: GETTING RESULTS"
./get_results.sh $ADMIN_KEY $IDX
sleep 2


TOKEN=`curl -s -X POST -H "x-api-key: $VOTER_KEY" -H "Content-Type: application/json" --data @token.json https://9cylao0on7.execute-api.us-east-1.amazonaws.com/dev/token | jq -r '.token'`

echo $TOKEN
echo "GETTING BALLOT"
curl -s -H "Authorization: $TOKEN" -H "x-api-key: $VOTER_KEY" https://9cylao0on7.execute-api.us-east-1.amazonaws.com/dev/ballot/$VOTER_ID |jq

sleep 2
echo "CASTING VOTE"
curl -X POST -H "Authorization: $TOKEN" -H "x-api-key: $VOTER_KEY" -H "Content-Type: application/json" --data @votes.json https://9cylao0on7.execute-api.us-east-1.amazonaws.com/dev/vote/$VOTER_ID

sleep 2
echo "GETTING BALLOT (EMPTY)"
curl -s -H "Authorization: $TOKEN" -H "x-api-key: $VOTER_KEY" https://9cylao0on7.execute-api.us-east-1.amazonaws.com/dev/ballot/$VOTER_ID |jq

sleep 2
echo "ADMIN: GETTING RESULTS"
./get_results.sh $ADMIN_KEY $IDX

mv ballot.json.bak ballot.json
mv votes.json.bak votes.json
mv token.json.bak token.json