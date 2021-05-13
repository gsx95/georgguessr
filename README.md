### Prerequisites:

Install Go, AWS CLI, AWS SAM, npm

Also:

1. `npm install -g babel-cli`
2. `npm install -g babel-preset-minify`
3. `npm install -g http-server` for local testing
4. copy `example_config.json` to `config.json` and add your API key

### Build
Build all:  `./make.sh prod`
* Uses AWS SAM to create complete backend infrastructure in your AWS Account. Displays Changeset before deploying anything.


Rebuild Frontend: `./make.sh prod frontend`
* Only rebuilds the frontend files

Rebuild Backend: `./make.sh prod update`
* Only rebuilds and deploys backend sources, without uploading initial data to dynamodb again
