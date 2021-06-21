var freq_con_btn = document.querySelectorAll(".freq-c");
var unfollow_btn = document.querySelectorAll(".unfollow");


for (var i = 0; i < freq_con_btn.length; i++) {
    freq_con_btn[i].addEventListener("click", function () {
        this.classList.toggle("freq_clicked");
        if ($(this).hasClass("freq_clicked")) {
            console.log("I");
            $(this).text("Remove from Frequent Contact");
        } else {
            $(this).text("Set Frequent Contact");
        }
    });
}


for (var i = 0; i < unfollow_btn.length; i++) {
    unfollow_btn[i].addEventListener("click", function () {
        this.classList.toggle("unfollow_clicked");
        if ($(this).hasClass("unfollow_clicked")) {
            console.log("I");
            $(this).text("Follow");
        } else {
            $(this).text("Unfollow");
        }
    });
}

function followUser() {
    follower = localStorage.getItem("username");
}

function uploadPost(media_json) {
    if (document.getElementById('post-content').value !== "") {
        const Http = new XMLHttpRequest();
        const url = 'http://' + window.location.hostname + ':8080/api/post/compose';
        Http.open("POST", url, true);
        var body = "post_type=0&text=" + document.getElementById('post-content').value
        Http.onreadystatechange = function () {
            if (this.readyState == 4 && this.status == 200) {
                console.log(Http.responseText);
            }
        };
        if (media_json === undefined) {
            Http.send(body);
        } else {
            body += "&media_ids=[\"" + media_json.media_id + "\"]&media_types=[\"" + media_json.media_type + "\"]"
            Http.send(body);
        }
        window.location.reload();
    }
}
