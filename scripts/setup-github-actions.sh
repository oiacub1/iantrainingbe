#!/bin/bash

set -e

echo "GitHub Actions Setup Script for IanTrainingBackend"
echo "=================================================="
echo ""
echo "This script helps you configure the required secrets for GitHub Actions."
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed."
    echo "Please install it first: https://cli.github.com/"
    echo ""
    echo "For macOS:"
    echo "  brew install gh"
    echo ""
    echo "For Ubuntu/Debian:"
    echo "  sudo apt install gh"
    echo ""
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo "Please authenticate with GitHub CLI first:"
    echo "  gh auth login"
    echo ""
    exit 1
fi

echo "Step 1: Get AWS Access Keys"
echo "----------------------------"
echo "You need to create an IAM User in AWS with the following permissions:"
echo ""
echo "1. Lambda permissions:"
echo "   - lambda:UpdateFunctionCode"
echo "   - lambda:GetFunction"
echo "   - lambda:ListFunctions"
echo ""
echo "2. DynamoDB permissions:"
echo "   - dynamodb:CreateTable"
echo "   - dynamodb:DescribeTable"
echo "   - dynamodb:ListTables"
echo "   - dynamodb:UpdateTable"
echo "   - dynamodb:Wait"
echo ""
echo "3. Optional (for verification):"
echo "   - dynamodb:DescribeContinuousBackups"
echo ""
echo "Step 2: Configure GitHub Secrets"
echo "---------------------------------"
echo ""
echo "Run the following commands to set up secrets (replace values as needed):"
echo ""
echo "  # Set AWS Access Key ID"
echo "  gh secret set AWS_ACCESS_KEY_ID --repo=\$(gh repo view --json nameWithOwner -q '.nameWithOwner') --body='YOUR_ACCESS_KEY_ID'"
echo ""
echo "  # Set AWS Secret Access Key"
echo "  gh secret set AWS_SECRET_ACCESS_KEY --repo=\$(gh repo view --json nameWithOwner -q '.nameWithOwner') --body='YOUR_SECRET_ACCESS_KEY'"
echo ""
echo "  # Set AWS Region (default is us-east-1)"
echo "  gh secret set AWS_REGION --repo=\$(gh repo view --json nameWithOwner -q '.nameWithOwner') --body='us-east-1'"
echo ""
echo "Step 3: Test the workflow"
echo "-------------------------"
echo "1. Push to the 'main' branch to trigger deployment"
echo "2. Check the Actions tab in your GitHub repository"
echo "3. Review the workflow logs for any issues"
echo ""
echo "Troubleshooting:"
echo "----------------"
echo "1. If DynamoDB table creation fails, check IAM permissions"
echo "2. If Lambda update fails, ensure functions exist in AWS"
echo "3. Check AWS CloudWatch logs for detailed errors"
echo ""
echo "Note: Single production environment configuration:"
echo "  - Table name: training-platform"
echo "  - Lambda functions: training-platform-{function-name}"
echo ""

# Offer to create the secrets interactively
read -p "Do you want to set up the AWS secrets now? (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "Enter AWS Access Key ID: " access_key
    if [ -n "$access_key" ]; then
        gh secret set AWS_ACCESS_KEY_ID --repo="$(gh repo view --json nameWithOwner -q '.nameWithOwner')" --body="$access_key"
        echo "✓ AWS_ACCESS_KEY_ID secret set successfully!"
    else
        echo "✗ No Access Key ID provided, skipping."
    fi
    
    read -p "Enter AWS Secret Access Key: " secret_key
    if [ -n "$secret_key" ]; then
        gh secret set AWS_SECRET_ACCESS_KEY --repo="$(gh repo view --json nameWithOwner -q '.nameWithOwner')" --body="$secret_key"
        echo "✓ AWS_SECRET_ACCESS_KEY secret set successfully!"
    else
        echo "✗ No Secret Access Key provided, skipping."
    fi
    
    read -p "Enter AWS Region (press Enter for us-east-1): " region
    region=${region:-us-east-1}
    gh secret set AWS_REGION --repo="$(gh repo view --json nameWithOwner -q '.nameWithOwner')" --body="$region"
    echo "✓ AWS_REGION secret set to: $region"
fi

echo ""
echo "Setup complete! Remember to:"
echo "1. Ensure your IAM User has the necessary permissions"
echo "2. Push changes to trigger the workflow"
echo "3. Monitor the Actions tab for any issues"
echo ""
