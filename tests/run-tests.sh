ADMIN_KEY=tiDW8ev7zCFuQNqgG8Bi2Yen9X0nWry510RKqv60
VOTER_KEY=HqM0cBsM1Z7GOIileyUwv9vGqIeJuZB45q9hC2sc


IDX=`date +%s`
VOTER_ID="slanders$IDX"

sed -i.bak "s/IDX/$IDX/g" *.json

echo "creating ballot"
./create_ballot.sh $ADMIN_KEY
sleep 2

echo "getting results"
./get_results.sh $ADMIN_KEY $IDX
echo "done"
sleep 2

echo "getting ballot"
./get_ballot.sh $VOTER_KEY $VOTER_ID
echo "done"
sleep 2

echo "casting vote"
./cast_votes.sh $VOTER_KEY $VOTER_ID
echo "done"
sleep 2

echo "getting results"
./get_results.sh $ADMIN_KEY $IDX
echo "done"

mv ballot.json.bak ballot.json
mv votes.json.bak ballot.json