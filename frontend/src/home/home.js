import {countries} from "countries-list";
import u from "../utils";
import ProgressBar from 'progressbar.js';

export default {
    GuessrHome: {

        selectedAreas: [],
        selectedPolygon: null,
        selectionMap: null,
        progress: null,

        init: function () {
            u.initUtils();

            MicroModal.init();
            GuessrHome.progress = new ProgressBar.Line('#loading-modal-content', {
                color: "#00449e",
                trailColor: '#98add2',
                strokeWidth: 3,
                trailWidth: 3,
            });
            GuessrHome.progress.set(0);

            console.log(document.getElementById("BUILD_VERSION").value);

            u.byId("create-room-btn").onclick = GuessrHome.showCreateRoom;
            u.byId("enter-room-btn").onclick = GuessrHome.showEnterRoom;
            u.byId("create-and-play-btn").onclick = GuessrHome.createRoom;
            u.byId("enter-btn").onclick = GuessrHome.enterRoom;

            GuessrHome.initGoogleMaps();
        },

        initGoogleMaps: function () {
            GuessrHome.initSearch();
            GuessrHome.selectionMap = new google.maps.Map(document.getElementById('selection-map'), {
                center: {lat: 37.869260, lng: -122.254811},
                zoom: 1,
                fullscreenControl: true,
                mapTypeControl: false,
                streetViewControl: false,
            });

            let selectArea = function (polygon) {
                for (let i = 0; i < GuessrHome.selectedAreas.length; i++) {
                    setPolygonColor(GuessrHome.selectedAreas[i], '#000000')
                }
                setPolygonColor(polygon, '#ff0000')
                GuessrHome.selectedPolygon = polygon
            };

            let setPolygonColor = function (polygon, color) {
                let options = {
                    strokeColor: color,
                    strokeOpacity: 0.8,
                    strokeWeight: 3,
                    fillColor: color,
                    fillOpacity: 0.35
                };
                options.paths = polygon.getPaths().getArray();
                polygon.setOptions(options);
            };

            let deleteSelectedPolygon = function () {
                GuessrHome.selectedPolygon.setMap(null);
                for (let i = 0; i < GuessrHome.selectedAreas.length; i++) {
                    if (GuessrHome.selectedAreas[i] === GuessrHome.selectedPolygon) {
                        GuessrHome.selectedAreas.splice(i, 1);
                        break;
                    }
                }
                GuessrHome.selectedPolygon = null;
            };

            window.addEventListener('keydown', function (e) {
                if (e.keyCode !== 46 && e.keyCode !== 8) {
                    return
                }
                if (GuessrHome.selectedPolygon === null || GuessrHome.selectedPolygon === undefined) {
                    return
                }
                deleteSelectedPolygon();
                event.preventDefault();
            });

            GuessrHome.drawingManager = new google.maps.drawing.DrawingManager({
                drawingMode: google.maps.drawing.OverlayType.POLYGON,
                drawingControl: true,
                drawingControlOptions: {
                    position: google.maps.ControlPosition.TOP_CENTER,
                    drawingModes: ['polygon']
                },
                markerOptions: {icon: 'https://developers.google.com/maps/documentation/javascript/examples/full/images/beachflag.png'},
            });

            google.maps.event.addListener(GuessrHome.drawingManager, 'overlaycomplete', function (event) {
                if (event.type === 'polygon') {
                    event.overlay.setMap(null);
                    if (event.type === 'polygon') {
                        GuessrHome.selectedAreas.push(new google.maps.Polygon({
                            paths: event.overlay.getPath().getArray(),
                            strokeColor: '#000000',
                            strokeOpacity: 0.8,
                            strokeWeight: 3,
                            fillColor: '#000000',
                            fillOpacity: 0.35
                        }));
                    }
                    GuessrHome.selectedAreas[GuessrHome.selectedAreas.length - 1].setMap(GuessrHome.selectionMap);
                    GuessrHome.selectedAreas[GuessrHome.selectedAreas.length - 1].addListener('click', function (e, p1, p2) {
                        selectArea(this)
                    });
                }
            });

            GuessrHome.drawingManager.setMap(GuessrHome.selectionMap);
        },

        enterRoom: function () {
            let id = u.byId("enter-room-id").value;
            u.byId("enter-room-id").setCustomValidity("");
            if (id.length === 0) {
                GuessrHome.showRoomIDInvalid();
                return;
            }
            u.doGetRequest("/exists/" + id, function (response) {
                if (response.status !== 200) {
                    GuessrHome.showRoomIDIsnvalid();
                    return;
                }
                window.location.href = "/game?id=" + id;
            }, function (response) {
                GuessrHome.showRoomIDInvalid();
            });
        },

        showRoomIDInvalid: function () {
            u.byId("enter-room-id").setCustomValidity("No room found with this id.");
            u.byId("enter-room-id").reportValidity();
        },

        createRoom: function () {

            let timeSliderValue = u.byId("timelimit").value;
            let timeSeconds = 0;

            if (timeSliderValue < 5) {
                timeSeconds = timeSliderValue * 10;
            } else {
                timeSeconds = (timeSliderValue - 5) * 60;
            }

            const data = {
                "maxRounds": parseInt(u.byId("rounds").value),
                "timeLimit": timeSeconds,
                "maxPlayers": parseInt(u.byId("maxplayer").value)
            };

            const select = u.byId("set-geo-limits");
            switch (select.value) {
                case "unlimited":
                    return GuessrHome.createRoomUnlimited(data);
                case "list":
                    return GuessrHome.createRoomList(data);
                case "place-search":
                    return GuessrHome.createRoomPlaceSearch(data);
                case "custom":
                    return GuessrHome.createRoomCustom(data);
            }
        },

        createRoomPlaceSearch: function (data) {
            let placesElements = document.getElementsByName("selected-places-item");
            let places = [];
            for (let i = 0; i < placesElements.length; i++) {
                let placeElement = placesElements[i];
                let place = placeElement.getAttribute("place");
                let country = placeElement.getAttribute("country");
                let location = JSON.parse(placeElement.getAttribute("location"));
                places.push({
                    "name": place,
                    "country": country,
                    "location": location
                })
            }
            data["places"] = places;
            GuessrHome.createRoomAndRedirect("places", data);
        },

        createRoomCustom: function (data) {
            let areasData = [];
            for (let i = 0; i < GuessrHome.selectedAreas.length; i++) {
                areasData[i] = [];
                let area = GuessrHome.selectedAreas[i];
                let path = area.getPath().getArray();
                for (let j = 0; j < path.length; j++) {
                    areasData[i][j] = path[j].toJSON();
                }
            }
            data["areas"] = areasData;
            GuessrHome.createRoomAndRedirect("custom", data);
        },

        createRoomList: function (data) {
            data["country"] = u.byId("country-selector").value;
            data["continent"] = u.byId("continent-selector").value;
            GuessrHome.createRoomAndRedirect("list", data);
        },

        createRoomUnlimited: function (data) {
            GuessrHome.createRoomAndRedirect("unlimited", data)
        },

        createRoomAndRedirect: function(roomType, data) {
            MicroModal.close('modal-1');
            MicroModal.show('loading-modal');
            u.doPostRequest("/rooms?type=" + roomType, data, function (roomData) {
                GuessrHome.progress.set(0.05);
                GuessrHome.generateStreetviews(roomData["roomId"], roomData["generatedPositions"], data["maxRounds"])
            }, function(errorMsg) {
                console.log(errorMsg);
                GuessrHome.showCreateErrorMsg(errorMsg)
            }, 200);
        },

        generateStreetviews: function(roomId, generatedPositions, roundsNum) {
            let svService = new google.maps.StreetViewService();
            let panos = [];
            let successCount = 0;
            let count = generatedPositions.length;

            for (let i = 0; i < count; i++) {
                let pos = generatedPositions[i];
                getStreetViewForPos(svService, i, pos.lat, pos.lng, 0, function (round, dist, panoId, latLng) {
                    if(panoId == null) {
                    } else {
                        panos.push({id: panoId, pos: latLng});
                        successCount++;
                        let progr = Math.min(successCount / roundsNum, 0.8);
                        GuessrHome.progress.set(progr);
                    }
                });
            }

            function getStreetViewForPos(svService, round, lat, lon, count, callback) {
                let searchRad = [49, 100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000];
                svService.getPanorama({
                    location: {"lat": lat, "lng": lon},
                    "radius": searchRad[count],
                    "source": "outdoor"
                }, function (data, status) {
                    checkStreetView(data, status, function(data) {
                        callback(round, searchRad[count], data.location.pano, data.location.latLng.toJSON());
                    }, function(errorMsg) {
                        if (count + 1 === searchRad.length) {
                            callback(round, null, null, null);
                        } else {
                            getStreetViewForPos(svService, round, lat, lon, count + 1, callback);
                        }
                    });
                });
            }

            function checkStreetView(data, status, okCallback, errorCallback) {
                if (status !== "OK") {
                    errorCallback("status wrong");
                    return;
                }
                if(data.links.length !== 2) {
                    errorCallback("not two links");
                    return;
                }
                return okCallback(data);
            }

            function waitForProcessToFinish() {
                if (successCount >= roundsNum) {
                    GuessrHome.progress.set(0.9);
                    let data = {
                        "roomId": roomId,
                        "panoramas": panos.slice(0, roundsNum),
                    };
                    u.doPostRequest("/panoramas", data, function () {
                        GuessrHome.progress.set(1);
                        window.location.href = "/game?id=" + roomId;
                    }, function(errorMsg) {
                        console.log(errorMsg);
                        GuessrHome.showCreateErrorMsg(errorMsg)
                    }, 200);

                } else {
                    setTimeout(waitForProcessToFinish, 100);
                }
            }
            waitForProcessToFinish();

        },

        showCreateErrorMsg: function (error) {
            let text = error["msg"];
            console.log("Got error");
            console.log(error);
            MicroModal.close('loading-modal');
            MicroModal.show('error-modal');
            u.byId("create-game-error-text").innerText = text;
        },

        initSearch: function () {
            const input = u.byId("search-place-input");
            const options = {
                fields: ["address_components", "name", "geometry"],
                types: ["(cities)"],
            };
            const autocomplete = new google.maps.places.Autocomplete(input, options);
            autocomplete.addListener("place_changed", function placeSelected(place) {
                let selectedCity = autocomplete.getPlace();
                let name = selectedCity.name;
                let addrComponents = selectedCity.address_components;
                let latLng = selectedCity["geometry"]["location"].toJSON();
                let countryCode = "";
                for (let i = 0; i < addrComponents.length; i++) {
                    let addr = addrComponents[i];
                    if (addr["types"].includes("country")) {
                        countryCode = addr["short_name"];
                    }
                }
                GuessrHome.addSpecificCityToList(name, countryCode, latLng);
                input.value = "";

            });
        },

        selectedSpecificCities: [],

        addSpecificCityToList: function (name, country, location) {
            let table = u.byId("specific-cities-table");
            let tr = u.addElement("tr", table, "");
            let td = u.addElement("td", tr, name + ", " + country);
            let td2 = u.addElement("td", tr, "");
            let btn = u.addElement("button", td2, "");
            btn.innerHTML = '<i class="fas fa-trash"></i>';
            GuessrHome.selectedSpecificCities.push({
                "name": name,
                "country": country,
            });
            let idx = GuessrHome.selectedSpecificCities.length - 1;
            btn.onclick = function () {
                GuessrHome.deleteSpecificCity(tr, idx);
            };

            td.setAttribute("country", country);
            td.setAttribute("place", name);
            td.setAttribute("location", JSON.stringify(location));
            td.setAttribute("name", "selected-places-item");
        },

        deleteSpecificCity: function (tr, idx) {
            GuessrHome.selectedSpecificCities.splice(idx, 1);
            u.byId("specific-cities-table").removeChild(tr);
        },

        showEnterRoom: function () {
            MicroModal.show('modal-2');
        },

        showCreateRoom: function () {
            MicroModal.show('modal-1');

            let geoDropDown = u.byId("set-geo-limits");
            let geoMap = u.byId("selection-map");
            let maxPlayerSlider = u.byId("maxplayer");
            let maxPlayerLabel = u.byId("maxplayer-label");
            let roundSlider = u.byId("rounds");
            let roundLabel = u.byId("rounds-label");
            let timeLimitSlider = u.byId("timelimit");
            let timeLimitLabel = u.byId("timelimit-label");
            let secondLimits = ["10s", "20s", "30s", "40s", "50s"];

            let countrySelect = u.byId("country-selector");
            let continentSelect = u.byId("continent-selector");

            let citySearchInput = u.byId("search-place-input");
            let citySearchTable = u.byId("specific-cities-table");

            let updateCountryList = function() {
                let selectedContinent = continentSelect.value;
                countrySelect.innerText = "";
                u.addElement("option", countrySelect, "All Countries").value = "all";
                let cList = [];
                for (let code in countries) {
                    let country = countries[code];
                    if (country["continent"] === selectedContinent) {
                        country["code"] = code;
                        cList.push(country);
                    }
                }

                cList.sort((a,b) => (a.name > b.name) ? 1 : ((b.name > a.name) ? -1 : 0));

                for(let i in cList) {
                    let country = cList[i];
                    u.addElement("option", countrySelect, country.name).value = country.code;
                }
            };
            continentSelect.onchange = updateCountryList;
            updateCountryList();

            geoDropDown.onchange = function () {
                geoMap.style.display = geoDropDown.value === "custom" ? "block" : "none";
                continentSelect.style.display = geoDropDown.value === "list" ? "block" : "none";
                countrySelect.style.display = geoDropDown.value === "list" ? "block" : "none";
                citySearchInput.style.display = geoDropDown.value === "place-search" ? "block" : "none";
                citySearchTable.style.display = geoDropDown.value === "place-search" ? "block" : "none";
            };
            maxPlayerSlider.oninput = function () {
                maxPlayerLabel.innerText = "Player: " + maxPlayerSlider.value;
            };
            roundSlider.oninput = function () {
                roundLabel.innerText = "Rounds: " + roundSlider.value;
            };

            let timeLimitToString = function (value) {
                if (value <= secondLimits.length) {
                    return secondLimits[value - 1];
                }
                return (value - 5) + "m";
            };

            timeLimitSlider.oninput = function () {
                timeLimitLabel.innerText = "Time limit: " + timeLimitToString(timeLimitSlider.value);
            }
        },
    }
}