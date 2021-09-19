let API_ENDPOINT = "";
let API_KEY = "";

function initUtils(){
    API_ENDPOINT = document.getElementById("API_ENDPOINT").value;
    API_KEY = document.getElementById("API_KEY").value;
}

function getRenderedTemplate(templName, params) {
    var template = document.getElementById(templName).innerHTML;
    return Mustache.render(template, params);
}

function byId(id) {
    return document.getElementById(id);
}

function addElement(tag, parent, innerText) {
    let element = document.createElement(tag);
    parent.appendChild(element);
    if(innerText !== undefined && innerText !== null) {
        element.innerText = innerText;
    }
    return element;
}

function distanceToText(distanceInMeters) {
    if (distanceInMeters < 1000) {
        return distanceInMeters + "m";
    } else if (distanceInMeters < 100000) {
        let km = distanceInMeters / 1000;
        let mets = distanceInMeters % 1000;
        return ~~km + "." + (("" + mets).substring(0, 1)) + "km";
    } else {
        let km = distanceInMeters / 1000;
        return  ~~km + "km";
    }
}

function getRequestParameter(parameterName) {
    var result = null, tmp = [];
    var items = location.search.substr(1).split("&");
    for (var index = 0; index < items.length; index++) {
        tmp = items[index].split("=");
        if (tmp[0] === parameterName) result = decodeURIComponent(tmp[1]);
    }
    return result;
}

function doGetRequest(requestPath, onload, onerror) {
    fetch(API_ENDPOINT + requestPath, {
        headers: {
            'x-api-key': API_KEY,
        },
    })
        .then(onload)
        .catch(onerror);
}

function doGetRequestJSON(requestPath, onload, onerror, always) {
    fetch(API_ENDPOINT + requestPath, {
        headers: {
            'x-api-key': API_KEY,
        },
    })
        .then((resp) => {
            if (resp.status !== 200) {
                throw new Error("Got status " + resp.status + " on " + requestPath)
            }
            return resp;
        })
        .then((resp) => resp.json())
        .then(onload)
        .catch(onerror)
        .finally(always);
}

function doPostRequest(requestPath, data, callback, errorCallback, expectedStatus) {
    fetch(API_ENDPOINT + requestPath, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': API_KEY,
        },
        body: JSON.stringify(data),
    })
        .then((resp) => {
            if(resp.status !== expectedStatus) {
                resp.json().then(errorCallback)
            }else {
                resp.json().then(callback)
            }
        })
        .catch((error) => {
            console.log('Error:', error);
            errorCallback("Connection error " + error)
        });
}
export default {
    initUtils,
    getRenderedTemplate,
    byId,
    addElement,
    distanceToText,
    getRequestParameter,
    doGetRequest,
    doGetRequestJSON,
    doPostRequest
}