let endResultsMap;

let resultMarkers = [];
let resultLines = [];
let playerIcons = {};
let allRounds = [];
let allPlayers = [];

function initEndResults() {
    gameID = getRequestParameter("id");
    initUtils();
    initMaps();
    doGetRequestJSON("/game/endresults/" + gameID, function (resp) {
        let results = resp;
        let roundsNum = results["rounds"];
        allPlayers = results["players"];
        let rounds = results["gameRounds"];

        let notFinishedMsg;

        if (rounds.length < roundsNum) {
            notFinishedMsg = "not all rounds played!"
        }

        let notFinishedPlayers = new Set([]);

        for (let i = 0; i < rounds.length; i++) {
            let round = rounds[i];
            let scores = round["scores"];
            for (let j = 0; j < allPlayers.length; j++) {
                let playerName = allPlayers[j];
                if (scores[playerName] === undefined || scores[playerName] === null) {
                    notFinishedPlayers.add(playerName);
                }
            }
        }

        if (notFinishedPlayers.size !== 0) {
            notFinishedMsg = "Some players did not finish yet: ";
            for (let player of notFinishedPlayers) {
                notFinishedMsg += " " + player + " "
            }
        }

        if (notFinishedMsg !== undefined && notFinishedMsg !== null) {
            byId("result-msg").innerText = notFinishedMsg;
        }

        buildResults(rounds, allPlayers);
        allRounds = rounds;
    }, function (err) {
        console.log(err);
    });
}

function buildResults(rounds, allPlayers) {



    let mapIcons = [
        "https://i.ibb.co/bQqCvPG/icon-0.png",
        "https://i.ibb.co/D98JKkn/icon-1.png",
        "https://i.ibb.co/HFfZ9WJ/icon-2.png",
        "https://i.ibb.co/KLpK0py/icon-3.png",
        "https://i.ibb.co/yn9FMtc/icon-4.png",
        "https://i.ibb.co/SVPGvD6/icon-5.png",
        "https://i.ibb.co/QHB2PZM/icon-6.png",
        "https://i.ibb.co/Dtf5VKJ/icon-7.png",
        "https://i.ibb.co/GHLgHKk/icon-8.png",
        "https://i.ibb.co/rxvXG5S/icon-9.png"
    ];

    for (let i = 0; i < allPlayers.length; i++) {
        let player = allPlayers[i];
        playerIcons[player] = mapIcons[i];
    }

    let roundsWithScores = roundsToScores(rounds);
    let resultContent = getResultContent(roundsWithScores, allPlayers, playerIcons);
    byId("result-table").innerHTML = getRenderedTemplate("ResultsTableRows", {"results": resultContent});

    placeMarkersForRound(rounds);

    let selectRounds = ["All"];
    for (let i = 0; i < rounds.length; i++) {
        selectRounds.push("" + (i + 1));
    }

    byId("select-round-btns").innerHTML = getRenderedTemplate("SelectRoundButtons", {"rounds": selectRounds});
    byId("select-round-btns").onchange = selectedRoundChanged;
}

function getResultContent(rounds, allPlayers, playerIcons) {

    let results = [];
    let wins = {};
    let distanceSums = {};

    for (let j = 0; j < allPlayers.length; j++) {
        wins[allPlayers[j]] = 0;
        distanceSums[allPlayers[j]] = 0;
    }

    for (let i = 0; i < rounds.length; i++) {
        let round = rounds[i];
        round = round.sort((a, b) => (a["distance"]["distance"] > b["distance"]["distance"]) ? 1 : ((b["distance"]["distance"] > a["distance"]["distance"]) ? -1 : 0));
        let winner = round[0];
        wins[winner["player"]] = wins[winner["player"]] + 1;

        for (let j = 0; j < round.length; j++) {
            let roundResult = round[j];
            let player = roundResult["player"];
            let distance = roundResult["distance"]["distance"];
            distanceSums[player] += distance;
        }
    }

    for (let i = 0; i < allPlayers.length; i++) {
        let player = allPlayers[i];
        results.push(
            {
                "name": decodeURIComponent(player),
                "wins": wins[player],
                "icon": playerIcons[player],
                "distance": distanceToText(distanceSums[player])
            }
            )
    }


    return results;
}

