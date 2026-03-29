resource "aws_dynamodb_table" "training_platform" {
  name           = "training-platform-${var.environment}"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "GSI1PK"
    type = "S"
  }

  attribute {
    name = "GSI1SK"
    type = "S"
  }

  attribute {
    name = "GSI2PK"
    type = "S"
  }

  attribute {
    name = "GSI2SK"
    type = "S"
  }

  global_secondary_index {
    name            = "GSI1"
    hash_key        = "GSI1PK"
    range_key       = "GSI1SK"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "GSI2"
    hash_key        = "GSI2PK"
    range_key       = "GSI2SK"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  tags = {
    Name        = "training-platform-${var.environment}"
    Environment = var.environment
    Project     = "training-platform"
    ManagedBy   = "terraform"
  }

  lifecycle {
    prevent_destroy = true
  }
}

output "dynamodb_table_name" {
  value       = aws_dynamodb_table.training_platform.name
  description = "Name of the DynamoDB table"
}

output "dynamodb_table_arn" {
  value       = aws_dynamodb_table.training_platform.arn
  description = "ARN of the DynamoDB table"
}

output "dynamodb_stream_arn" {
  value       = aws_dynamodb_table.training_platform.stream_arn
  description = "ARN of the DynamoDB stream"
}
