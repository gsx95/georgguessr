let lastRequestTime = 0;

function showHome() {
    lastRequestTime = 0;

    MicroModal.init();

    updateRoomTable();

    byId("create-room-btn").onclick = showCreateRoom;
}

function updateRoomTable() {
    doGetRequestJSON("/available-rooms",
        (rooms) => {
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

function showCreateRoom() {
        MicroModal.show('modal-1');

        let pwdDropDown = byId("set-pwd");
        let pwdField = byId("pwd");
        let maxPlayerSlider = byId("maxplayer");
        let maxPlayerLabel = byId("maxplayer-label");
        let roundSlider = byId("rounds");
        let roundLabel = byId("rounds-label");
        let timeLimitSlider = byId("timelimit");
        let timeLimitLabel = byId("timelimit-label");
        let secondLimits = ["10s", "20s", "30s", "40s", "50s"];

        pwdDropDown.onchange = function() {
            pwdField.style.visibility = pwdDropDown.value == "protected" ? "visible" : "hidden";
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