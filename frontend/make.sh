#!/bin/bash -e

echo "------------------------"
echo "Building Frontend..."
echo "------------------------"

BUILD_VERSION=$1
TARGET=$2
if [[ ${TARGET} = "remote" ]]
then
    RES=$(aws cloudformation describe-stack-resources --stack-name georgguessr)
    API_KEY_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::ApiKey") | .PhysicalResourceId')
    API_KEY_VALUE=$(aws apigateway get-api-key --include-value --api-key ${API_KEY_ID} | jq -r '.value')
    API_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::RestApi") | .PhysicalResourceId')
    API_ENDPOINT="https://$API_ID.execute-api.$AWS_DEFAULT_REGION.amazonaws.com/Prod"
else
    API_KEY_VALUE=""
    API_ENDPOINT="http://127.0.0.1:3000"
fi

npm install
npx webpack --env apiKey="$API_KEY_VALUE" --env api="$API_ENDPOINT" --env mapsKey="$GUESSR_MAPS_API_KEY" --env buildVersion="$BUILD_VERSION"

echo "------------------------"
echo "Done!"
echo ""
