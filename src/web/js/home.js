const ROOMS_ENDPOINT = "http://localhost:3000/rooms?lastRequest=";
const HOME_VIEW_NAME = "HOME";


class HomeView {

    constructor() {
        if (HomeView._instance) {
            return HomeView._instance
        }
        this.lastRequestTime = 0;
        HomeView._instance = this;
    }

    show() {
        CURRENT_VIEW = HOME_VIEW_NAME;
        renderWholeView("HomeTemplate");
        this.updateRoomTable();

        document.getElementById("create-room-btn").onclick = function (){
            showMapView("not-implemented");
        }
    }

    updateRoomTable() {
        doGetRequestJSON(ROOMS_ENDPOINT + this.lastRequestTime,
            (rooms) => {
                rooms.sort(function(x, y){
                    return y.created - x.created;
                });
                let rows = getRenderedTemplate("HomeTableRowTemplate", {"rooms": rooms});
                let roomsTableBody = document.getElementById("home-table-body");
                roomsTableBody.innerHTML = rows;
            }, (err) => {
                console.log(err);
            }, () => {
                if(CURRENT_VIEW === HOME_VIEW_NAME) {
                    setTimeout(this.updateRoomTable, 1000);
                }
            }
        );
        this.lastRequestTime = new Date(new Date().toISOString()).getTime();
    }
}