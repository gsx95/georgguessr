#!/usr/bin/env bash

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