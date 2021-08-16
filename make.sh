#!/bin/bash -e

mode=$1
if [[ -f config.env ]]
then
    export $(cat config.env | sed 's/#.*//g' | xargs)
fi

build_and_deploy_backend() {

    cwd=$(pwd)
    cd src/backend/layer
    zip -r layer.zip bin
    cd $cwd
    awk -v key=$GUESSR_MAPS_API_KEY '{gsub("<%= MAPS_KEY %>", key, $0); print}' src/backend/layer/bin/template.html > src/backend/layer/bin/index.html
    echo "------------------------"
    echo "Building Backend..."
    echo "------------------------"
    if (( $? != 0 )); then
        exit
    fi
    echo "Build Go Lambdas..."
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
        rm samconfig.toml
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
    API_ENDPOINT="https://$API_ID.execute-api.eu-west-1.amazonaws.com/Prod"


    echo "---"
    echo "$API_KEY_VALUE"
    echo "---"

    cd src/frontend
    npm install
    npx webpack --env apiKey="$API_KEY_VALUE" --env api="$API_ENDPOINT" --env mapsKey="$GUESSR_MAPS_API_KEY"
    cd ../../

    echo "------------------------"
    echo "Done!"
    echo ""
}

###################################

if [[ ${mode} != "frontend" ]]; then
    build_and_deploy_backend
fi

if [[ ${mode} != "backend" ]]; then
    build_frontend
fi

echo "Run the frontend local with ' http-server dist '"