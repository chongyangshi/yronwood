const API_BASE = "http://127.0.0.1:18080"
const ACCESS_TYPE_PUBLIC = "public"
const ACCESS_TYPE_PRIVATE = "private"

var state_authenticated = false

function set_authenticated() {
    $("#authenticateIcon").removeClass("oi-lock-locked");
    $("#authenticateIcon").addClass("oi-lock-unlocked");
    $("#uploadIcon").removeAttr("hidden");
    console.log("Secret set")
}

function set_basic_auth_secret(secret) {
    sessionStorage.setItem("yronwood_basic_auth_secret", secret);
}

function get_basic_auth_secret() {
    return sessionStorage.getItem("yronwood_basic_auth_secret");
}

function list_images(access_type) {
    $.ajax({
        url: API_BASE + "/list",
        type: "POST",
        data: JSON.stringify({
            "access_type": access_type,
            "auth": {
                "secret": get_basic_auth_secret(),
            }
        }),
        success: function(result){
            if (result.images !== null) {
                var row_size = 0;
                var current_row = null;
                for (let image of result.images) {
                    if (image.file_name == null || image.file_name === "") {
                        continue
                    }
                    if (row_size % 3 == 0) {
                        current_row = $("<div class='row padded-row'></div>");
                        $("#images-container").append(current_row);
                        row_size = 0;
                    }
                    var image_link = API_BASE + "/uploads/" + encodeURIComponent(access_type) + "/" + encodeURIComponent(image.file_name)
                    var secret = get_basic_auth_secret();
                    if (secret != null && secret != "") {
                        image_link += "?secret=" + encodeURIComponent(secret)
                    }
                    $(current_row).append("<div class='col-sm grid-image'><a target='_blank' href='"+ image_link + "'><img class='grid-image' src='" + image_link +"' /></a></div>");
                    row_size++;
                }
            }
        },
        error: function(result){
            if (result.responseText == undefined || result.responseText == "") {
                $("#yronwood-error").text("Error: unknown (" + result.statusText +")");
            } else {
                var err = $.parseJSON(xhr.responseText)
                $("#yronwood-error").text("Error: " + err.message + " (" + err.code + ")");
            }
        }
    });
}

$( document ).ready(function() {
    list_images(ACCESS_TYPE_PUBLIC);
});