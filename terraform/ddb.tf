# Table
resource "aws_dynamodb_table" "butane_table" {
  name         = "ButaneTable"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  ttl {
    attribute_name = "delete_after"
    enabled        = true
  }
}

locals {
  fn_name = "session_log"
}

resource "aws_lambda_function" "session_log" {
  # Zipping the function code using `zip` executable in Makefile
  filename         = "outputs/${local.fn_name}.zip"
  timeout          = 5
  function_name    = local.fn_name
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  description      = "Handles DynamoDB stream events and logs write actions to DynamoDB"
  role             = aws_iam_role.fns.arn
  source_code_hash = filebase64sha256("outputs/${local.fn_name}.zip")
  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.butane_table.name
    }
  }
}

resource "aws_iam_role_policy" "session_log" {
  name   = "${local.fn_name}_policy"
  role   = aws_iam_role.fns.id
  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Sid      = ""
        Action   = "logs:*"
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Sid    = "APIAccessForDynamoDBStreams",
        Effect = "Allow",
        Action = [
          "dynamodb:GetRecords",
          "dynamodb:GetShardIterator",
          "dynamodb:DescribeStream",
          "dynamodb:ListStreams"
        ],
        Resource = aws_dynamodb_table.butane_table.stream_arn
      }
    ]
  })
}

resource "aws_cloudwatch_log_group" "session_log" {
  name              = "/aws/lambda/${local.fn_name}"
  retention_in_days = 3
}

resource "aws_lambda_event_source_mapping" "butane_table_stream" {
  depends_on        = [aws_lambda_function.session_log]
  batch_size        = 1
  starting_position = "LATEST"
  function_name     = local.fn_name
  event_source_arn  = aws_dynamodb_table.butane_table.stream_arn
  filter_criteria {
    filter {
      pattern = jsonencode({
        eventName = ["INSERT", "REMOVE"]
        dynamodb  = {
          Keys = {
            PK = {
              S = [{ prefix = "SS#" }]
            }
          }
        }
      })
    }
  }
}

resource "aws_lambda_permission" "butane_table_stream" {
  statement_id  = "AllowExecutionFromDynamoDB"
  action        = "lambda:InvokeFunction"
  function_name = local.fn_name
  principal     = "dynamodb.amazonaws.com"
  source_arn    = aws_dynamodb_table.butane_table.stream_arn
}
