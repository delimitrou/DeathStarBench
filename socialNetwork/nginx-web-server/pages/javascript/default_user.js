const generate_default_user = () => {

    let url = "api/user/register";
    let url2 = "api/post/compose";
    let user1 = { first_name: "Mark", last_name: "Zuckerberg", username: "mark", password: "123" };
    let user2 = { first_name: "Elon", last_name: "Mask", username: "elon", password: "123" };
    let user3 = { first_name: "Bill", last_name: "Gates", username: "bill", password: "123" };
    $.ajax({
        url: url,
        type: "POST",
        data: user1,
        contentType: "application/x-www-form-urlencoded",
        success: function (result, status, xhr) {
            console.log(result);
        },
        error: function (xhr, status, error) {
            console.log(error);
        }
    }
    );

    $.ajax({
        url: url,
        type: "POST",
        data: user2,
        contentType: "application/x-www-form-urlencoded",
        success: function (result, status, xhr) {
            console.log(result);
        },
        error: function (xhr, status, error) {
            console.log(error);
        }
    }
    );

    $.ajax({
        url: url,
        type: "POST",
        data: user3,
        contentType: "application/x-www-form-urlencoded",
        success: function (result, status, xhr) {
            console.log(result);
        },
        error: function (xhr, status, error) {
            console.log(error);
        }
    }
    );
}
console.log("generate default user...");
generate_default_user();