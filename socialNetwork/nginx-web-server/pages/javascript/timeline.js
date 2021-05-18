function showTimeline(type) {
    const start = 0
    const stop = 100

    while (document.getElementsByClassName("post-text").firstChild) {
        document.getElementsByClassName("post-text").removeChild(document.getElementsByClassName("post-text").firstChild);
    }
    if (start !== "" && stop !== "") {
        var params = "start=" + start + "&stop=" + stop;
        const Http = new XMLHttpRequest();
        // const url = 'http://' + window.location.hostname + ':8080/api/user-timeline/read';
        const url = 'http://' + window.location.hostname + ':8080/api/' + type + '/read';
        Http.open("GET", url + "?" + params, true);
        Http.onreadystatechange = function () {
            if (this.readyState == 2 && this.status == 401) {
                console.log("unauthorized user login")
                window.location.href = 'http://' + window.location.hostname + ":8080/index.html";
                localStorage.clear();
            }
            else if (this.readyState == 4 && this.status == 200) {
                var getFromUrl = new URL(location.href);
                var curUser = getFromUrl.searchParams.get("username");
                if (localStorage.getItem("username") == null) {
                    localStorage.setItem("username", curUser)
                    followUsername()
                }
                console.log(document.getElementById("show-post"));
                if (type === "user-timeline") {
                    document.getElementById("mentioned_user").innerText = localStorage.getItem("username");
                }
                var resp_json = JSON.parse(Http.responseText);
                post_cards = document.getElementsByClassName("post-card");
                post_texts = document.getElementsByClassName("post-text");
                post_times = document.getElementsByClassName("post-time");
                post_creators = document.getElementsByClassName("post-creator");
                post_images = document.getElementsByClassName("post-img");
                post_footer = document.getElementsByClassName("post-footer");
                showUsername();
                validPost = 0;
                for (var i = 0; i < resp_json.length; i++) {
                    if (i == post_cards.length - 1) {
                        var itm = post_cards[i];
                        var cln = itm.cloneNode(true); //clone the post_card[i]
                        document.getElementById("card-block").appendChild(cln);
                    }
                    var post_json = resp_json[i];
                    var media_json = post_json["media"];
                    post_cards[i].style.display = "block";
                    post_texts[i].innerHTML = replaceMentionWithHTMLLinks(post_json["text"]);

                    post_times[i].innerText = getTime(post_json["timestamp"]);
                    post_creators[i].innerText = post_json["creator"]["username"];
                    for (var j = 0; j < media_json.length; j++) {
                        post_images[i].src = "http://" + window.location.hostname + ":8081/get-media/?filename=" +
                            media_json[j]["media_id"] + "." +
                            media_json[j]["media_type"];
                    }
                    validPost += 1;
                }
                if (validPost === 0) {
                    var itm = post_cards[0];
                    var cln = itm.cloneNode(true); //clone the post_card[i]
                    document.getElementById("card-block").appendChild(cln);
                    post_cards[0].style.display = "block";
                    post_texts[0].style.fontSize = "x-large"
                    post_footer[0].style.display = "none"

                }

            }
        };

        Http.send(null);
    }
}

function show_Mentioned_User_Timeline(mentioned_user) {
    const start = 0
    const stop = 100

    while (document.getElementsByClassName("post-text").firstChild) {
        document.getElementsByClassName("post-text").removeChild(document.getElementsByClassName("post-text").firstChild);
    }
    if (start !== "" && stop !== "") {
        var params = "start=" + start + "&stop=" + stop;
        const Http = new XMLHttpRequest();
        const url = 'http://' + window.location.hostname + ':8080/api/home-timeline/read';
        Http.open("GET", url + "?" + params, true);
        Http.onreadystatechange = function () {
            if (this.readyState == 2 && this.status == 401) {
                console.log("unauthorized user login")
                window.location.href = 'http://' + window.location.hostname + ":8080/index.html";
                localStorage.clear();
            }
            else if (this.readyState == 4 && this.status == 200) {
                var resp_json = JSON.parse(Http.responseText);
                post_cards = document.getElementsByClassName("post-card");
                post_texts = document.getElementsByClassName("post-text");
                post_times = document.getElementsByClassName("post-time");
                post_creators = document.getElementsByClassName("post-creator");
                post_images = document.getElementsByClassName("post-img");
                post_footer = document.getElementsByClassName("post-footer");
                document.getElementById("mentioned_user").innerText = mentioned_user;
                showUsername();
                validPost = 0;
                for (var i = 0; i < resp_json.length; i++) {
                    if (i == post_cards.length - 1) {
                        var itm = post_cards[i];
                        var cln = itm.cloneNode(true); //clone the post_card[i]
                        document.getElementById("card-block").appendChild(cln);
                    }
                    var post_json = resp_json[i];
                    if (post_json["creator"]["username"].localeCompare(mentioned_user) != 0) {
                        continue;
                    }
                    var media_json = post_json["media"];
                    post_cards[i].style.display = "block";
                    post_texts[i].innerHTML = replaceMentionWithHTMLLinks(post_json["text"]);
                    //console.log(post_json["time"]);
                    post_times[i].innerText = getTime(post_json["timestamp"]);
                    post_creators[i].innerText = post_json["creator"]["username"];
                    for (var j = 0; j < media_json.length; j++) {
                        post_images[i].src = "http://" + window.location.hostname + ":8081/get-media/?filename=" +
                            media_json[j]["media_id"] + "." +
                            media_json[j]["media_type"];
                    }
                    validPost += 1;
                }
                if (validPost === 0) {
                    var itm = post_cards[0];
                    var cln = itm.cloneNode(true); //clone the post_card[i]
                    document.getElementById("card-block").appendChild(cln);
                    post_cards[0].style.display = "block";
                    post_texts[0].innerHTML = "The user has not post since you followed!";
                    post_texts[0].style.fontSize = "x-large"
                    // post_times[0].innerText = ""
                    // post_creators[0].innerText = "Welcome to DeathStar!"
                    post_footer[0].style.display = "none"

                }
                // console.log(resp_json);
            }
        };
        Http.send(null);

    }
}


function showUsername() {
    if (localStorage.getItem("username") != undefined && localStorage.getItem("username") != null) {
        var username = localStorage.getItem("username");
    }
    document.getElementById("username").innerText = username;
}

function getTime(time) {
    var new_time = Number(time);
    var a = new Date(new_time);
    var s = a.toDateString();
    var hour = ("0" + a.getHours()).substr(-2);
    var min = ("0" + a.getMinutes()).substr(-2);
    var time = hour + ':' + min + ' ' + s;
    return time;
}
function replaceMentionWithHTMLLinks(text) {
    return text.replace(/(^|\s)@(\w+)/g, '$1<a href="profile.html?username=$2">@$2</a>');

}

function get_follower() {
    username = document.getElementById("username");
    const Http = new XMLHttpRequest();
    const url = 'http://' + window.location.hostname + ':8080/api/user/get_follower';
    Http.open("GET", url, true);
    Http.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            var resp_json = JSON.parse(Http.responseText);
            console.log(resp_json);
        }
    }
    Http.send(null);
}

function logout() {
    localStorage.clear()
    document.cookie = "login_token=;"
    window.location.reload()
}