resource "aws_iam_policy" "lambda_log" {
  name = "lambda-exec-policy"
  description = "Allows lambda functions to create log events"
  policy =  "${file("conf/lambda-log-policy.json")}"
}

resource "aws_iam_policy" "dynamo_full" {
  name = "dynamo-policy"
  description = "Allows access to DynamoDB data"
  policy =  "${file("conf/dynamo-policy.json")}"
}

resource "aws_iam_policy" "kinesis_read" {
  name = "kinesis-read"
  description = "Allows access to Kinesis stream"
  policy =  "${file("conf/lambda-kinesis-policy.json")}"
}


resource "aws_iam_policy" "apigateway_full" {
  name = "api-gateway-policy"
  description = "Allows access to modify api keys and usage plans in api gateway"
  policy =  "${file("conf/api-gateway-admin-policy.json")}"
}

# API Lambda Permissions

resource "aws_iam_role" "netvote_api_lambda" {
  name = "netvote-api-lambda"
  assume_role_policy = "${file("conf/lambda-assume-role-policy.json")}"
}

resource "aws_iam_role_policy_attachment" "lambda_exec_attach" {
  role = "${aws_iam_role.netvote_api_lambda.name}"
  policy_arn = "${aws_iam_policy.lambda_log.arn}"
}

resource "aws_iam_role_policy_attachment" "dynamo_attach" {
  role = "${aws_iam_role.netvote_api_lambda.name}"
  policy_arn = "${aws_iam_policy.dynamo_full.arn}"
}

resource "aws_iam_role_policy_attachment" "kinesis_attach" {
  role = "${aws_iam_role.netvote_api_lambda.name}"
  policy_arn = "${aws_iam_policy.kinesis_read.arn}"
}


# API KEY creation also needs access to API gateway modification

resource "aws_iam_role" "api_key_lambda" {
  name = "netvote-api-key-lambda"
  assume_role_policy = "${file("conf/lambda-assume-role-policy.json")}"
}

resource "aws_iam_role_policy_attachment" "lambda_exec_api_key_attach" {
  role = "${aws_iam_role.api_key_lambda.name}"
  policy_arn = "${aws_iam_policy.lambda_log.arn}"
}

resource "aws_iam_role_policy_attachment" "dynamo_api_key_attach" {
  role = "${aws_iam_role.api_key_lambda.name}"
  policy_arn = "${aws_iam_policy.dynamo_full.arn}"
}

resource "aws_iam_role_policy_attachment" "apigateway_api_key_attach" {
  role = "${aws_iam_role.api_key_lambda.name}"
  policy_arn = "${aws_iam_policy.apigateway_full.arn}"
}

//TODO add API KEY usage plan ids to enviornment parameters
resource "aws_lambda_function" "create_api_key" {
  filename = "lambdas.zip"
  function_name = "create-api-key"
  role = "${aws_iam_role.api_key_lambda.arn}"
  handler = "create-api-key.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "SYSTEM: Creates a chaincode account and API Key, stores in DynamoDB"
}


resource "aws_lambda_function" "cast_votes" {
  filename = "lambdas.zip"
  function_name = "cast-votes"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "cast-votes.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "VOTER: Casts votes for a voter from API Gateway"
}

resource "aws_lambda_permission" "cast_ballot_votes" {
  function_name = "${aws_lambda_function.cast_votes.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.cast_ballot_vote.http_method}${aws_api_gateway_resource.voter_ballot_by_id.path}"
}

resource "aws_lambda_function" "send_sms_code" {
  filename = "lambdas.zip"
  function_name = "send-sms-code"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "send-sms-code.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "VOTER: Sends an SMS code for Two-Factor Authentication"
}

resource "aws_lambda_permission" "send_sms_code" {
  function_name = "${aws_lambda_function.send_sms_code.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.smscode.http_method}${aws_api_gateway_resource.smscode.path}"
}

resource "aws_lambda_function" "get_ballot_results" {
  filename = "lambdas.zip"
  function_name = "get-ballot-results"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "get-ballot-results.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "ADMIN: gets results for a ballot"
}

resource "aws_lambda_permission" "get_ballot_results" {
  function_name = "${aws_lambda_function.get_ballot_results.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_ballot_results.http_method}${aws_api_gateway_resource.get_ballot_results_for_id.path}"
}

resource "aws_lambda_function" "create_ballot" {
  filename = "lambdas.zip"
  function_name = "create-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "create-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "ADMIN: creates a ballot on the blockchain"
}

resource "aws_lambda_permission" "create_ballot" {
  function_name = "${aws_lambda_function.create_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.create_ballot.http_method}${aws_api_gateway_resource.admin_ballot.path}"
}

resource "aws_lambda_function" "get_admin_ballot" {
  filename = "lambdas.zip"
  function_name = "get-admin-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "get-admin-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "ADMIN: gets the ballot that was originally saved (from DynamoDB)"
}

resource "aws_lambda_permission" "get_admin_ballot" {
  function_name = "${aws_lambda_function.get_admin_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_admin_ballot.http_method}${aws_api_gateway_resource.ballot_by_id.path}"
}

resource "aws_lambda_function" "delete_ballot" {
  filename = "lambdas.zip"
  function_name = "delete-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "delete-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "ADMIN: deletes a ballot, decisions, and results by ballot Id"
}

resource "aws_lambda_permission" "delete_ballot" {
  function_name = "${aws_lambda_function.delete_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.delete_ballot.http_method}${aws_api_gateway_resource.ballot_by_id.path}"
}

resource "aws_lambda_function" "get_voter_ballot" {
  filename = "lambdas.zip"
  function_name = "get-voter-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "get-voter-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "VOTER: initializes and retrieves ballot for a voter"
}

resource "aws_lambda_permission" "get_voter_ballot" {
  function_name = "${aws_lambda_function.get_voter_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_voter_ballot.http_method}${aws_api_gateway_resource.voter_ballot.path}"
}

resource "aws_lambda_permission" "get_voter_ballot_by_id" {
  function_name = "${aws_lambda_function.get_voter_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGatewayById"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_voter_ballot_by_id.http_method}${aws_api_gateway_resource.voter_ballot_by_id.path}"
}

resource "aws_lambda_function" "kinesis_logger" {
  filename = "lambdas.zip"
  function_name = "vote-kinesis-logger"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "kinesis-logger.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
  publish = true
  timeout = 10
  description = "TEST: logs the Vote kinesis stream entries to cloudwatch"
}

resource "aws_lambda_event_source_mapping" "event_source_mapping" {
  batch_size = 100
  event_source_arn = "${aws_kinesis_stream.votes.arn}"
  enabled = true
  function_name = "${aws_lambda_function.kinesis_logger.arn}"
  starting_position = "LATEST"
}
