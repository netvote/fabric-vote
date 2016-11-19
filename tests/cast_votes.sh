curl -X POST -H "x-api-key: $1" -H "Content-Type: application/json" --data @votes.json https://9cylao0on7.execute-api.us-east-1.amazonaws.com/dev/vote/$2
