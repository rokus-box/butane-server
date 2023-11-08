resource "aws_apigatewayv2_api" "http" {
  name          = "Delete Later"
  description   = "Example API. Delete this after you're done with it."
  protocol_type = "HTTP"
  cors_configuration {
    allow_headers = ["X-Mfa-Challenge", "Authorization"]
    allow_methods = ["OPTIONS", "GET", "POST", "PATCH", "DELETE"]
    allow_origins = ["http://localhost:4200"]
  }
}

resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.http.id
  auto_deploy = true
  name        = "prod"
}

#
#resource "aws_apigatewayv2_integration" "lambdas" {
#  for_each           = var.functions
#  api_id             = aws_apigatewayv2_api.http.id
#  integration_uri    = aws_lambda_function.lambdas[each.key].invoke_arn
#  integration_method = "POST"
#  integration_type   = "AWS_PROXY"
#}
#
#resource "aws_apigatewayv2_route" "lambdas" {
#  for_each  = toset(var.endpoints)
#  api_id    = aws_apigatewayv2_api.http.id
#  route_key = split(";", each.value)[1]
#  target    = "integrations/${aws_apigatewayv2_integration.lambdas[split(";", each.value)[0]].id}"
#}
#

resource "aws_lambda_permission" "api_gw" {
  for_each      = var.functions
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  function_name = aws_lambda_function.lambdas[each.key].function_name
  source_arn    = "${aws_apigatewayv2_api.http.execution_arn}/*/*"
}

output "api_url" {
  value = aws_apigatewayv2_stage.prod.invoke_url
}
