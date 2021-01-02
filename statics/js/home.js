const ROOMS_ENDPOINT = "http://localhost:3000/rooms?lastRequest=";
const HOME_VIEW_NAME = "HOME";


let lastRequestTime;


function showHome() {
    CURRENT_VIEW = HOME_VIEW_NAME;
    initVars();
    renderWholeView("HomeTemplate");
    updateRoomTable();
}

function initVars() {
    lastRequestTime = 0;
}

function updateRoomTable() {
    doGetRequestJSON(ROOMS_ENDPOINT + lastRequestTime,
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
                setTimeout(updateRoomTable, 1000);
            }
        }
    );
    lastRequestTime = new Date(new Date().toISOString()).getTime();
}