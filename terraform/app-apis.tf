//TODO: replace once terraform supports cognito pools
variable "authorizer_id"{
  default = "l5g35e"
}

resource "aws_api_gateway_rest_api" "netvote_mobile_api" {
  name = "Netvote Mobile API"
  description = "This is a mobile-facing API"
}

resource "aws_api_gateway_deployment" "mobile_dev" {
  depends_on = ["aws_api_gateway_method.admin_create_ballot","aws_api_gateway_integration.admin_create_ballot"]
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  stage_name = "netvote_mobile_dev"
}

resource "aws_api_gateway_resource" "vote" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_mobile_api.root_resource_id}"
  path_part = "vote"
}

resource "aws_api_gateway_resource" "vote_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.vote.id}"
  path_part = "ballot"
}

resource "aws_api_gateway_method" "voter_get_ballots" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.vote_ballot.id}"
  http_method = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "voter_get_ballots" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.vote_ballot.id}"
  http_method = "${aws_api_gateway_method.voter_get_ballots.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.voter_get_ballots.arn}/invocations"
}

resource "aws_api_gateway_resource" "vote_ballot_id" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.vote_ballot.id}"
  path_part = "{ballotId}"
}

resource "aws_api_gateway_method" "voter_get_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.vote_ballot_id.id}"
  http_method = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "voter_get_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.vote_ballot_id.id}"
  http_method = "${aws_api_gateway_method.voter_get_ballot.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.voter_get_ballot.arn}/invocations"
}

resource "aws_api_gateway_method" "voter_cast_votes" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.vote_ballot_id.id}"
  http_method = "POST"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "voter_cast_votes" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.vote_ballot_id.id}"
  http_method = "${aws_api_gateway_method.voter_cast_votes.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.voter_cast_votes.arn}/invocations"
}

resource "aws_api_gateway_resource" "admin" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_mobile_api.root_resource_id}"
  path_part = "admin"
}

resource "aws_api_gateway_resource" "admin_ballot_list" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.admin.id}"
  path_part = "ballot"
}

resource "aws_api_gateway_method" "admin_get_ballots" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_list.id}"
  http_method = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "admin_get_ballots" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_list.id}"
  http_method = "${aws_api_gateway_method.admin_get_ballots.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.admin_get_ballots.arn}/invocations"
}

resource "aws_api_gateway_method" "admin_create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_list.id}"
  http_method = "POST"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "admin_create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_list.id}"
  http_method = "${aws_api_gateway_method.admin_create_ballot.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.admin_create_ballot.arn}/invocations"
}

resource "aws_api_gateway_resource" "admin_ballot_id" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.admin_ballot_list.id}"
  path_part = "{ballotId}"
}

resource "aws_api_gateway_method" "admin_update_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  http_method = "PUT"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "admin_update_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  http_method = "${aws_api_gateway_method.admin_update_ballot.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.admin_update_ballot.arn}/invocations"
}

resource "aws_api_gateway_method" "admin_get_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  http_method = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "admin_get_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  http_method = "${aws_api_gateway_method.admin_get_ballot.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.admin_get_ballot.arn}/invocations"
}

resource "aws_api_gateway_method" "admin_delete_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  http_method = "DELETE"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "admin_delete_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  http_method = "${aws_api_gateway_method.admin_delete_ballot.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.admin_delete_ballot.arn}/invocations"
}

resource "aws_api_gateway_resource" "admin_ballot_id_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.admin_ballot_id.id}"
  path_part = "results"
}

resource "aws_api_gateway_method" "admin_ballot_get_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id_results.id}"
  http_method = "GET"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "admin_ballot_get_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot_id_results.id}"
  http_method = "${aws_api_gateway_method.admin_ballot_get_results.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.admin_get_ballot_results.arn}/invocations"
}

#/security
resource "aws_api_gateway_resource" "app_security" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_mobile_api.root_resource_id}"
  path_part = "security"
}

resource "aws_api_gateway_resource" "app_code" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.app_security.id}"
  path_part = "code"
}

resource "aws_api_gateway_resource" "app_smscode" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  parent_id = "${aws_api_gateway_resource.app_code.id}"
  path_part = "sms"
}

resource "aws_api_gateway_method" "app_smscode" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.app_smscode.id}"
  http_method = "POST"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${var.authorizer_id}"
  api_key_required = true
}

resource "aws_api_gateway_integration" "app_smscode" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_mobile_api.id}"
  resource_id = "${aws_api_gateway_resource.app_smscode.id}"
  http_method = "${aws_api_gateway_method.app_smscode.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.send_sms_code.arn}/invocations"
}
