#!/bin/bash
cd /workdir
git clone https://github.com/gsx95/georgguessr.git
mv phantomjs-2.1.1-linux-x86_64/bin/phantomjs georgguessr/src/backend/layer/bin/phantomjs
cd georgguessr
./make.sh
