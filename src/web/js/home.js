let selectedAreas = [];
let selectedPolygon;
let selectionMap;

function showHome() {

    initUtils();

    MicroModal.init();

    byId("create-room-btn").onclick = showCreateRoom;
    byId("enter-room-btn").onclick = showEnterRoom;
    byId("create-and-play-btn").onclick = createRoom;
    byId("enter-btn").onclick = enterRoom;
}

function enterRoom() {
    let id = byId("enter-room-id").value;
    byId("enter-room-id").setCustomValidity("");
    if (id.length === 0) {
        showRoomIDInvalid();
        return;
    }
    doGetRequest("/exists/" + id, function(response) {
        if(response.status !== 200) {
            showRoomIDInvalid();
            return;
        }
        window.location.href = "/game.html?id=" + id;
    }, function(response) {
        showRoomIDInvalid();
    });
}

function showRoomIDInvalid() {
    console.log("invalid");
    byId("enter-room-id").setCustomValidity("No room found with this id.");
    byId("enter-room-id").reportValidity();
}

function createRoom() {

    let timeSliderValue = byId("timelimit").value;
    let timeSeconds = 0;

    if(timeSliderValue < 5) {
        timeSeconds = timeSliderValue * 10;
    } else {
        timeSeconds = (timeSliderValue - 5) * 60;
    }

    const data = {
        "maxRounds": parseInt(byId("rounds").value),
        "timeLimit": timeSeconds,
        "maxPlayers": parseInt(byId("maxplayer").value)
    };

    const select = byId("set-geo-limits");
    switch(select.value) {
        case "unlimited": return createRoomUnlimited(data);
        case "list": return createRoomList(data);
        case "place-search": return createRoomPlaceSearch(data);
        case "custom": return createRoomCustom(data);
    }
}

function createRoomPlaceSearch(data) {
    let placesElements = document.getElementsByName("selected-places-item");
    let places = [];
    for (let i = 0; i < placesElements.length; i++) {
        let placeElement = placesElements[i];
        let place = placeElement.getAttribute("place");
        let country = placeElement.getAttribute("country");
        places.push({
            "name": place,
            "country": country
        })
    }

    data["places"] = places;

    doPostRequest("/rooms?type=places", data, function (response) {
        window.location.href = "/createRoom.html?id=" + response;
    });
}

function createRoomCustom(data) {
    doPostRequest("/rooms?type=custom", data, function (response) {
        window.location.href = "/createRoom.html?id=" + response;
    });
}

function createRoomList(data) {
    var continent = byId("continent-selector").value;
    var country = byId("country-selector").value;
    var city = byId("city-selector").value;

    data["continent"] = continent;
    data["country"] = country;
    data["city"] = city;


    doPostRequest("/rooms?type=list", data, function (response) {
        window.location.href = "/createRoom.html?id=" + response;
    });
}

function createRoomUnlimited(data) {
    doPostRequest("/rooms?type=unlimited", data, function (response) {
        window.location.href = "/createRoom.html?id=" + response;
    });
}

function initSearch() {
    const input = byId("search-place-input");
    const options = {
        fields: ["address_components", "name"],
        types: ["(cities)"],
    };
    const autocomplete = new google.maps.places.Autocomplete(input, options);
    autocomplete.addListener("place_changed", function placeSelected(place) {
        let selectedCity = autocomplete.getPlace();
        let name = selectedCity.name;
        let addrComponents = selectedCity.address_components;
        let countryCode = "";
        for(let i = 0;i < addrComponents.length; i++) {
            let addr = addrComponents[i];
            if (addr["types"].includes("country")) {
                countryCode = addr["short_name"];
            }
        }
        addSpecificCityToList(name, countryCode);
        input.value = "";

    });
}

selectedSpecificCities = [];

function addSpecificCityToList(name, country) {
    let table = byId("specific-cities-table");
    let tr = addElement("tr", table, "");
    let td = addElement("td", tr, name + ", " + country);
    let td2 = addElement("td", tr, "");
    let btn = addElement("button", td2, "");
    btn.innerHTML = '<i class="fas fa-trash"></i>';
    selectedSpecificCities.push({
        "name": name,
        "country": country,
    });
    let idx = selectedSpecificCities.length - 1;
    btn.onclick = function() {
        deleteSpecificCity(tr, idx);
    };

    td.setAttribute("country", country);
    td.setAttribute("place", name);
    td.setAttribute("name", "selected-places-item");
}

function deleteSpecificCity(tr, idx) {
    selectedSpecificCities.splice(idx, 1);
    byId("specific-cities-table").removeChild(tr);
}



