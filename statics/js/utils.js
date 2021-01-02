let CURRENT_VIEW = HOME_VIEW_NAME;

function doGetRequestJSON(url, onload, onerror, always) {
    fetch(url)
        .then((resp) => {
            if(resp.status !== 200) {
                throw Error("Got status " + resp.status + " on " + url)
            }
            return resp;
        })
        .then((resp) => resp.json())
        .then(onload)
        .catch(onerror)
        .finally(function() {
            if(always !== undefined) {
                always();
            }
        });
}