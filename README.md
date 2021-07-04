### Prerequisites:

Install Go, AWS CLI, AWS SAM, npm

Also:

1. `npm install -g http-server` for local testing
2. copy `example_config.json` to `config.json` and add your API key
3. `cd src/frontend; npm install`

### Build
##### Has to be deployed in `eu-west-1` for now !

For your first deployment, do `./make.sh guided` so that your AWS SAM configuration can be saved to disk.

After that:

Build and deploy all:  `./make.sh`
* Uses AWS SAM to create complete backend infrastructure in your AWS Account. Displays Changeset before deploying anything.


Rebuild Frontend: `./make.sh frontend`
* Only rebuilds the frontend

Rebuild Backend: `./make.sh backend`
* Only rebuilds and deploys the backend
