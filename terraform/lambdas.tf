//resource "aws_lambda_permission" "send_sms_code" {
//  function_name = "${aws_lambda_function.send_sms_code.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.smscode.http_method}${aws_api_gateway_resource.smscode.path}"
//}
//
//resource "aws_lambda_permission" "app_send_sms_code" {
//  function_name = "${aws_lambda_function.send_sms_code.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.smscode.http_method}${aws_api_gateway_resource.app_smscode.path}"
//}
//
//resource "aws_lambda_function" "cast_votes" {
//  filename = "lambdas.zip"
//  function_name = "cast-votes"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "cast-votes.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "VOTER: Casts votes for a voter from API Gateway"
//}
//
//resource "aws_lambda_permission" "cast_ballot_votes" {
//  function_name = "${aws_lambda_function.cast_votes.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.cast_ballot_vote.http_method}${aws_api_gateway_resource.voter_ballot_by_id.path}"
//}
//
//resource "aws_lambda_function" "get_ballot_results" {
//  filename = "lambdas.zip"
//  function_name = "get-ballot-results"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "get-ballot-results.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "ADMIN: gets results for a ballot"
//}
//
//resource "aws_lambda_permission" "get_ballot_results" {
//  function_name = "${aws_lambda_function.get_ballot_results.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_ballot_results.http_method}${aws_api_gateway_resource.get_ballot_results_for_id.path}"
//}
//
//resource "aws_lambda_function" "create_ballot" {
//  filename = "lambdas.zip"
//  function_name = "create-ballot"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "create-ballot.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "ADMIN: creates a ballot on the blockchain"
//}
//
//resource "aws_lambda_permission" "create_ballot" {
//  function_name = "${aws_lambda_function.create_ballot.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.create_ballot.http_method}${aws_api_gateway_resource.admin_ballot.path}"
//}
//
//resource "aws_lambda_function" "get_account_ballots" {
//  filename = "lambdas.zip"
//  function_name = "get-account-ballots"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "get-account-ballots.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "ADMIN: gets all ballots from DynamoDB for current account"
//}
//
//resource "aws_lambda_permission" "get_account_ballots" {
//  function_name = "${aws_lambda_function.get_account_ballots.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_account_ballots.http_method}${aws_api_gateway_resource.admin_ballot.path}"
//}
//
//resource "aws_lambda_function" "get_admin_ballot" {
//  filename = "lambdas.zip"
//  function_name = "get-admin-ballot"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "get-admin-ballot.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "ADMIN: gets the ballot that was originally saved (from DynamoDB)"
//}
//
//resource "aws_lambda_permission" "get_admin_ballot" {
//  function_name = "${aws_lambda_function.get_admin_ballot.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_admin_ballot.http_method}${aws_api_gateway_resource.ballot_by_id.path}"
//}
//
//resource "aws_lambda_function" "delete_ballot" {
//  filename = "lambdas.zip"
//  function_name = "delete-ballot"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "delete-ballot.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "ADMIN: deletes a ballot, decisions, and results by ballot Id"
//}
//
//resource "aws_lambda_permission" "delete_ballot" {
//  function_name = "${aws_lambda_function.delete_ballot.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.delete_ballot.http_method}${aws_api_gateway_resource.ballot_by_id.path}"
//}
//
//resource "aws_lambda_function" "get_voter_ballot" {
//  filename = "lambdas.zip"
//  function_name = "get-voter-ballot"
//  role = "${aws_iam_role.netvote_api_lambda.arn}"
//  handler = "get-voter-ballot.handler"
//  runtime = "nodejs4.3"
//  source_code_hash = "${base64sha256(file("lambdas.zip"))}"
//  publish = true
//  timeout = 10
//  description = "VOTER: initializes and retrieves ballot for a voter"
//}
//
//resource "aws_lambda_permission" "get_voter_ballot" {
//  function_name = "${aws_lambda_function.get_voter_ballot.function_name}"
//  statement_id = "AllowExecutionFromApiGateway"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_voter_ballot.http_method}${aws_api_gateway_resource.voter_ballot.path}"
//}
//
//resource "aws_lambda_permission" "get_voter_ballot_by_id" {
//  function_name = "${aws_lambda_function.get_voter_ballot.function_name}"
//  statement_id = "AllowExecutionFromApiGatewayById"
//  action = "lambda:InvokeFunction"
//  principal = "apigateway.amazonaws.com"
//  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_api.id}/*/${aws_api_gateway_method.get_voter_ballot_by_id.http_method}${aws_api_gateway_resource.voter_ballot_by_id.path}"
//}
//
//resource "aws_api_gateway_rest_api" "netvote_api" {
//  name = "Netvote API"
//  description = "This is a user-facing API"
//}
//
//resource "aws_api_gateway_deployment" "netvote_dev" {
//  depends_on = ["aws_api_gateway_method.create_ballot","aws_api_gateway_integration.create_ballot"]
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  stage_name = "netvote_dev"
//}
//
//#######
//#
//#   ADMIN BALLOT CRUD
//#
//####
//
//### CREATE
//
//resource "aws_api_gateway_resource" "admin_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
//  path_part = "ballot"
//}
//
//resource "aws_api_gateway_method" "create_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.admin_ballot.id}"
//  http_method = "POST"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "create_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.admin_ballot.id}"
//  http_method = "${aws_api_gateway_method.create_ballot.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.create_ballot.arn}/invocations"
//}
//
//resource "aws_api_gateway_method" "get_account_ballots" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.admin_ballot.id}"
//  http_method = "GET"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "get_account_ballots" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.admin_ballot.id}"
//  http_method = "${aws_api_gateway_method.get_account_ballots.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_account_ballots.arn}/invocations"
//}
//
//### {ballotId}
//
//resource "aws_api_gateway_resource" "ballot_by_id" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.admin_ballot.id}"
//  path_part = "{ballotId}"
//}
//
//
//### GET BALLOT
//
//resource "aws_api_gateway_method" "get_admin_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
//  http_method = "GET"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "get_admin_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
//  http_method = "${aws_api_gateway_method.get_admin_ballot.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_admin_ballot.arn}/invocations"
//}
//
//
//
//### DELETE BALLOT
//
//resource "aws_api_gateway_method" "delete_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
//  http_method = "DELETE"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "delete_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
//  http_method = "${aws_api_gateway_method.delete_ballot.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.delete_ballot.arn}/invocations"
//}
//
//
//
//#######
//#
//#   SEND SMS 2FA
//#
//####
//
//#/security
//resource "aws_api_gateway_resource" "security" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
//  path_part = "security"
//}
//
//resource "aws_api_gateway_resource" "code" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.security.id}"
//  path_part = "code"
//}
//
//resource "aws_api_gateway_resource" "smscode" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.code.id}"
//  path_part = "sms"
//}
//
//resource "aws_api_gateway_method" "smscode" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.smscode.id}"
//  http_method = "POST"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "smscode" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.smscode.id}"
//  http_method = "${aws_api_gateway_method.smscode.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.send_sms_code.arn}/invocations"
//}
//
//
//#######
//#
//#   VOTER BALLOT
//#
//####
//
//#/voter/
//resource "aws_api_gateway_resource" "voter" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
//  path_part = "voter"
//}
//
//#/voter/{voterid}
//resource "aws_api_gateway_resource" "voterid" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.voter.id}"
//  path_part = "{voterId}"
//}
//
//#/voter/{voterid}/ballot
//resource "aws_api_gateway_resource" "voter_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.voterid.id}"
//  path_part = "ballot"
//}
//
//#GET /voter/{voterid}/ballot
//resource "aws_api_gateway_method" "get_voter_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.voter_ballot.id}"
//  http_method = "GET"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//#LAMBDA: get voter ballot
//resource "aws_api_gateway_integration" "get_voter_ballot" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.voter_ballot.id}"
//  http_method = "${aws_api_gateway_method.get_voter_ballot.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_voter_ballot.arn}/invocations"
//}
//
//#/voter/{voterid}/ballot/{ballotid}
//resource "aws_api_gateway_resource" "voter_ballot_by_id" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.voter_ballot.id}"
//  path_part = "{ballotId}"
//}
//
//#GET /voter/{voterid}/ballot/{ballotid}
//resource "aws_api_gateway_method" "get_voter_ballot_by_id" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.voter_ballot_by_id.id}"
//  http_method = "GET"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "get_voter_ballot_by_id" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.voter_ballot_by_id.id}"
//  http_method = "${aws_api_gateway_method.get_voter_ballot_by_id.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_voter_ballot.arn}/invocations"
//}
//
//#######
//#
//#   VOTER CAST VOTE
//#
//####
//
//#POST /voter/{voterid}/ballot/{ballotid}
//resource "aws_api_gateway_method" "cast_ballot_vote" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.voter_ballot_by_id.id}"
//  http_method = "POST"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//#LAMBDA: get voter ballot
//resource "aws_api_gateway_integration" "cast_ballot_vote" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.voter_ballot_by_id.id}"
//  http_method = "${aws_api_gateway_method.cast_ballot_vote.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.cast_votes.arn}/invocations"
//}
//
//
//
//#######
//#
//#   RESULTS
//#
//####
//
//#/results/
//resource "aws_api_gateway_resource" "get_results" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
//  path_part = "results"
//}
//
//#/results/ballot/
//resource "aws_api_gateway_resource" "get_ballot_results" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.get_results.id}"
//  path_part = "ballot"
//}
//
//#/results/ballot/{ballotId}
//resource "aws_api_gateway_resource" "get_ballot_results_for_id" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  parent_id = "${aws_api_gateway_resource.get_ballot_results.id}"
//  path_part = "{ballotId}"
//}
//
//resource "aws_api_gateway_method" "get_ballot_results" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.get_ballot_results_for_id.id}"
//  http_method = "GET"
//  authorization = "NONE"
//  api_key_required = true
//}
//
//resource "aws_api_gateway_integration" "get_ballot_results" {
//  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
//  resource_id = "${aws_api_gateway_resource.get_ballot_results_for_id.id}"
//  http_method = "${aws_api_gateway_method.get_ballot_results.http_method}"
//  integration_http_method = "POST"
//  type = "AWS_PROXY"
//  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_ballot_results.arn}/invocations"
//}