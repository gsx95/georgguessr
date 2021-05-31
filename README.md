### Prerequisites:

Install Go, AWS CLI, AWS SAM, npm

Also:

1. `npm install -g babel-cli`
2. `npm install -g babel-preset-minify --save-dev`
3. `npm install -g http-server` for local testing
4. copy `example_config.json` to `config.json` and add your API key

### Build
For your first deployment, do `./make.sh guided` so that your AWS SAM configuration can be saved to disk.


Build and deploy all:  `./make.sh`
* Uses AWS SAM to create complete backend infrastructure in your AWS Account. Displays Changeset before deploying anything.


Rebuild Frontend: `./make.sh frontend`
* Only rebuilds the frontend files

Rebuild Backend: `./make.sh update`
* Only rebuilds and deploys backend sources, without uploading initial data to dynamodb again
