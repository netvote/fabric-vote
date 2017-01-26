
resource "aws_lambda_function" "voter_cast_votes" {
  filename = "lambdas-app.zip"
  function_name = "voter-cast-votes"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "voter-cast-votes.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: casts votes for a voter from API Gateway"
}

resource "aws_lambda_permission" "voter_cast_votes" {
  function_name = "${aws_lambda_function.voter_cast_votes.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.voter_cast_votes.http_method}${aws_api_gateway_resource.vote_ballot_id.path}"
}

resource "aws_lambda_function" "admin_get_ballot_results" {
  filename = "lambdas-app.zip"
  function_name = "admin-get-ballot-results"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-get-ballot-results.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: gets results for a ballot"
}

resource "aws_lambda_permission" "admin_get_ballot_results" {
  function_name = "${aws_lambda_function.admin_get_ballot_results.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_ballot_get_results.http_method}${aws_api_gateway_resource.admin_ballot_id_results.path}"
}

resource "aws_lambda_function" "admin_create_ballot" {
  filename = "lambdas-app.zip"
  function_name = "admin-create-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-create-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: creates a ballot on the blockchain"
}

resource "aws_lambda_permission" "admin_create_ballot" {
  function_name = "${aws_lambda_function.admin_create_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_create_ballot.http_method}${aws_api_gateway_resource.admin_ballot_list.path}"
}

resource "aws_lambda_function" "admin_get_ballots" {
  filename = "lambdas-app.zip"
  function_name = "admin-get-ballots"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-get-ballots.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: gets all ballots from DynamoDB created by this user"
}

resource "aws_lambda_permission" "admin_get_ballot_list" {
  function_name = "${aws_lambda_function.admin_get_ballots.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_get_ballots.http_method}${aws_api_gateway_resource.admin_ballot_list.path}"
}

resource "aws_lambda_function" "admin_update_ballot" {
  filename = "lambdas-app.zip"
  function_name = "admin-update-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-update-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: updates the ballot that was originally saved (from DynamoDB)"
}

resource "aws_lambda_permission" "admin_update_ballot" {
  function_name = "${aws_lambda_function.admin_update_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_update_ballot.http_method}${aws_api_gateway_resource.admin_ballot_id.path}"
}

resource "aws_lambda_function" "admin_get_ballot" {
  filename = "lambdas-app.zip"
  function_name = "admin-get-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-get-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: gets the ballot that was originally saved (from DynamoDB)"
}

resource "aws_lambda_permission" "admin_get_ballot" {
  function_name = "${aws_lambda_function.admin_get_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_get_ballot.http_method}${aws_api_gateway_resource.admin_ballot_id.path}"
}

resource "aws_lambda_function" "admin_share_ballot" {
  filename = "lambdas-app.zip"
  function_name = "admin-share-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-share-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: sends a link to sms or email"
}


resource "aws_lambda_permission" "admin_share_ballot" {
  function_name = "${aws_lambda_function.admin_share_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_share_ballot.http_method}${aws_api_gateway_resource.admin_ballot_id_share.path}"
}

resource "aws_lambda_function" "admin_delete_ballot" {
  filename = "lambdas-app.zip"
  function_name = "admin-delete-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "admin-delete-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: deletes a ballot, decisions, and results by ballot Id"
}

resource "aws_lambda_permission" "admin_delete_ballot" {
  function_name = "${aws_lambda_function.admin_delete_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.admin_delete_ballot.http_method}${aws_api_gateway_resource.admin_ballot_id.path}"
}

resource "aws_lambda_function" "voter_get_ballot" {
  filename = "lambdas-app.zip"
  function_name = "voter-get-ballot"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "voter-get-ballot.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: initializes and retrieves ballot for a voter"
}

resource "aws_lambda_permission" "voter_get_ballot" {
  function_name = "${aws_lambda_function.voter_get_ballot.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.voter_get_ballot.http_method}${aws_api_gateway_resource.vote_ballot_id.path}"
}

resource "aws_lambda_function" "voter_get_ballots" {
  filename = "lambdas-app.zip"
  function_name = "voter-get-ballots"
  role = "${aws_iam_role.netvote_api_lambda.arn}"
  handler = "voter-get-ballots.handler"
  runtime = "nodejs4.3"
  source_code_hash = "${base64sha256(file("lambdas-app.zip"))}"
  publish = true
  timeout = 10
  description = "APP: retrieves all ballots for a voter"
}

resource "aws_lambda_permission" "voter_get_ballots" {
  function_name = "${aws_lambda_function.voter_get_ballots.function_name}"
  statement_id = "AllowExecutionFromApiGateway"
  action = "lambda:InvokeFunction"
  principal = "apigateway.amazonaws.com"
  source_arn = "arn:aws:execute-api:${var.region}:${var.account}:${aws_api_gateway_rest_api.netvote_mobile_api.id}/*/${aws_api_gateway_method.voter_get_ballot.http_method}${aws_api_gateway_resource.vote_ballot.path}"
}