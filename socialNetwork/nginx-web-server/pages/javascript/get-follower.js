function showFollowers() {

    while (document.getElementsByClassName("post-text").firstChild) {
        document.getElementsByClassName("post-text").removeChild(document.getElementsByClassName("post-text").firstChild);
    }
    username = document.getElementById("username");
    const Http = new XMLHttpRequest();
    const url = 'http://' + window.location.hostname + ':8080/api/user/get_follower';
    Http.open("GET", url, true);
    Http.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            let resp_json = JSON.parse(Http.responseText);

            follower_div = document.getElementsByClassName("follower-div");
            follower_id = document.getElementsByClassName("follower-id");
            for (let i = 0; i < resp_json.length; i++) {
                console.log("i : " + i);
                console.log("follower_div.length - 1: " + (follower_div.length - 1))
                if (i == follower_div.length - 1) {
                    let itm = follower_div[i];
                    let cln = itm.cloneNode(true);
                    document.getElementById("follower-list").appendChild(cln);
                }
                let follower = resp_json[i];
                follower_div[i].style.display = "block";
                follower_id[i].innerText = follower["follower_id"];
            }
        }
    };
    Http.send(null);

}

showFollowers();