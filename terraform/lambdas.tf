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

//resource "aws_lambda_permission" "cast_votes" {
//  function_name = "${aws_lambda_function.cast_votes.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_integration.cast_votes.integration_http_method}${aws_api_gateway_resource.cast_votes.path}"
//}

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

//resource "aws_lambda_permission" "send_sms_code" {
//  function_name = "${aws_lambda_function.send_sms_code.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_integration.send_sms_code.integration_http_method}${aws_api_gateway_resource.send_sms_code.path}"
//}

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

//resource "aws_lambda_permission" "get_results" {
//  function_name = "${aws_lambda_function.get_results.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_integration.get_results.integration_http_method}${aws_api_gateway_resource.get_results.path}"
//}

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

resource "aws_lambda_permission" "create_ballot" {
  function_name = "${aws_lambda_function.create_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_integration.create_ballot.integration_http_method}${aws_api_gateway_resource.create_ballot.path}"
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

//resource "aws_lambda_permission" "get_ballot" {
//  function_name = "${aws_lambda_function.get_ballot.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_integration.get_ballot.integration_http_method}${aws_api_gateway_resource.get_ballot.path}"
//}
