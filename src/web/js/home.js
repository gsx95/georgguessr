let lastRequestTime = 0;

function showHome() {
    lastRequestTime = 0;
    updateRoomTable();

    document.getElementById("create-room-btn").onclick = function () {
        showMapView("not-implemented");
    }
}

function updateRoomTable() {
    doGetRequestJSON("/available-rooms",
        (rooms) => {
            rooms.sort(function (x, y) {
                return y.created - x.created;
            });
            let rows = getRenderedTemplate("HomeTableRowTemplate", {"rooms": rooms});
            let roomsTableBody = document.getElementById("home-table-body");
            roomsTableBody.innerHTML = rows;
        }, (err) => {
            console.log(err);
        }, () => {
            setTimeout(updateRoomTable, 1000);
        }
    );
    lastRequestTime = new Date(new Date().toISOString()).getTime();
}