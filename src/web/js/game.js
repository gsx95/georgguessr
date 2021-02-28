let guessMap;
let resultMap;
let gameID;
let marker;
let streetview;
let guessPos;
let roundNo = 1;
let gameStats;
let correctMarker;
let distanceLine;
let guessMarker;
let secondsLeft;
let timerId;
let currentPano;
let startPos;
let timerStopped = true;
let playerName;

let markers = [];
let lines = [];

function showGame() {
    setGuessMapResizable();
    showPlayerNamePrompt();
    byId("refresh-results-btn").onclick = function() {showResults(false);};
}

function initGameMaps() {
    initUtils();
    byId("guess-btn").onclick = endRound;
    byId("result-btn").onclick = nextRound;
    gameID = getRequestParameter("id");

    guessMap = new google.maps.Map(byId("guess-map"), {
        center: {lat: 37.869260, lng: -122.254811},
        zoom: 1,
        fullscreenControl: true,
        mapTypeControl: false,
        streetViewControl: false,
    });

    resultMap = new google.maps.Map(byId("result-map"), {
        center: {lat: 37.869260, lng: -122.254811},
        zoom: 1,
        fullscreenControl: false,
        mapTypeControl: false,
        streetViewControl: false,
        zoomControl: false
    });


    guessMap.addListener("click", (data) => {
        guessPos = data.latLng;
        drawMarker(data.latLng);
        enableGuessButton();
    });

    streetview = new google.maps.StreetViewPanorama(document.getElementById("pano"),
        {
            addressControl: false,
            fullscreenControl: false,
            showRoadLabels: false,
            zoomControl: true,
            panControl: false,
        }
    );
    doGetRequestJSON("/game/stats/" + gameID, function (resp) {
        gameStats = resp;
        secondsLeft = gameStats.timeLimit;
        byId("round-no").innerText = "Round " + roundNo + "/" + gameStats.rounds;
        setStartView(roundNo);
        byId("to-start-btn").onclick = backHome;
    }, function (err) {
        console.log(err);
    });
}

function backHome() {
    streetview.setPano(currentPano);
}

function nextRound() {
    byId("game-controls").style.visibility = "visible";
    byId("guess-map-container").style.visibility = "visible";
    byId("stop-overlay").style.display = "none";
    byId("stop-popup").style.display = "none";

    distanceLine.setMap(null);
    guessMarker.setMap(null);
    correctMarker.setMap(null);
    marker.setMap(null);

    let guessBtn = byId("guess-btn");
    guessBtn.disabled = true;
    guessBtn.classList.add("btn-disabled");

    Object.assign(byId("guess-map-container").style, {
        width: "300px",
        height: "200px",
    });

    roundNo++;
    byId("round-no").innerText = "Round " + roundNo + "/" + gameStats.rounds;
    secondsLeft = gameStats.timeLimit;
    setStartView(roundNo);
}

function endRound() {
    clearInterval(timerId);
    byId("game-controls").style.visibility = "hidden";
    byId("guess-map-container").style.visibility = "hidden";
    byId("stop-overlay").style.display = "block";
    byId("stop-popup").style.display = "block";
    showResults(true);

}

function showResults(postResults) {
    for(let i = 0; i<markers.length;i++) {
        let m = markers[i];
        m.setMap(null);
    }
    for(let i = 0; i<lines.length;i++) {
        let l = lines[i];
        l.setMap(null);
    }


    correctMarker = new google.maps.Marker({
        position: startPos,
        icon: {
            size: new google.maps.Size(60, 30),
            scaledSize: new google.maps.Size(60, 30),
            url: "https://i.ibb.co/PgFftmS/flag-2.png"
        }
    });
    correctMarker.setMap(resultMap);

    let distances = [];

    if (guessPos !== null && guessPos !== undefined) {
        // show my guess marker
        guessMarker = showMarkerAndLine(guessPos, 0);
        var bounds = new google.maps.LatLngBounds();
        bounds.extend(guessMarker.getPosition());
        bounds.extend(correctMarker.getPosition());
        resultMap.fitBounds(bounds);


        let meters = calculateDistanceInMeter(guessMarker, correctMarker);
        if(postResults === true) {
            doPostRequest("/game/guess/" + gameID + "/" + roundNo + "/" + playerName, {
                "distance": meters,
                "guess": {
                    "lat": guessPos.lat(),
                    "lon": guessPos.lng()
                }
            }, function (resp) {
                console.log(resp);
            });
        }
        distances = [{"name": "Your", "distance": meters}];
    }

    doGetRequestJSON("/game/guesses/" + gameID + "/" + roundNo, function (response) {
        for (let name in response) {
            console.log("compare " + name + " with " + playerName);
            if(name.toLowerCase() === playerName.toLowerCase()) {
                console.log("continue");
                continue;
            }
            if (response.hasOwnProperty(name)) {
                console.log("add to list");
                let score = response[name];
                let dist = score["distance"];
                let pos = score["guess"];
                distances.push({
                    "name": name,
                    "distance": dist,
                    "lat": pos["lat"],
                    "lon": pos["lon"]
                });
            }
        }

        for(let i = 0; i<distances.length;i++) {
            let d = distances[i];
            let name = d["name"];
            let dist = d["distance"];

            if (dist < 1000) {
                d["text"] = name + ": " + dist + "m";
            } else if (dist < 100000) {
                let km = dist / 1000;
                let mets = dist % 1000;
                d["text"] = name + ": " + ~~km + "." + (("" + mets).substring(0, 1)) + "km";
            } else {
                let km = dist / 1000;
                d["text"] =  name + ": " + ~~km + "km";
            }
        }
        showResultDistances(distances);
    });
}

