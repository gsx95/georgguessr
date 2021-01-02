function start() {
    initRenderer();
    showHomeView();

}

function showHomeView() {
    let homeView = new HomeView();
    homeView.show();
}

function showMapView(gameID) {
    let mapView = new MapView(gameID);
    mapView.show();
}