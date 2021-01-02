let mainDiv;

function initRenderer() {
    mainDiv = document.getElementById("main");
}

function renderWholeView(templName, params) {
    var template = document.getElementById(templName).innerHTML;
    mainDiv.innerHTML = Mustache.render(template, params);
}

function getRenderedTemplate(templName, params) {
    var template = document.getElementById(templName).innerHTML;
    return Mustache.render(template, params);
}