let icons = [
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

function showMarkerAndLine(pos, iconIndex) {
    console.log("show maker " + pos["lat"] + "   " + iconIndex);
    let marker = new google.maps.Marker({
        position: pos,
        icon: {
            size: new google.maps.Size(30, 52),
            scaledSize: new google.maps.Size(30, 52),
            url: icons[iconIndex]
        }
    });
    marker.setMap(resultMap);

    let line = new google.maps.Polyline({
        path: [pos, startPos],
        geodesic: true,
        strokeColor: '#ff9634',
        strokeOpacity: 1.0,
        strokeWeight: 2
    });

    line.setMap(resultMap);
    markers.push(marker);
    lines.push(line);
    return marker;
}

function showResultDistances(distanceTexts) {
    byId("result-text").innerHTML = "";
    for(let i = 0;i<distanceTexts.length;i++) {
        let d = distanceTexts[i];
        let text = d["text"];
        addElement("p", byId("result-text"), text);
        if(d["lat"] !== undefined && d["lat"] !== null) {
            showMarkerAndLine({lat: d["lat"], lng: d["lon"]}, i + 1)
        }
    }
}

function calculateDistanceInMeter(marker1, marker2){
    let distance = google.maps.geometry.spherical.computeDistanceBetween(marker1.getPosition(), marker2.getPosition());
    return ~~distance;
}

function setStartView(round) {
    updateStreetView(round, function () {
        updateTimer();
        timerId = setInterval(updateTimer, 1000);
    });

}

function updateTimer() {
    if(!timerStopped) {
        secondsLeft = secondsLeft - 1;
    }
    byId("timer").innerText = timeToString();
    if(secondsLeft === 0) {
        endRound();
    }
}

function timeToString() {
    let s = secondsLeft + 0;
    return "Time: " + (s-(s%=60))/60+(9<s?':':':0')+s;
}

function updateStreetView(round, callback) {
    doGetRequestJSON("/game/pos/" + gameID + "/" + round, function (resp) {
        streetview.setPano(resp.panoId);
        currentPano = resp.panoId;
        startPos = {lat: resp.lat, lng: resp.lon};
        callback();
    }, function (err) {
        console.log(err);
    });
}

function enableGuessButton() {
    let guessBtn = byId("guess-btn");
    guessBtn.removeAttribute("disabled");
    guessBtn.classList.remove("btn-disabled");
}

function drawMarker(latlng) {
    if (marker !== null && marker !== undefined) {
        marker.setMap(null);
    }
    marker = new google.maps.Marker({
        position: latlng,
    });
    marker.setMap(guessMap);
}

function setGuessMapResizable() {
    interact('.resizable').resizable({
        edges: {top: true, left: false, bottom: false, right: true},
        listeners: {
            move: function (event) {
                let {x, y} = event.target.dataset;

                x = (parseFloat(x) || 0) + event.deltaRect.left;
                y = (parseFloat(y) || 0) + event.deltaRect.top;

                event.rect.width = event.rect.width < 350 ? 300 : event.rect.width;
                event.rect.height = event.rect.height < 250 ? 200 : event.rect.height;

                Object.assign(event.target.style, {
                    width: `${event.rect.width}px`,
                    height: `${event.rect.height}px`,
                });

                Object.assign(event.target.dataset, {x, y})
            }
        }
    })
}

function startGame() {
    checkName(function() {
        playerName = byId("player-name-input").value;
        doPostRequest("/game/players/" + gameID + "/" + playerName);
        byId("insert-player-name-overlay").style.display = "none";
        byId("insert-player-name-container").style.display = "none";
        timerStopped = false;
    });
}

function showPlayerNamePrompt() {
    let input = byId("player-name-input");
    let startBtn = byId("start-game-button");

    startBtn.onclick = startGame;

    input.onblur=checkName;

    input.onkeyup = function() {
        if(input.value.length !== 0 && startBtn.disabled) {
            startBtn.removeAttribute("disabled");
            startBtn.classList.remove("btn-disabled");
            input.classList.remove("not-valid");
        } else if(input.value.length === 0 && !startBtn.disabled){
            startBtn.disabled = true;
            startBtn.classList.add("btn-disabled");
            input.classList.add("not-valid");
        }
    }
}

function checkName(validCallback, notValidCallback) {
    doGetRequestJSON("/game/players/" + gameID, function (response) {
        let input = byId("player-name-input");
        let startBtn = byId("start-game-button");
        let valid = response["players"].indexOf(input.value) < 0;
        if (!valid) {
            startBtn.disabled = true;
            startBtn.classList.add("btn-disabled");
            input.classList.add("not-valid");
            if(notValidCallback) {
                notValidCallback();
            }
            return;
        }
        if(validCallback) {
            validCallback();
        }
    }, function(err) {
        console.log(err);
    });
}