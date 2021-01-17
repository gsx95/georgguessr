let lastRequestTime = 0;
let selectedAreas = [];
let selectedPolygon;
let selectionMap;

function showHome() {
    lastRequestTime = 0;

    MicroModal.init();
    updateRoomTable();

    byId("create-room-btn").onclick = showCreateRoom;
}

function updateRoomTable() {
    doGetRequestJSON("/available-rooms",
        (rooms) => {
            if (rooms === null || rooms === undefined) {
                return;
            }
            rooms.sort(function (x, y) {
                return y.created - x.created;
            });
            let rows = getRenderedTemplate("HomeTableRowTemplate", {"rooms": rooms});
            let roomsTableBody = byId("home-table-body");
            roomsTableBody.innerHTML = rows;
        }, (err) => {
            console.log(err);
        }, () => {
            setTimeout(updateRoomTable, 1000);
        }
    );
    lastRequestTime = new Date(new Date().toISOString()).getTime();
}

function initMap() {
    selectionMap = new google.maps.Map(document.getElementById('selection-map'), {
        center: {lat: 37.869260, lng: -122.254811},
        zoom: 1,
        fullscreenControl: true,
        mapTypeControl: false,
        streetViewControl: false,
    })

    let selectArea = function (polygon) {
        for(let i = 0; i < selectedAreas.length; i++) {
          setPolygonColor(selectedAreas[i], '#000000')
        }
        setPolygonColor(polygon, '#ff0000')
        selectedPolygon = polygon
    }

    let setPolygonColor = function (polygon, color) {
        var options = {
          strokeColor: color,
          strokeOpacity: 0.8,
          strokeWeight: 3,
          fillColor: color,
          fillOpacity: 0.35
        }
        options.paths = polygon.getPaths().getArray()
        polygon.setOptions(options)
    }

    let deleteSelectedPolygon = function () {
        selectedPolygon.setMap(null)
        for(let i = 0;i < selectedAreas.length; i++) {
          if(selectedAreas[i] == selectedPolygon) {
            selectedAreas.splice(i, 1);
            break;
          }
        }
        selectedPolygon = null
    }

    window.addEventListener('keydown', function(e) {
      console.log(e.keyCode);
      if (e.keyCode != 46 && e.keyCode != 8) {
        return
      }
      if(selectedPolygon == null || selectedPolygon == undefined) {
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
      if (event.type == 'polygon') {
        event.overlay.setMap(null);
        if(event.type == 'polygon') {
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
        console.log(selectedAreas.length);
      }
    });

    this.drawingManager.setMap(selectionMap);
}

function continentSelected() {
    let continentCode = byId("continent-selector").value;
    let countrySelect = byId("country-selector");
    countrySelect.innerHTML = "";
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

function countrySelected() {

}

function citySelected() {

}

function showCreateRoom() {
        MicroModal.show('modal-1');

        let pwdDropDown = byId("set-pwd");
        let pwdField = byId("pwd");
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

        continentSelect.onchange = continentSelected;
        countrySelect.onchange = countrySelected;
        citySelect.onchange = citySelected;

        pwdDropDown.onchange = function() {
            pwdField.style.visibility = pwdDropDown.value == "protected" ? "visible" : "hidden";
            pwdField.style.marginBottom = pwdDropDown.value == "protected" ? "30px" : "0px";
            pwdField.focus();
        }
        geoDropDown.onchange = function() {
            geoMap.style.display = geoDropDown.value == "custom" ? "block" : "none";

            continentSelect.style.display = geoDropDown.value == "list" ? "block" : "none";
            countrySelect.style.display = geoDropDown.value == "list" ? "block" : "none";
            citySelect.style.display = geoDropDown.value == "list" ? "block" : "none";
        }
        maxPlayerSlider.oninput = function() {
            maxPlayerLabel.innerText = "Player: " + maxPlayerSlider.value;
        }
        roundSlider.oninput = function() {
            roundLabel.innerText = "Rounds: " + roundSlider.value;
        }

        var timeLimitToString = function(value) {
            if(value <= secondLimits.length) {
                return secondLimits[value - 1];
            }
            return (value - 5) + "m";
        }

        timeLimitSlider.oninput = function() {
            timeLimitLabel.innerText = "Time limit: " + timeLimitToString(timeLimitSlider.value);
        }
}