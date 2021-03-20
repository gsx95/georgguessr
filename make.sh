#!/usr/bin/env bash

declare -a pids

if [[ $1 = "example" ]]; then
    echo "------------------------"
    echo "Uploading example..."
    echo "------------------------"
    aws dynamodb batch-write-item --request-items file://src/resources/exampleRoom.json
    echo "<script>window.location.href = 'file:///$PWD/statics/createRoom.html?id=01a67502-c1a1-47e7-88f0-d3fc366c543c'</script>" > temp.html
    open temp.html
    sleep 5
    rm temp.html
    exit 0
fi

if [[ $2 != "frontend" ]]; then
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
    sam deploy
    if (( $? != 0 )); then
        exit
    fi
    echo "------------------------"
    echo "Done!"
    echo ""
    echo "------------------------"
    echo "Building Frontend..."
    echo "------------------------"
fi
RES=$(aws cloudformation describe-stack-resources --stack-name georgguessr)
API_KEY_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::ApiKey") | .PhysicalResourceId')
API_KEY_VALUE=$(aws apigateway get-api-key --include-value --api-key ${API_KEY_ID} | jq -r '.value')
API_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::RestApi") | .PhysicalResourceId')
API_ENDPOINT="https://$API_ID.execute-api.eu-central-1.amazonaws.com/Prod"

rm -r statics
mkdir -p statics
mkdir -p statics/js
mkdir -p statics/css
mkdir -p statics/img

cp src/web/css/styles.css statics/css/styles.css
cp src/web/css/gamestyles.css statics/css/gamestyles.css
cp src/web/css/results.css statics/css/results.css
cp -r src/web/img/* statics/img/

cp src/web/index.html statics/index.html;
cp src/web/game.html statics/game.html;
cp src/web/createRoom.html statics/createRoom.html;
cp src/web/results.html statics/results.html;

MAPS_API_KEY=$(cat config.json| jq -r ".maps_api_key");
sed -i '' "s/{{maps-api-key}}/$MAPS_API_KEY/g" statics/index.html;
sed -i '' "s,{{api-endpoint}},$API_ENDPOINT,g" statics/index.html;
sed -i '' "s/{{api-key}}/$API_KEY_VALUE/g" statics/index.html;

sed -i '' "s/{{maps-api-key}}/$MAPS_API_KEY/g" statics/game.html;
sed -i '' "s,{{api-endpoint}},$API_ENDPOINT,g" statics/game.html;
sed -i '' "s/{{api-key}}/$API_KEY_VALUE/g" statics/game.html;

sed -i '' "s/{{maps-api-key}}/$MAPS_API_KEY/g" statics/createRoom.html;
sed -i '' "s,{{api-endpoint}},$API_ENDPOINT,g" statics/createRoom.html;
sed -i '' "s/{{api-key}}/$API_KEY_VALUE/g" statics/createRoom.html;

sed -i '' "s/{{maps-api-key}}/$MAPS_API_KEY/g" statics/results.html;
sed -i '' "s,{{api-endpoint}},$API_ENDPOINT,g" statics/results.html;
sed -i '' "s/{{api-key}}/$API_KEY_VALUE/g" statics/results.html;

if [[ $1 = "prod" ]]; then
  babel src/web/js/* --presets minify > statics/js/georgguessr.js;
  sed -i '' "/dev-script/d" statics/index.html;
  sed -i '' "/PROD/d" statics/index.html;
  sed -i '' "/dev-script/d" statics/game.html;
  sed -i '' "/PROD/d" statics/game.html;
  sed -i '' "/dev-script/d" statics/createRoom.html;
  sed -i '' "/PROD/d" statics/createRoom.html;
  sed -i '' "/dev-script/d" statics/results.html;
  sed -i '' "/PROD/d" statics/results.html;
else
  cp -R src/web/js/ statics/js/
fi

echo "------------------------"
echo "Done!"
echo ""

if [[ $2 = "frontend" ]]; then
    exit 0
fi


if [[ $2 = "update" ]]; then
    exit 0
fi


echo "------------------------"
echo "Uploading to DynamoDB..."
echo "------------------------"
aws dynamodb batch-write-item --request-items file://src/resources/continents.json


while read l1; do
    aws dynamodb batch-write-item --request-items "$l1";
done <src/resources/cities.jsonData

while read l2; do
    aws dynamodb batch-write-item --request-items "$l2";
done <src/resources/countries.jsonData

echo "------------------------"
echo "Done!"
echo ""

open statics/index.html