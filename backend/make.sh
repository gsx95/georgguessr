#!/bin/bash -e

version=$1
target=$2
guided=$3

echo ${version}
echo ${version} > pkg/build.version
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
