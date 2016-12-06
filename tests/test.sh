ADMIN_KEY=qyiJBqfcHQaxj7oyTBo9C6vmVBNRGEO172qgDG95
VOTER_KEY=eIxNknhkOr3bG5uk4HimK5P6cq6A050haN3pi9yu

IDX=`date +%s`
VOTER_ID="slanders$IDX"

echo $IDX

sed -i.bak "s/IDX/$IDX/g" *.json

echo "ADMIN: CREATING BALLOT"
./create_ballot.sh $ADMIN_KEY
echo ""
sleep 2

echo "ADMIN: GETTING RESULTS"
./get_results.sh $ADMIN_KEY $IDX
echo ""
sleep 2

echo "getting ballot"
./get_ballot.sh $VOTER_KEY $VOTER_ID
echo ""
sleep 2

echo "VOTER: REQUEST SMS CODE"
./request_sms_code.sh $VOTER_KEY
echo "ENTER CODE:"
read code

echo "VOTER: CASTING VOTE"
./cast_votes.sh $VOTER_KEY $VOTER_ID $code
sleep 2

echo "ADMIN: GETTING RESULTS"
./get_results.sh $ADMIN_KEY $IDX
sleep 2

mv ballot.json.bak ballot.json
mv votes.json.bak votes.json
mv token.json.bak token.json
mv smsrequest.json.bak smsrequest.json