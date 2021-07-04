import './style.css';

import Home from './home/home.js';
import Game from './game/game.js'
import EndResults from './results/results.js'

function startHome() {
    Home.GuessrHome.init();
}


function showGame() {
    Game.GuessrGame.showGame();
}

function initGameMaps() {
    Game.GuessrGame.initMaps();
}

function initEndResults() {
    EndResults.GuessrResults.initResults();
}

window.startHome = startHome;
window.GuessrHome = Home.GuessrHome;

window.showGame = showGame;
window.initGameMaps = initGameMaps;
window.GuessrGame = Game.GuessrGame;

window.initEndResults = initEndResults;
window.GuessrResults = EndResults.GuessrResults;