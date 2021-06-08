function showfollowees() {

    while (document.getElementsByClassName("post-text").firstChild) {
        document.getElementsByClassName("post-text").removeChild(document.getElementsByClassName("post-text").firstChild);
    }
    username = document.getElementById("username");
    const Http = new XMLHttpRequest();
    const url = 'http://' + window.location.hostname + ':8080/api/user/get_followee';
    Http.open("GET", url, true);
    Http.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            let resp_json = JSON.parse(Http.responseText);

            followee_div = document.getElementsByClassName("followee-div");
            followee_id = document.getElementsByClassName("followee-id");
            console.log(followee_id)
            for (let i = 0; i < resp_json.length; i++) {
                console.log("i : " + i);
                console.log("followee_div.length - 1: " + (followee_div.length - 1))
                if (i == followee_div.length - 1) {
                    let itm = followee_div[i];
                    let cln = itm.cloneNode(true);
                    document.getElementById("followee-list").appendChild(cln);
                }
                let followee = resp_json[i];
                followee_div[i].style.display = "block";
                followee_id[i].innerText = followee["followee_id"];
            }
        }
    };
    Http.send(null);

}

showfollowees();