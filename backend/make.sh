#!/bin/bash -e

target=$1
guided=$2

cwd=$(pwd)
cd layer
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
date +"%Y%m%d%H%M%S" > pkg/build.version
awk -v key=$GUESSR_MAPS_API_KEY '{gsub("<%= MAPS_KEY %>", key, $0); print}' layer/bin/template.html > layer/bin/index.html
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

if [[ ${target} = "remote" ]]
then
    echo ${guided}
    echo "------------------------"
    echo "Deploying Backend..."
    echo "------------------------"
    if [[ ${guided} = "guided" ]]; then
        rm -f samconfig.toml
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
