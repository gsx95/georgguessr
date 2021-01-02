const MAP_VIEW_NAME = "MAP";

class MapView {

    constructor(gameID) {
        this.gameID = gameID;
    }

    show() {
        CURRENT_VIEW = MAP_VIEW_NAME;
        renderWholeView("MapTemplate");
        this.initMap();
    }

    initMap() {
        const fenway = { lat: 42.345573, lng: -71.098326 };
        const map = new google.maps.Map(document.getElementById("map"), {
            center: fenway,
            zoom: 14,
        });
        const panorama = new google.maps.StreetViewPanorama(
            document.getElementById("map"),
            {
                position: fenway,
                pov: {
                    heading: 34,
                    pitch: 10,
                }
            }
        );
        panorama.setOptions({
            addressControl: false,
            fullscreenControl: false,
            motionTracking: false,
            motionTrackingControl: false,
            showRoadLabels: false,
            panControl: false,
            zoomControl: false,
            enableCloseButton: false,
        })
        map.setStreetView(panorama);
    }
}