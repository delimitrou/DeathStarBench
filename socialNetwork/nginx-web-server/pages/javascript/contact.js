let followUsername = () => {
    let username = localStorage.getItem("username");
    document.querySelectorAll(".follow-username").forEach(function (element) {
        element.setAttribute("value", username);
    })
}
function showUsername() {
    if (localStorage.getItem("username") != undefined && localStorage.getItem("username") != null) {
        var username = localStorage.getItem("username");
    }
    document.getElementById("username").textContent = username;
    console.log(username);
}

followUsername();
