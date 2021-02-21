let guessMap;
let gameID;
let marker;

function showGame() {
    setGuessMapResizable();

}

function initGameMaps() {
    initUtils();
    gameID = getRequestParameter("id");
    console.log(gameID);

    guessMap = new google.maps.Map(byId("guess-map"), {
        center: {lat: 37.869260, lng: -122.254811},
        zoom: 1,
        fullscreenControl: true,
        mapTypeControl: false,
        streetViewControl: false,
    });

    doGetRequestJSON("/game/stats/" + gameID, function (resp) {
        let rounds = resp.rounds;
        for (let i = 1; i <= rounds; i++) {
            doGetRequestJSON("/game/pos/" + gameID + "/" + i, function (resp) {
                //drawMarker(resp);
            }, function (err) {
                console.log(err);
            })
        }
    }, function (err) {
        console.log(err);
    });


    const fenway = { lat: 42.345573, lng: -71.098326 };
    const panorama = new google.maps.StreetViewPanorama(
        document.getElementById("pano"),
        {
            position: fenway,
            addressControl: false,
            fullscreenControl: false,
            showRoadLabels: false,
            zoomControl: true,
            panControl: false,
        }
    );

    guessMap.addListener("click", (data) => {
        drawMarker(data.latLng);
        enableGuessButton();
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