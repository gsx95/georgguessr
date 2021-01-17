Prerequisites:

1. `npm install -g babel-cli`
2. `npm install -g babel-preset-minify`
3. copy `example_config.json` to `config.json` and add your API key

Build for Dev:  `./make.sh`

Build for Prod:  `./make.sh prod`

Only Build Prod Frontend: `./make.sh prod frontend`

How to update/replace `code.json` and `countries.json`:

1. Download Country Codes (slim-2) from here https://github.com/lukes/ISO-3166-Countries-with-Regional-Codes
2. Download Cities With Population from https://public.opendatasoft.com/explore/dataset/worldcitiespop/information/
3. Download Capitals from https://github.com/Stefie/geojson-world
4. run geo.py to generate `codes.json` and `countries.json`