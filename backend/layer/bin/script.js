var page = require('webpage').create(),
    system = require('system'),
    address;
address = system.args[1];
page.open(address, function (status) {
    var processedData = null;
    function getData() {
        return page.evaluate(function() {
            var ele = document.getElementById("streetview-data");
            var status = ele.getAttribute("status");
            if(status === "ready") {
                var panos = ele.innerText.substring(1);
                return '{"panos": [' + panos + ']}'
            }
            return "";
        });
    }
    function waitForProcessToFinish() {
        processedData = getData();
        if (processedData === "") {
            setTimeout(waitForProcessToFinish, 100);
        } else {
            console.log(processedData);
            phantom.exit();
        }
    }
    waitForProcessToFinish();
});