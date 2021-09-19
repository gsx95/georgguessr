
<h1 align="center">
  <br>
  <a href="https://github.com/gsx95/georgguessr"><img src="https://gsx95-public-assets.s3.eu-west-1.amazonaws.com/logo.png" alt="GeorgGuessr" width="400"></a>
</h1>

<h4 align="center">An open source clone of <a href="https://geoguessr.com" target="_blank">Geoguessr</a> for you to use.</h4>

<p align="center">
</p>

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#how-to-build">How To Use</a> •
  <a href="#license">Development</a> •
  <a href="#license">License</a>
</p>
<p align="center">
<img src="https://gsx95-public-assets.s3.eu-west-1.amazonaws.com/demo.gif" alt="Demo" width="600" style="border-radius:5px"/>
</p>

## Disclaimer

This project is work in progress. Feel free to open an <a href="https://github.com/gsx95/georgguessr/issues">issue</a> if you encounter any bugs.

## Key Features

* Play Alone or with friends
  - 1 to 10 players
  - set time limit from 30s to 20m or no time limit at all
* Play in your favorite places
  - Search for places
  - draw custom play areas
  - play in your country or continent
  - just play randomly in the whole world

<p align="middle">
<img src="https://gsx95-public-assets.s3.eu-west-1.amazonaws.com/places.gif" width="40%" style="border-radius:5px" />
<img src="https://gsx95-public-assets.s3.eu-west-1.amazonaws.com/areas.gif" width="40%" style="border-radius:5px"/> 
</p>

* Overview of all player's picks after each round
  - see which locations your friends picked and how far they were off

<p align="middle">
<img src="https://gsx95-public-assets.s3.eu-west-1.amazonaws.com/summary1.png" width="40%" style="border-radius:5px"/>
<img src="https://gsx95-public-assets.s3.eu-west-1.amazonaws.com/summary2.png" width="40%" style="border-radius:5px"/> 
</p>

* Various options for streetview  (TODO)
  - show street names?
  - show compass?
  - building information on mini map?
* No login required
  - Share games and invite friends with a simple link
* Easy and cheap hosting in AWS
  - Easy: Fully serverless - uses AWS SAM to deploy the backend lambdas and the static HTML files for the frontend
  - Cheap: Uses AWS Lambda, API Gateway, DynamoDB for cheap hosting (first ~50k games per month fall into the AWS free tier)


## How to build

To clone and run this application, you'll need [npm](https://www.npmjs.com), [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html), [zip](https://formulae.brew.sh/formula/zip) and [GoLang >1.16](https://golang.org) installed on your computer. 

### Deploy to AWS

From your command line:

```bash
# Clone this repository
$ git clone git@github.com:gsx95/georgguessr.git

# Go into the repository
$ cd georgguessr

# Build and deploy to your AWS account. Use 'guided' target the first time to configure your deployment options.
# If you opted to save your config to a file, just use 'make remote' in the future.
$ make guided

```

### Deploy locally

For developing and debugging purposes, you can spin up the full stack locally.

```bash
# run a local dynamodb
$ docker-compose -f test/docker-compose-dynamodb.yml up -d

# run the backend locally and build frontend
$ make local

# run frontend server
$ http-server frontend/dist

```
## License

MIT
