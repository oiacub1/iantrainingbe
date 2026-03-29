#!/bin/bash

set -e

# Configuration
TABLE_NAME=${DYNAMODB_TABLE_NAME:-"iantraining"}
REGION=${AWS_REGION:-"us-east-1"}

echo "Setting up DynamoDB table: $TABLE_NAME in region $REGION"

# Create the main table with GSIs
aws dynamodb create-table \
    --table-name "$TABLE_NAME" \
    --attribute-definitions \
        AttributeName=PK,AttributeType=S \
        AttributeName=SK,AttributeType=S \
        AttributeName=GSI1PK,AttributeType=S \
        AttributeName=GSI1SK,AttributeType=S \
    --key-schema \
        AttributeName=PK,KeyType=HASH \
        AttributeName=SK,KeyType=RANGE \
    --billing-mode PAY_PER_REQUEST \
    --global-secondary-indexes \
        '[{
            "IndexName": "GSI1",
            "KeySchema": [
                {"AttributeName":"GSI1PK","KeyType":"HASH"},
                {"AttributeName":"GSI1SK","KeyType":"RANGE"}
            ],
            "Projection":{"ProjectionType":"ALL"}
        }]' \
    --region "$REGION"

echo "Waiting for table to become active..."
aws dynamodb wait table-exists --table-name "$TABLE_NAME" --region "$REGION"

echo "Table $TABLE_NAME created successfully!"

# Show table description
aws dynamodb describe-table --table-name "$TABLE_NAME" --region "$REGION"
