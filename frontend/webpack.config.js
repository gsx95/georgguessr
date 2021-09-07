const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = env => ({
    entry: {
        index: './src/index.js',
     },
    plugins: [
          new HtmlWebpackPlugin({
                 filename: "index.html",
                 title: 'GeorgGuessr',
                  templateParameters: {
                      'API_EP': env.api,
                      'API_KEY': env.apiKey,
                      'MAPS_KEY': env.mapsKey,
                  },
                 template: './src/home/home.html',
       }),
        new HtmlWebpackPlugin({
            filename: "game/index.html",
            title: 'GeorgGuessr',
            templateParameters: {
                'API_EP': env.api,
                'API_KEY': env.apiKey,
                'MAPS_KEY': env.mapsKey,
            },
            template: './src/game/game.html',
        }),
        new HtmlWebpackPlugin({
            filename: "results/index.html",
            title: 'GeorgGuessr',
            templateParameters: {
                'API_EP': env.api,
                'API_KEY': env.apiKey,
                'MAPS_KEY': env.mapsKey,
            },
            template: './src/results/results.html',
        }),
    ],
    output: {
        filename: 'bundle.js',
        path: path.resolve(__dirname, './dist'),
        clean: true,
    },
    module: {
        rules: [
            {
                test: /\.css$/i,
                use: ['style-loader', 'css-loader'],
            },
            {
                test: /\.(png|svg|jpg|jpeg|gif)$/i,
                type: 'asset/resource',
            },
        ],

    },
});
