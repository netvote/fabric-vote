resource "aws_api_gateway_rest_api" "netvote_api" {
  name = "Netvote API"
  description = "This is a user-facing API"
}

resource "aws_api_gateway_resource" "create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  parent_id = "${aws_api_gateway_rest_api.netvote_api.root_resource_id}"
  path_part = "ballot"
}

resource "aws_api_gateway_method" "create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.create_ballot.id}"
  http_method = "POST"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "create_ballot" {
  rest_api_id = "${aws_api_gateway_rest_api.netvote_api.id}"
  resource_id = "${aws_api_gateway_resource.create_ballot.id}"
  http_method = "POST"
  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.create_ballot.arn}/invocations"
}