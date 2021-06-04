#!/usr/bin/env bash

mode=$1

build_and_deploy_backend() {
    echo "------------------------"
    echo "Building Backend..."
    echo "------------------------"
    sam build
    if (( $? != 0 )); then
        exit
    fi
    echo "------------------------"
    echo "Done!"
    echo ""
    echo "------------------------"
    echo "Deploying Backend..."
    echo "------------------------"
    if [[ ${mode} = "guided" ]]; then
        sam deploy --guided
    else
        sam deploy
    fi

    if (( $? != 0 )); then
        exit
    fi
    echo "------------------------"
    echo "Done!"
    echo ""
}

build_frontend() {
    echo "------------------------"
    echo "Building Frontend..."
    echo "------------------------"

    RES=$(aws cloudformation describe-stack-resources --stack-name georgguessr)
    API_KEY_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::ApiKey") | .PhysicalResourceId')
    API_KEY_VALUE=$(aws apigateway get-api-key --include-value --api-key ${API_KEY_ID} | jq -r '.value')
    API_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::RestApi") | .PhysicalResourceId')
    API_ENDPOINT="https://$API_ID.execute-api.eu-central-1.amazonaws.com/Prod"

    MAPS_API_KEY=$(cat config.json| jq -r ".maps_api_key");

    echo "---"
    echo "$API_KEY_VALUE"
    echo "---"

    cd src/frontend
    npx webpack --env apiKey="$API_KEY_VALUE" --env api="$API_ENDPOINT" --env mapsKey="$MAPS_API_KEY"
    cd ../../

    echo "------------------------"
    echo "Done!"
    echo ""
}

upload_geo_data_to_dynamodb() {
    echo "------------------------"
    echo "Uploading to DynamoDB..."
    echo "------------------------"

    while read l1; do
        aws dynamodb batch-write-item --request-items "$l1";
    done <src/resources/cities.jsonData

    echo "------------------------"
    echo "Done!"
    echo ""
    echo ""
}

###################################

if [[ ${mode} != "frontend" ]]; then
    build_and_deploy_backend
fi

build_frontend

if [[ ${mode} = "frontend" ]]; then
    exit 0
fi


if [[ ${mode} = "update" ]]; then
    exit 0
fi

upload_geo_data_to_dynamodb

echo "Run the frontend local with ' http-server dist'"