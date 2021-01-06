#!/usr/bin/env bash

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
RES=$(aws cloudformation describe-stack-resources --stack-name georgguessr)
API_KEY_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::ApiKey") | .PhysicalResourceId')
API_KEY_VALUE=$(aws apigateway get-api-key --include-value --api-key ${API_KEY_ID} | jq -r '.value')
API_ID=$(echo ${RES} | jq -r '.StackResources[]  | select(.ResourceType == "AWS::ApiGateway::RestApi") | .PhysicalResourceId')
API_ENDPOINT="https://$API_ID.execute-api.eu-central-1.amazonaws.com/Prod"


echo "API KEY:  $API_KEY_VALUE"
echo "API ENDPOINT:  $API_ENDPOINT"

rm -r statics
mkdir -p statics
mkdir -p statics/js
mkdir -p statics/css
mkdir -p statics/img

cp src/web/css/styles.css statics/css/styles.css
cp src/web/img/* statics/img/

cp src/web/index.html statics/index.html;
API_KEY=$(cat config.json| jq -r ".google_api_key");
sed -i '' "s/{{google-api-key}}/$API_KEY/g" statics/index.html;

if [[ $1 = "prod" ]]; then
  babel src/web/js/* --presets minify > statics/js/georgguessr.js;
  sed -i '' "/dev-script/d" statics/index.html;
  sed -i '' "/PROD/d" statics/index.html;
else
  cp -R src/web/js/ statics/js/
fi

echo "------------------------"
echo "Done!"
echo ""