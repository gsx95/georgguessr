let GuessrRoomCreation = {

    searchRad: [50, 100, 500, 1000, 5000, 10000, 50000, 100000],
    svService: null,

    processRoom: function() {
        initUtils();
        GuessrRoomCreation.svService = new google.maps.StreetViewService();

        let gameID = getRequestParameter("id");
        let processedCount = 0;

        doGetRequestJSON("/game/stats/" + gameID, function (resp) {
            let rounds = resp.rounds;
            for (let i = 1; i <= rounds; i++) {
                doGetRequestJSON("/game/pos/" + gameID + "/" + i, function (resp) {
                    console.log(resp);
                    GuessrRoomCreation.getStreetViewForPos(resp.lat, resp.lon, 0, function (panoId) {
                        doPostRequestString("/game/pano/" + gameID + "/" + i, panoId, function (resp) {
                            processedCount++;
                        })
                    });
                }, function (err) {
                    console.log(err);
                });
            }

            function waitForProcessToFinish() {
                if (processedCount >= rounds) {
                    window.location.href = "/game?id=" + gameID;
                } else {
                    setTimeout(waitForProcessToFinish, 100);
                }
            }

            waitForProcessToFinish();

        }, function (err) {
            console.log(err);
        });

    },

    getStreetViewForPos: function(lat, lon, count, callback) {
        GuessrRoomCreation.svService.getPanorama({location: {"lat": lat, "lng": lon}, "radius": GuessrRoomCreation.searchRad[count]}, function (data, status) {
            if (status !== "OK") {
                if (count + 1 === GuessrRoomCreation.searchRad.length) {
                    return null;
                }
                return GuessrRoomCreation.getStreetViewForPos(lat, lon, count + 1, callback)
            }
            const location = data.location;
            callback(location.pano);
        });
    },
};