function initMaps() {
    endResultsMap = new google.maps.Map(byId("result-map"), {
        center: {lat: 37.869260, lng: -122.254811},
        zoom: 1,
        fullscreenControl: false,
        mapTypeControl: false,
        streetViewControl: false,
        zoomControl: false
    });
}

let colors = [
    "#ffb760",
    "#ff0511",
    "#1bbc43",
    "#ff37ba",
    "#29b6b6",
    "#d5d50b",
    "#4d3471",
    "#333da1",
    "#fffafe"
];

function selectedRoundChanged() {

    let selectedRound = byId("select-round-btns").value;
    let rounds = [];
    if (selectedRound === "All") {
        rounds = allRounds;
    } else {

        for (let i = 0; i < allRounds.length; i++) {
            let round = allRounds[i];
            if (round["No"] === selectedRound - 1) {
                rounds = [round];
                break;
            }
        }
    }

    placeMarkersForRound(rounds);
    let roundsWithScores = roundsToScores(rounds);
    let resultContent = getResultContent(roundsWithScores, allPlayers, playerIcons);
    byId("result-table").innerHTML = getRenderedTemplate("ResultsTableRows", {"results": resultContent});

}

function placeMarkersForRound(rounds) {
    for (let i = 0; i < resultMarkers.length; i++) {
        resultMarkers[i].setMap(null);
    }
    resultMarkers = [];
    for (let i = 0; i < resultLines.length; i++) {
        resultLines[i].setMap(null);
    }
    resultLines = [];

    endResultsMap.fitBounds(new google.maps.LatLngBounds(null));


    for (let i = 0; i < rounds.length; i++) {
        let round = rounds[i];
        let startPosition = round["startPosition"];
        let scores = round["scores"];

        let guesses = [];
        Object.keys(scores).forEach(function (player, index) {
            let guessPos = scores[player]["guess"];
            guesses.push({"name": player, "guess": guessPos});
        });

        placeMarkersAndLines(startPosition, guesses, colors[i]);
    }

    let bounds = new google.maps.LatLngBounds();
    for (let i = 0; i < resultMarkers.length; i++) {
        let marker = resultMarkers[i];
        bounds.extend(marker.getPosition());
    }
    endResultsMap.fitBounds(bounds);
}

function placeMarkersAndLines(startPosition, guesses, color) {
    let startPos = {lat: startPosition.lat, lng: startPosition.lon};
    let startMarker = new google.maps.Marker({
        position: startPos,
        icon: {
            size: new google.maps.Size(60, 30),
            scaledSize: new google.maps.Size(60, 30),
            url: "https://i.ibb.co/PgFftmS/flag-2.png"
        }
    });
    startMarker.setMap(endResultsMap);
    resultMarkers.push(startMarker);

    for (let i = 0; i < guesses.length; i++) {
        let guess = guesses[i];
        let guessPos = {lat: guess["guess"].lat, lng: guess["guess"].lon};
        let guessMarker = new google.maps.Marker({
            position: guessPos,
            icon: {
                size: new google.maps.Size(30, 52),
                scaledSize: new google.maps.Size(30, 52),
                url: playerIcons[guess["name"]]
            }
        });
        guessMarker.setMap(endResultsMap);
        resultMarkers.push(guessMarker);
        let line = new google.maps.Polyline({
            path: [guessPos, startPos],
            geodesic: true,
            strokeColor: color,
            strokeOpacity: 1.0,
            strokeWeight: 2
        });
        line.setMap(endResultsMap);
        resultLines.push(line);
    }
}

function roundsToScores(rounds) {
    let roundsWithScores = [];
    for (let i = 0; i < rounds.length; i++) {
        let round = rounds[i];
        let scores = round["scores"];
        let roundScores = [];

        Object.keys(scores).forEach(function (player, index) {
            let distance = scores[player];
            roundScores.push({"player": player, "distance": distance})
        });
        roundsWithScores.push(roundScores);
    }
    return roundsWithScores;
}