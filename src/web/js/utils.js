let API_ENDPOINT = "";
let API_KEY = "";

function doGetRequestJSON(requestPath, onload, onerror, always) {
    fetch(API_ENDPOINT + requestPath, {
        headers: {
            'x-api-key': API_KEY,
        },
    })
        .then((resp) => {
            if(resp.status !== 200) {
                throw new Error("Got status " + resp.status + " on " + url)
            }
            return resp;
        })
        .then((resp) => resp.json())
        .then(onload)
        .catch(onerror)
        .finally(always);
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