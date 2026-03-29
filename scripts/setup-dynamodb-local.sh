#!/bin/bash

set -e

# Configuration for local DynamoDB
TABLE_NAME=${DYNAMODB_TABLE_NAME:-"iantraining"}
ENDPOINT=${DYNAMODB_ENDPOINT:-"http://localhost:8000"}

echo "Setting up local DynamoDB table: $TABLE_NAME at $ENDPOINT"

# Create the main table with GSIs for local DynamoDB
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
    --endpoint-url "$ENDPOINT" \
    --region us-east-1

echo "Waiting for table to become active..."
aws dynamodb wait table-exists --table-name "$TABLE_NAME" --endpoint-url "$ENDPOINT" --region us-east-1

echo "Table $TABLE_NAME created successfully!"

# Show table description
aws dynamodb describe-table --table-name "$TABLE_NAME" --endpoint-url "$ENDPOINT" --region us-east-1
