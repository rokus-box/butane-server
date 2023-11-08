resource "aws_apigatewayv2_integration" "oauth" {
  api_id             = aws_apigatewayv2_api.http.id
  integration_uri    = aws_lambda_function.lambdas["oauth"].invoke_arn
  integration_method = "POST"
  integration_type   = "AWS_PROXY"
  response_parameters {
    mappings = {
      "integration.response.header.Access-Control-Allow-Origin" = "'*'"
    }
    status_code = "200"
  }
}

resource "aws_apigatewayv2_integration" "vault" {
  api_id             = aws_apigatewayv2_api.http.id
  integration_uri    = aws_lambda_function.lambdas["vault"].invoke_arn
  integration_method = "POST"
  integration_type   = "AWS_PROXY"
}


resource "aws_apigatewayv2_integration" "session" {
  api_id             = aws_apigatewayv2_api.http.id
  integration_uri    = aws_lambda_function.lambdas["sessions"].invoke_arn
  integration_method = "POST"
  integration_type   = "AWS_PROXY"
}

#############################################
## ENDPOINTS
#############################################

resource "aws_apigatewayv2_route" "oauth" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "POST /oauth/{provider}/authenticate"
  target    = "integrations/${aws_apigatewayv2_integration.oauth.id}"
}

resource "aws_apigatewayv2_route" "get_vaults" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "GET /vault"
  target    = "integrations/${aws_apigatewayv2_integration.vault.id}"
}

resource "aws_apigatewayv2_route" "create_vault" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "POST /vault"
  target    = "integrations/${aws_apigatewayv2_integration.vault.id}"
}

resource "aws_apigatewayv2_route" "update_vault" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "PATCH /vault/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.vault.id}"
}

resource "aws_apigatewayv2_route" "delete_vault" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "DELETE /vault/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.vault.id}"
}

resource "aws_apigatewayv2_route" "get_sessions" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "GET /sessions"
  target    = "integrations/${aws_apigatewayv2_integration.session.id}"
}

resource "aws_apigatewayv2_route" "delete_session" {
  api_id    = aws_apigatewayv2_api.http.id
  route_key = "DELETE /sessions"
  target    = "integrations/${aws_apigatewayv2_integration.session.id}"
}
