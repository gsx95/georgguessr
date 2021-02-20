Prerequisites:

1. `npm install -g babel-cli`
2. `npm install -g babel-preset-minify`
3. copy `example_config.json` to `config.json` and add your API key

Build for Dev:  `./make.sh`

Build for Prod:  `./make.sh prod`

Only Build Prod Frontend: `./make.sh prod frontend`

Update Infrastructure without uploading to dynamodb again: `./make.sh prod update`
