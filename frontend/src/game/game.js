import u from '../utils.js';

export default {

    GuessrGame: {

        roundNo: 1,
        timerStopped: true,
        gameEnded: false,

        markers: [],
        lines: [],

        // TODO: serve from own domain
        icons: [
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
        ],

        guessMap: null,
        resultMap: null,
        gameID: null,
        marker: null,
        streetview: null,
        guessPos: null,
        gameStats: null,
        secondsLeft: null,
        timerId: null,
        currentPano: null,
        startPos: null,
        playerName: null,

        showGame: function () {
            GuessrGame.gameID = u.getRequestParameter("id");
            u.byId("share-id").innerText = GuessrGame.gameID;
            console.log(u.byId("share-link"));
            u.byId("share-link").value = window.location.href;
            u.byId("copy-game-link-btn").onclick = GuessrGame.copyGameLink;
            GuessrGame.setGuessMapResizable();
            GuessrGame.showPlayerNamePrompt();
            u.byId("refresh-results-btn").onclick = function () {
                GuessrGame.showResults(false);
            };
        },

        copyGameLink: function () {
            u.byId("share-link").select();
            document.execCommand('copy');
        },

        initMaps: function () {
            GuessrGame.gameID = u.getRequestParameter("id");
            u.initUtils();
            u.byId("guess-btn").onclick = GuessrGame.endRound;
            u.byId("result-btn").onclick = GuessrGame.nextRound;

            GuessrGame.guessMap = new google.maps.Map(u.byId("guess-map"), {
                center: {lat: 37.869260, lng: -122.254811},
                zoom: 1,
                fullscreenControl: true,
                mapTypeControl: false,
                streetViewControl: false,
            });

            GuessrGame.resultMap = new google.maps.Map(u.byId("result-map"), {
                center: {lat: 37.869260, lng: -122.254811},
                zoom: 1,
                fullscreenControl: false,
                mapTypeControl: false,
                streetViewControl: false,
                zoomControl: false
            });


            GuessrGame.guessMap.addListener("click", (data) => {
                GuessrGame.guessPos = data.latLng;
                GuessrGame.drawMarker(data.latLng);
                GuessrGame.enableGuessButton();
            });

            GuessrGame.streetview = new google.maps.StreetViewPanorama(document.getElementById("pano"),
                {
                    addressControl: false,
                    fullscreenControl: false,
                    showRoadLabels: false,
                    zoomControl: true,
                    panControl: false,
                }
            );
            u.doGetRequestJSON("/game/stats/" + GuessrGame.gameID, function (resp) {
                GuessrGame.gameStats = resp;
                GuessrGame.secondsLeft = GuessrGame.gameStats.timeLimit;
                u.byId("round-no").innerText = "Round " + GuessrGame.roundNo + "/" + GuessrGame.gameStats.gameRounds.length;
                GuessrGame.setStartView(GuessrGame.roundNo);
                u.byId("to-start-btn").onclick = GuessrGame.backHome;
            }, function (err) {
                console.log(err);
            });
        },

        backHome: function () {
            GuessrGame.streetview.setPano(GuessrGame.currentPano);
        },

        nextRound: function () {
            if (GuessrGame.gameEnded) {
                return GuessrGame.showEndResults();
            }

            u.byId("game-controls").style.visibility = "visible";
            u.byId("guess-map-container").style.visibility = "visible";
            u.byId("stop-overlay").style.display = "none";
            u.byId("stop-popup").style.display = "none";
            u.byId("result-table").innerText = "";


            let guessBtn = u.byId("guess-btn");
            guessBtn.disabled = true;
            guessBtn.classList.add("btn-disabled");

            for (let i = 0; i < GuessrGame.markers.length; i++) {
                let m = GuessrGame.markers[i];
                m.setMap(null);
            }
            GuessrGame.markers = [];
            for (let i = 0; i < GuessrGame.lines.length; i++) {
                let l = GuessrGame.lines[i];
                l.setMap(null);
            }
            GuessrGame.lines = [];
            Object.assign(u.byId("guess-map-container").style, {
                width: "300px",
                height: "200px",
            });

            GuessrGame.roundNo++;
            u.byId("round-no").innerText = "Round " + GuessrGame.roundNo + "/" + GuessrGame.gameStats.gameRounds.length;
            GuessrGame.secondsLeft = GuessrGame.gameStats.timeLimit;
            GuessrGame.setStartView(GuessrGame.roundNo);
        },

        showEndResults: function () {
            window.location.href = "/results?id=" + GuessrGame.gameID;
        },

        endRound: function () {
            clearInterval(GuessrGame.timerId);
            u.byId("game-controls").style.visibility = "hidden";
            u.byId("guess-map-container").style.visibility = "hidden";
            u.byId("stop-overlay").style.display = "block";
            u.byId("stop-popup").style.display = "block";

            if (GuessrGame.roundNo === GuessrGame.gameStats.gameRounds.length) {
                u.byId("result-btn").innerText = "VIEW RESULTS";
                GuessrGame.gameEnded = true;
            }

            GuessrGame.showResults(true);
        },

        showResults: function (postResults) {
            for (let i = 0; i < GuessrGame.markers.length; i++) {
                let m = GuessrGame.markers[i];
                m.setMap(null);
            }
            for (let i = 0; i < GuessrGame.lines.length; i++) {
                let l = GuessrGame.lines[i];
                l.setMap(null);
            }

            let correctMarker = new google.maps.Marker({
                position: GuessrGame.startPos,
                icon: {
                    size: new google.maps.Size(60, 30),
                    scaledSize: new google.maps.Size(60, 30),
                    url: "https://i.ibb.co/PgFftmS/flag-2.png"
                }
            });
            GuessrGame.markers.push(correctMarker);
            correctMarker.setMap(GuessrGame.resultMap);

            let distances = [];

            if (GuessrGame.guessPos !== null && GuessrGame.guessPos !== undefined) {
                // show my guess marker
                let guessMarker = GuessrGame.showMarkerAndLine(GuessrGame.guessPos, GuessrGame.icons[0]);
                var bounds = new google.maps.LatLngBounds();
                bounds.extend(guessMarker.getPosition());
                bounds.extend(correctMarker.getPosition());
                GuessrGame.resultMap.fitBounds(bounds);
                GuessrGame.markers.push(guessMarker);

                let meters = GuessrGame.calculateDistanceInMeter(guessMarker, correctMarker);
                if (postResults === true) {
                    u.doPostRequest("/game/guess/" + GuessrGame.gameID + "/" + GuessrGame.roundNo + "/" + GuessrGame.playerName, {
                        "distance": meters,
                        "guess": {
                            "lat": GuessrGame.guessPos.lat(),
                            "lng": GuessrGame.guessPos.lng()
                        }
                    }, function (resp) {
                    }, function (error) {
                        console.log("Got error " + error)
                    }, 200);
                }
                distances = [{"name": "You", "distance": meters}];
            }

            u.doGetRequestJSON("/game/guesses/" + GuessrGame.gameID + "/" + GuessrGame.roundNo, function (response) {
                for (let name_raw in response) {
                    let name = decodeURIComponent(name_raw);
                    if (name.toLowerCase() === GuessrGame.playerName.toLowerCase()) {
                        continue;
                    }
                    if (response.hasOwnProperty(name_raw)) {
                        let score = response[name_raw];
                        let dist = score["distance"];
                        let pos = score["guess"];
                        console.log(pos);
                        distances.push({
                            "name": name,
                            "distance": dist,
                            "lat": pos["lat"],
                            "lon": pos["lng"]
                        });
                    }
                }

                for (let i = 0; i < distances.length; i++) {
                    let d = distances[i];
                    let dist = d["distance"];
                    d["icon"] = GuessrGame.icons[i];

                    d["distance_text"] = u.distanceToText(dist);
                }

                distances.sort((a, b) => (a["distance"] > b["distance"]) ? 1 : ((b["distance"] > a["distance"]) ? -1 : 0));
                console.log(distances);
                GuessrGame.showResultDistances(distances);
            });
        },

        showMarkerAndLine: function (pos, iconUrl) {
            let marker = new google.maps.Marker({
                position: pos,
                icon: {
                    size: new google.maps.Size(30, 52),
                    scaledSize: new google.maps.Size(30, 52),
                    url: iconUrl
                }
            });
            marker.setMap(GuessrGame.resultMap);

            let line = new google.maps.Polyline({
                path: [pos, GuessrGame.startPos],
                geodesic: true,
                strokeColor: '#ff9634',
                strokeOpacity: 1.0,
                strokeWeight: 2
            });

            line.setMap(GuessrGame.resultMap);
            GuessrGame.markers.push(marker);
            GuessrGame.lines.push(line);
            return marker;
        },

        showResultDistances: function (distances) {
            u.byId("result-table").innerHTML = "";
            let rows = u.getRenderedTemplate("ResultsTableRows", {"results": distances});
            u.byId("result-table").innerHTML = rows;

            for (let i = 0; i < distances.length; i++) {
                let d = distances[i];
                if (d["lat"] !== undefined && d["lat"] !== null) {
                    console.log(d);
                    GuessrGame.showMarkerAndLine({lat: d["lat"], lng: d["lon"]}, d["icon"])
                }
            }
        },

        calculateDistanceInMeter: function (marker1, marker2) {
            let distance = google.maps.geometry.spherical.computeDistanceBetween(marker1.getPosition(), marker2.getPosition());
            return ~~distance;
        },

        setStartView: function (round) {
            GuessrGame.updateStreetView(round, function () {
                GuessrGame.updateTimer();
                GuessrGame.timerId = setInterval(GuessrGame.updateTimer, 1000);
            });

        },

        updateTimer: function () {
            if (!GuessrGame.timerStopped) {
                GuessrGame.secondsLeft = GuessrGame.secondsLeft - 1;
            }
            u.byId("timer").innerText = GuessrGame.timeToString();
            if (GuessrGame.secondsLeft === 0) {
                GuessrGame.endRound();
            }
        },

        timeToString: function () {
            let s = GuessrGame.secondsLeft + 0;
            return "Time: " + (s - (s %= 60)) / 60 + (9 < s ? ':' : ':0') + s;
        },

        updateStreetView: function (round, callback) {
            u.doGetRequestJSON("/game/pos/" + GuessrGame.gameID + "/" + round, function (resp) {
                console.log(resp);
                if (resp.areas !== undefined && resp.areas !== null) {
                    GuessrGame.showAreasInGuessMap(resp.areas)
                }
                GuessrGame.streetview.setPano(resp.panoId);
                GuessrGame.currentPano = resp.panoId;
                GuessrGame.startPos = {lat: resp.lat, lng: resp.lon};
                callback();
            }, function (err) {
                console.log(err);
            });
        },

        enableGuessButton: function () {
            let guessBtn = u.byId("guess-btn");
            guessBtn.removeAttribute("disabled");
            guessBtn.classList.remove("btn-disabled");
        },

        showAreasInGuessMap: function (areas) {
            let outerBounds = [ // whole world
                new google.maps.LatLng(-85.1054596961173, -180),
                new google.maps.LatLng(85.1054596961173, -180),
                new google.maps.LatLng(85.1054596961173, 180),
                new google.maps.LatLng(-85.1054596961173, 180),
                new google.maps.LatLng(-85.1054596961173, 0)
            ];
            let allBounds = [];
            allBounds.push(outerBounds);
            for (let i = 0; i < areas.length; i++) {
                let area = areas[i];
                let gArea = [];
                for (let j = 0; j < area.length; j++) {
                    let point = area[j];
                    let gPoint = new google.maps.LatLng(point.lat, point.lng);
                    gArea.push(gPoint);
                }
                if (google.maps.geometry.spherical.computeSignedArea(gArea) >= 0) {
                    allBounds.push(gArea);
                } else {
                    allBounds.push(gArea.reverse());
                }
            }
            new google.maps.Polygon({
                paths: allBounds,
                strokeColor: "#000000",
                strokeOpacity: 0.2,
                strokeWeight: 3,
                map: GuessrGame.guessMap
            });

        },

        drawMarker: function (latlng) {
            if (GuessrGame.marker !== null && GuessrGame.marker !== undefined) {
                GuessrGame.marker.setMap(null);
            }
            GuessrGame.marker = new google.maps.Marker({
                position: latlng,
            });
            GuessrGame.marker.setMap(GuessrGame.guessMap);
            GuessrGame.markers.push(GuessrGame.marker);
        },

        setGuessMapResizable: function () {
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
        },

        startGame: function () {
            GuessrGame.checkName(function () {
                GuessrGame.playerName = u.byId("player-name-input").value;
                u.doPostRequest("/game/players/" + GuessrGame.gameID + "/" + GuessrGame.playerName, null,
                    function (resp) { }, function (err) {
                        console.log("Got error " + err)
                    }, 200);
                u.byId("insert-player-name-overlay").style.display = "none";
                u.byId("insert-player-name-container").style.display = "none";
                GuessrGame.timerStopped = false;
            });
        },

        showPlayerNamePrompt: function () {
            let input = u.byId("player-name-input");
            let startBtn = u.byId("start-game-button");

            startBtn.onclick = GuessrGame.startGame;

            input.onblur = GuessrGame.checkName;

            input.onkeyup = function () {
                if (input.value.length !== 0 && startBtn.disabled) {
                    startBtn.removeAttribute("disabled");
                    startBtn.classList.remove("btn-disabled");
                    input.classList.remove("not-valid");
                } else if (input.value.length === 0 && !startBtn.disabled) {
                    startBtn.disabled = true;
                    startBtn.classList.add("btn-disabled");
                    input.classList.add("not-valid");
                }
            }
        },

        checkName: function (validCallback, notValidCallback) {
            u.doGetRequestJSON("/game/players/" + GuessrGame.gameID, function (response) {
                let input = u.byId("player-name-input");
                let startBtn = u.byId("start-game-button");
                let players = response["players"];
                let valid = true;
                for (let i = 0; i < players.length; i++) {
                    if (input.value.toLowerCase() === decodeURIComponent(players[i]).toLowerCase()) {
                        valid = false;
                    }
                }
                if (!valid) {
                    startBtn.disabled = true;
                    startBtn.classList.add("btn-disabled");
                    input.classList.add("not-valid");
                    if (notValidCallback) {
                        notValidCallback();
                    }
                    return;
                }
                if (validCallback) {
                    validCallback();
                }
            }, function (err) {
                console.log(err);
            });
        },
    }
}