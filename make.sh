#!/bin/bash -e

local=$1
mode=$2

if [[ -f config.env ]]
then
    export $(cat config.env | sed 's/#.*//g' | xargs)
fi

build_and_deploy_backend() {

    IS_LOCAL=$1
    cwd=$(pwd)
    cd src/backend/layer
    if [[ ! -f ./bin/phantomjs ]]; then
        mkdir -p tmp
        cd tmp
        wget https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-2.1.1-linux-x86_64.tar.bz2
        tar -xf ./phantomjs-2.1.1-linux-x86_64.tar.bz2 -C ./
        mv phantomjs-2.1.1-linux-x86_64/bin/phantomjs ../bin/phantomjs
        cd ..
        rm -rf tmp
    fi
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

    if [[ -z "$IS_LOCAL" ]]
    then
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
    else
        echo "============================================================"
        echo "====                                                    ===="
        echo "==== DON'T FORGET TO START THE DYNAMO DB DOCKER COMPOSE ===="
        echo "====                                                    ===="
        echo "============================================================"
        sam local start-api --docker-network sam-local-network
    fi
}

build_frontend() {
    echo "------------------------"
    echo "Building Frontend..."
    echo "------------------------"

    IS_LOCAL=$1
    if [[ -z "$IS_LOCAL" ]]
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

    cd src/frontend
    npm install
    npx webpack --env apiKey="$API_KEY_VALUE" --env api="$API_ENDPOINT" --env mapsKey="$GUESSR_MAPS_API_KEY"
    cd ../../

    echo "------------------------"
    echo "Done!"
    echo ""
}

###################################
if [[ ${local} != "local" ]]; then
    echo "deploying remote"
    if [[ ${mode} != "frontend" ]]; then
        build_and_deploy_backend
    fi
    if [[ ${mode} != "backend" ]]; then
        build_frontend
    fi
else
    echo "running locally"
    build_frontend "$local"
    build_and_deploy_backend "$local"
fi