Prerequisites:

1. `npm install -g babel-cli`
2. `npm install -g babel-preset-minify`
3. `npm install -g http-server` for local testing
4. copy `example_config.json` to `config.json` and add your API key

Build all:  `./make.sh prod`

Rebuild Frontend: `./make.sh prod frontend`

Update Backend without uploading to dynamodb again: `./make.sh prod update`
