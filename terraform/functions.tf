resource "aws_iam_role" "fns" {
  name               = "ButaneServerLambdaRole"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "fns" {
  name   = "ButaneServerLambdaPolicy"
  role   = aws_iam_role.fns.id
  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Sid      = "LogsAccess"
        Action   = "logs:*"
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Sid      = "DynamoDBAccess"
        Action   = "dynamodb:*"
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}

#variable "env_vars" {
#  type = string
#}

resource "aws_lambda_function" "lambdas" {
  for_each         = var.functions
  # Zipping the function code using `zip` executable in Makefile
  filename         = "outputs/${each.key}.zip"
  timeout          = 5
  function_name    = each.key
  handler          = each.value.handler
  runtime          = each.value.runtime
  description      = each.value.description
  role             = aws_iam_role.fns.arn
  source_code_hash = filebase64sha256("outputs/${each.key}.zip")
  environment {
    variables = jsondecode(file("../env.json"))
  }
}

resource "aws_cloudwatch_log_group" "lambdas" {
  for_each          = var.functions
  name              = "/aws/lambda/${each.key}"
  retention_in_days = 3
}
