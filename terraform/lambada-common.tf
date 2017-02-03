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


resource "aws_iam_policy" "ses_send_email" {
  name = "api-ses-policy"
  description = "Allows access to send ses email"
  policy =  "${file("conf/ses-lambda-policy.json")}"
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

resource "aws_iam_role_policy_attachment" "ses_attach" {
  role = "${aws_iam_role.netvote_api_lambda.name}"
  policy_arn = "${aws_iam_policy.ses_send_email.arn}"
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