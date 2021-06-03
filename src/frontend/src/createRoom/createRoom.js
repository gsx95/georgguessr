import u from '../utils.js';

export default {

    GuessrRoomCreation: {

        searchRad: [50, 100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000],
        svService: null,

        processRoom: function () {
            u.initUtils();
            GuessrRoomCreation.svService = new google.maps.StreetViewService();

            let gameID = u.getRequestParameter("id");
            let processedCount = 0;

            u.doGetRequestJSON("/game/stats/" + gameID, function (resp) {
                console.log(resp);
                let rounds = resp["gameRounds"];
                console.log(rounds);
                for (let i = 1; i <= rounds.length; i++) {
                    u.doGetRequestJSON("/game/pos/" + gameID + "/" + i, function (resp) {
                        console.log(resp);
                        GuessrRoomCreation.getStreetViewForPos(resp.lat, resp.lon, 0, function (panoId) {
                            console.log("got pano id");
                            u.doPostRequestString("/game/pano/" + gameID + "/" + i, panoId, function (resp) {
                                processedCount++;
                                console.log("done")
                            })
                        });
                    }, function (err) {
                        console.log(err);
                    });
                }

                function waitForProcessToFinish() {
                    if (processedCount >= rounds.length) {
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

        getStreetViewForPos: function (lat, lon, count, callback) {
            GuessrRoomCreation.svService.getPanorama({
                location: {"lat": lat, "lng": lon},
                "radius": GuessrRoomCreation.searchRad[count]
            }, function (data, status) {
                if (status !== "OK") {
                    if (count + 1 === GuessrRoomCreation.searchRad.length) {
                        return null;
                    }
                    console.log(status);
                    console.log(count);
                    return GuessrRoomCreation.getStreetViewForPos(lat, lon, count + 1, callback)
                }
                const location = data.location;
                callback(location.pano);
            });
        },
    }
}