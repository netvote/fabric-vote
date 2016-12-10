resource "aws_api_gateway_rest_api" "netvote_api" {
  name = "Netvote API"
  description = "This is a user-facing API"
}

resource "aws_api_gateway_deployment" "netvote_dev" {
  depends_on = ["aws_api_gateway_method.create_ballot","aws_api_gateway_integration.create_ballot"]
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  stage_name = "netvote_dev"
}

#######
#
#   ADMIN BALLOT CRUD
#
####

### CREATE

resource "aws_api_gateway_resource" "admin_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
  path_part = "ballot"
}

resource "aws_api_gateway_method" "create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot.id}"
  http_method = "POST"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.admin_ballot.id}"
  http_method = "POST"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.create_ballot.arn}/invocations"
}

### {ballotId}

resource "aws_api_gateway_resource" "ballot_by_id" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.admin_ballot.id}"
  path_part = "{ballotId}"
}


### GET BALLOT

resource "aws_api_gateway_method" "get_admin_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
  http_method = "GET"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "get_admin_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
  http_method = "GET"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_admin_ballot.arn}/invocations"
}



### DELETE BALLOT

resource "aws_api_gateway_method" "delete_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
  http_method = "DELETE"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "delete_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.ballot_by_id.id}"
  http_method = "DELETE"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.delete_ballot.arn}/invocations"
}



#######
#
#   SEND SMS 2FA
#
####

#/security
resource "aws_api_gateway_resource" "security" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
  path_part = "security"
}

resource "aws_api_gateway_resource" "code" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.security.id}"
  path_part = "code"
}

resource "aws_api_gateway_resource" "smscode" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.code.id}"
  path_part = "sms"
}

resource "aws_api_gateway_method" "smscode" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.smscode.id}"
  http_method = "POST"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "smscode" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.smscode.id}"
  http_method = "POST"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.send_sms_code.arn}/invocations"
}


#######
#
#   VOTER BALLOT
#
####

#/voter/
resource "aws_api_gateway_resource" "voter" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
  path_part = "voter"
}

#/voter/{voterid}
resource "aws_api_gateway_resource" "voterid" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.voter.id}"
  path_part = "{voterId}"
}

#/voter/{voterid}/ballot
resource "aws_api_gateway_resource" "voter_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.voterid.id}"
  path_part = "ballot"
}

#GET /voter/{voterid}/ballot
resource "aws_api_gateway_method" "get_voter_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.voter_ballot.id}"
  http_method = "GET"
  authorization = "NONE"
  api_key_required = true
}

#LAMBDA: get voter ballot
resource "aws_api_gateway_integration" "get_voter_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.voter_ballot.id}"
  http_method = "${aws_api_gateway_method.get_voter_ballot.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_voter_ballot.arn}/invocations"
}

#######
#
#   VOTER CAST VOTE
#
####

#/voter/{voterid}/ballot/{ballotid}
resource "aws_api_gateway_resource" "voter_ballot_by_id" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.voter_ballot.id}"
  path_part = "{ballotId}"
}

#GET /voter/{voterid}/ballot
resource "aws_api_gateway_method" "cast_ballot_vote" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.voter_ballot_by_id.id}"
  http_method = "POST"
  authorization = "NONE"
  api_key_required = true
}

#LAMBDA: get voter ballot
resource "aws_api_gateway_integration" "cast_ballot_vote" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.voter_ballot_by_id.id}"
  http_method = "${aws_api_gateway_method.cast_ballot_vote.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.cast_votes.arn}/invocations"
}



#######
#
#   RESULTS
#
####

#/results/
resource "aws_api_gateway_resource" "get_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
  path_part = "results"
}

#/results/decision/
resource "aws_api_gateway_resource" "get_decision_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.get_results.id}"
  path_part = "decision"
}

#GET /results/decision/{decisionid}
resource "aws_api_gateway_resource" "get_decision_results_for_id" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_resource.get_decision_results.id}"
  path_part = "{decisionId}"
}


resource "aws_api_gateway_method" "get_decision_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.get_decision_results_for_id.id}"
  http_method = "GET"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "get_decision_results" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.get_decision_results_for_id.id}"
  http_method = "${aws_api_gateway_method.get_decision_results.http_method}"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.get_results.arn}/invocations"
}