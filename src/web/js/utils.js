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
    var element = document.createElement(tag);
    parent.appendChild(element);
    element.innerText = innerText;
    return element;
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

function doGetRequestJSON(requestPath, onload, onerror, always) {
    fetch(API_ENDPOINT + requestPath, {
        headers: {
            'x-api-key': API_KEY,
        },
    })
        .then((resp) => {
            if (resp.status !== 200) {
                throw new Error("Got status " + resp.status + " on " + url)
            }
            return resp;
        })
        .then((resp) => resp.json())
        .then(onload)
        .catch(onerror)
        .finally(always);
}

function doPostRequestString(requestPath, data, callback) {
    console.log("body: " + data);
    fetch(API_ENDPOINT + requestPath, {
        method: 'POST',
        headers: {
            'x-api-key': API_KEY,
        },
        body: data,
    })
        .then((resp) => resp.text())
        .then(callback)
        .catch((error) => {
            console.error('Error:', error);
        });
}

function doPostRequest(requestPath, data, callback) {
    fetch(API_ENDPOINT + requestPath, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': API_KEY,
        },
        body: JSON.stringify(data),
    })
        .then((resp) => resp.text())
        .then(callback)
        .catch((error) => {
            console.error('Error:', error);
        });
}