function initGoogleMaps() {
    initSearch();
    selectionMap = new google.maps.Map(document.getElementById('selection-map'), {
        center: {lat: 37.869260, lng: -122.254811},
        zoom: 1,
        fullscreenControl: true,
        mapTypeControl: false,
        streetViewControl: false,
    });

    let selectArea = function (polygon) {
        for(let i = 0; i < selectedAreas.length; i++) {
          setPolygonColor(selectedAreas[i], '#000000')
        }
        setPolygonColor(polygon, '#ff0000')
        selectedPolygon = polygon
    };

    let setPolygonColor = function (polygon, color) {
        var options = {
          strokeColor: color,
          strokeOpacity: 0.8,
          strokeWeight: 3,
          fillColor: color,
          fillOpacity: 0.35
        };
        options.paths = polygon.getPaths().getArray()
        polygon.setOptions(options)
    };

    let deleteSelectedPolygon = function () {
        selectedPolygon.setMap(null);
        for(let i = 0;i < selectedAreas.length; i++) {
          if(selectedAreas[i] === selectedPolygon) {
            selectedAreas.splice(i, 1);
            break;
          }
        }
        selectedPolygon = null
    }

    window.addEventListener('keydown', function(e) {
      if (e.keyCode !== 46 && e.keyCode !== 8) {
        return
      }
      if(selectedPolygon === null || selectedPolygon === undefined) {
        return
      }
      deleteSelectedPolygon();
      event.preventDefault();
    });

    this.drawingManager = new google.maps.drawing.DrawingManager({
        drawingMode: google.maps.drawing.OverlayType.POLYGON,
        drawingControl: true,
        drawingControlOptions: {
          position: google.maps.ControlPosition.TOP_CENTER,
          drawingModes: ['polygon']
        },
        markerOptions: {icon: 'https://developers.google.com/maps/documentation/javascript/examples/full/images/beachflag.png'},
    });

    google.maps.event.addListener(this.drawingManager, 'overlaycomplete', function(event) {
      if (event.type === 'polygon') {
        event.overlay.setMap(null);
        if(event.type === 'polygon') {
          selectedAreas.push(new google.maps.Polygon({
            paths: event.overlay.getPath().getArray(),
            strokeColor: '#000000',
            strokeOpacity: 0.8,
            strokeWeight: 3,
            fillColor: '#000000',
            fillOpacity: 0.35
          }));
        }
        selectedAreas[selectedAreas.length - 1].setMap(selectionMap);
        selectedAreas[selectedAreas.length - 1].addListener('click', function(e, p1, p2) {
          selectArea(this)
        });
      }
    });

    this.drawingManager.setMap(selectionMap);
}

function continentSelected() {
    let continentCode = byId("continent-selector").value;
    let countrySelect = byId("country-selector");
    countrySelect.innerHTML = "";
    addElement("option", countrySelect, "All Areas").value = "all";
    doGetRequestJSON("/countries/" + continentCode,
        (resp) => {
            let countries = resp.countries;
            if (countries === null || countries === undefined || countries.length === 0) {
                console.log("countries empty " + continentCode)
                return;
            }
            countries.sort(function (x, y) {
                return y.name - x.name;
            });

            for(let i = 0;i < countries.length; i++) {
                addElement("option", countrySelect, countries[i].name).value = countries[i].code;
            }
        }, (err) => {
            console.log(err);
        }, () => {}
    );
}

function showEnterRoom() {
    MicroModal.show('modal-2');
}

function showCreateRoom() {
        MicroModal.show('modal-1');

        let geoDropDown = byId("set-geo-limits");
        let geoMap = byId("selection-map");
        let maxPlayerSlider = byId("maxplayer");
        let maxPlayerLabel = byId("maxplayer-label");
        let roundSlider = byId("rounds");
        let roundLabel = byId("rounds-label");
        let timeLimitSlider = byId("timelimit");
        let timeLimitLabel = byId("timelimit-label");
        let secondLimits = ["10s", "20s", "30s", "40s", "50s"];

        let continentSelect = byId("continent-selector");
        let countrySelect = byId("country-selector");
        let citySelect = byId("city-selector");

        let citySearchInput = byId("search-place-input");
        let citySearchTable = byId("specific-cities-table");

        continentSelect.onchange = continentSelected;

        geoDropDown.onchange = function() {
            geoMap.style.display = geoDropDown.value === "custom" ? "block" : "none";

            continentSelect.style.display = geoDropDown.value === "list" ? "block" : "none";
            countrySelect.style.display = geoDropDown.value === "list" ? "block" : "none";
            citySelect.style.display = geoDropDown.value === "list" ? "block" : "none";
            citySearchInput.style.display = geoDropDown.value === "place-search" ? "block":"none";
            citySearchTable.style.display = geoDropDown.value === "place-search" ? "block":"none";
        };
        maxPlayerSlider.oninput = function() {
            maxPlayerLabel.innerText = "Player: " + maxPlayerSlider.value;
        };
        roundSlider.oninput = function() {
            roundLabel.innerText = "Rounds: " + roundSlider.value;
        };

        var timeLimitToString = function(value) {
            if(value <= secondLimits.length) {
                return secondLimits[value - 1];
            }
            return (value - 5) + "m";
        };

        timeLimitSlider.oninput = function() {
            timeLimitLabel.innerText = "Time limit: " + timeLimitToString(timeLimitSlider.value);
        }
}