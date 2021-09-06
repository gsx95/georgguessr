#!/bin/bash -e

echo "------------------------"
echo "Building Frontend..."
echo "------------------------"

if [[ $1 = "remote" ]]
then
    RES=$(aws cloudformation describe-stack-resources --stack-name georgguessr)
    API_KEY_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::ApiKey") | .PhysicalResourceId')
    API_KEY_VALUE=$(aws apigateway get-api-key --include-value --api-key ${API_KEY_ID} | jq -r '.value')
    API_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::RestApi") | .PhysicalResourceId')
    API_ENDPOINT="https://$API_ID.execute-api.eu-west-1.amazonaws.com/Prod"
else
    API_KEY_VALUE=""
    API_ENDPOINT="http://127.0.0.1:3000"
fi

npm install
npx webpack --env apiKey="$API_KEY_VALUE" --env api="$API_ENDPOINT" --env mapsKey="$GUESSR_MAPS_API_KEY"

echo "------------------------"
echo "Done!"
echo ""
