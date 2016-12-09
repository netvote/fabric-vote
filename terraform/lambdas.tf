
resource "aws_lambda_function" "cast_votes" {
  filename = "lambdas.zip"
  function_name = "cast-votes"
  role = "arn:aws:iam::845215180986:role/service-role/lambda-get-ballot"
  handler = "cast-votes.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  description = "VOTER: Casts votes for a voter from API Gateway"
}

resource "aws_lambda_function" "send_sms_code" {
  filename = "lambdas.zip"
  function_name = "send-sms-code"
  role = "arn:aws:iam::845215180986:role/service-role/lambda-get-ballot"
  handler = "send-sms-code.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  description = "VOTER: Sends an SMS code for Two-Factor Authentication"
}

resource "aws_lambda_function" "create_api_key" {
  filename = "lambdas.zip"
  function_name = "create-api-key"
  role = "arn:aws:iam::845215180986:role/service-role/lambda-get-ballot"
  handler = "send-sms-code.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  description = "SYSTEM: Creates a chaincode account and API Key, stores in DynamoDB"
}

resource "aws_lambda_function" "get_results" {
  filename = "lambdas.zip"
  function_name = "get-results"
  role = "arn:aws:iam::845215180986:role/service-role/lambda-get-ballot"
  handler = "get-results.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  description = "ADMIN: gets results for a particular decision"
}

resource "aws_lambda_function" "create_ballot" {
  filename = "lambdas.zip"
  function_name = "create-ballot"
  role = "arn:aws:iam::845215180986:role/service-role/lambda-get-ballot"
  handler = "get-results.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  description = "ADMIN: creates a ballot on the blockchain"
}

resource "aws_lambda_function" "get_ballot" {
  filename = "lambdas.zip"
  function_name = "get-ballot"
  role = "arn:aws:iam::845215180986:role/service-role/lambda-get-ballot"
  handler = "get-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  description = "VOTER: initializes and retrieves ballot for a voter"
}