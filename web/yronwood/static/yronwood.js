const ACCESS_TYPE_PUBLIC = "public"
const ACCESS_TYPE_PRIVATE = "private"

// Change to the base address of your Yronwood backend server, whose CORS policy must
// allow the origin of the web frontend if using different hostnames.
var API_BASE = "https://i.doge.at"
if (window.location.hostname == undefined || window.location.hostname == "") {
    // For local running, served by browser from file.
    API_BASE = "http://127.0.0.1:18080";
}

// Change text to reflect license choice for public images.
var LICENSE_TEXT = "Unless otherwise stated, all public contents of this gallery are original works © Chongyang Shi, and are licensed for re-use under the terms of CC BY 4.0.";

var current_accss_type = ACCESS_TYPE_PUBLIC
var current_page = 1

$(document).ready(function () {
    var token = get_basic_auth_token();
    if (token !== "") {
        set_authenticated();
    }
    list_images(current_accss_type, current_page);

    $(".license-text").text(LICENSE_TEXT);
});

function list_images(access_type, page) {
    const tags = getTagsFromCurrentURL();
    if (tags.length !== 0) {
        $("#yronwood-tags-info-text").text("Showing only images tagged with: ".concat(tags.join(", ")).concat("."));
        $("#yronwood-tags-info").show();
    }

    $.ajax({
        url: API_BASE + "/list",
        type: "POST",
        data: JSON.stringify({
            "access_type": access_type,
            "page": page,
            "token": get_basic_auth_token(),
            "tags": getTagsFromCurrentURL(),
        }),
        success: function (result) {
            $("#images-container").empty();
            if (result.images !== null) {
                var row_size = 0;
                var current_row = null;
                for (let image of result.images) {
                    if (image.file_name == null || image.file_name === "") {
                        continue
                    }

                    if (image.access_path == null || image.access_path === "") {
                        continue
                    }

                    if (row_size % 3 == 0) {
                        current_row = $("<div class='row padded-row'></div>");
                        $("#images-container").append(current_row);
                        row_size = 0;
                    }
                    var image_link = API_BASE + "/uploads/" + encodeURIComponent(image.access_path) + "/" + encodeURIComponent(image.file_name)

                    if (image.image_token != null && image.image_token !== "") {
                        image_link = insertParam(image_link, "token", encodeURIComponent(image.image_token))
                    }

                    thumbnail_link = insertParam(image_link, "thumbnail", "yes")
                    $(current_row).append("<div class='col-sm grid-image'><a target='_blank' href='" + image_link + "'><img class='grid-image' src='" + thumbnail_link + "' /></a></div>");

                    row_size++;
                }
            }

            // Disable backwards paging buttons if on first page
            if (page === 1) {
                $(".firstPage").addClass("disabled-paging-button");
                $(".firstPage").prop("disabled", true);
                $(".previousPage").addClass("disabled-paging-button");
                $(".previousPage").prop("disabled", true);
            } else {
                $(".firstPage").removeClass("disabled-paging-button");
                $(".firstPage").prop("disabled", false);
                $(".previousPage").removeClass("disabled-paging-button");
                $(".previousPage").prop("disabled", false);
            }

            // Disable forwards paging buttons unless the API states there are additional pages
            if (result.next_page === true) {
                $(".nextPage").removeClass("disabled-paging-button");
                $(".nextPage").prop("disabled", false);
            } else {
                $(".nextPage").addClass("disabled-paging-button");
                $(".nextPage").prop("disabled", true);
            }
        },
        error: function (result) {
            if (result.responseText == undefined || result.responseText == "") {
                $("#yronwood-error").text("Error: unknown (" + result.statusText + "). Connection to the API server may have failed.");
            } else {
                var err = $.parseJSON(result.responseText)
                $("#yronwood-error").text("Error: " + err.message + " (" + err.code + ")");
            }
        }
    });
}

$(document).on("click", "#authenticateIcon", function (event) {
    $('#authenticateModal').modal('show');
});

$(document).on("click", "#uploadIcon", function (event) {
    $('#uploadModal').modal('show');
});

$(document).on("click", "#authenticateButton", function (event) {
    $("#yronwood-success").text("");
    $("#yronwood-error").text("");
    $.ajax({
        url: API_BASE + "/authenticate",
        type: "POST",
        data: JSON.stringify({
            "secret": $("#authenticateSecret").val(),
        }),
        success: function (result) {
            if (result.token != null && result.token !== "") {
                $("#yronwood-success").text("Successfully authenticated!");
                set_basic_auth_token(result.token);
                set_authenticated();
                resetPaging();
            } else {
                $("#yronwood-error").text("Could not authenticate!");
            }
        },
        error: function (result) {
            if (result.responseText == undefined || result.responseText == "") {
                $("#yronwood-error").text("Error: unknown (" + result.statusText + ")");
            } else {
                var err = $.parseJSON(result.responseText)
                $("#yronwood-error").text("Error: " + err.message + " (" + err.code + ")");
            }
        }
    });
    // Clear the secret field value as it is a static modal.
    $("#authenticateSecret").val("");
});

$(document).keypress(function (e) {
    if ($("#authenticateModal").hasClass('show') && (e.keycode == 13 || e.which == 13)) {
        $("#authenticateButton").trigger("click");
    }
});

$(document).on("change", "#uploadFile", function () {
    var fileName = $(this).val();
    $(this).next('.custom-file-label').html(fileName);
})

function doUploadFile(payload, checksum, filename, sequence, total) {
    var fileNameComponents = filename.split(".")
    if (fileNameComponents == undefined || fileNameComponents.length < 2) {
        $("#yronwood-error").text("File must have an extension in name");
        return
    }

    $.ajax({
        url: API_BASE + "/upload",
        type: "PUT",
        data: JSON.stringify({
            "token": get_basic_auth_token(),
            "access_type": $("#accessTypeSelection").val(),
            "payload": payload,
            "checksum": checksum,
            "metadata": {
                "file_name": randomFileName(32) + "." + fileNameComponents[1],
                "tags": splitTags($("#fileTags").val()),
            }
        }),
        success: function (result) {
            $("#yronwood-success").text(`Image ${sequence} out of ${total} uploaded successfully.`);
        },
        error: function (result) {
            if (result.responseText == undefined || result.responseText == "") {
                $("#yronwood-error").text("Error: unknown (" + result.statusText + ")");
            } else {
                var err = $.parseJSON(result.responseText)
                $("#yronwood-error").text("Error: " + err.message + " (" + err.code + ")");
            }
        }
    });
    // Clear the upload path value as it is a static modal. We don't clear the file tags
    // as they are often reused between uploads.
    $("#uploadFile").val("");
}

$(document).on("click", "#uploadButton", function (event) {
    $("#yronwood-success").text("");
    $("#yronwood-error").text("");
    var files = $("#uploadFile").prop('files');
    if (files == undefined || files.length == 0) {
        $("#yronwood-error").text("You must select at least one file");
        $('#uploadModal').modal('hide');
        return
    }

    var total_files = files.length;
    Object.keys(files).forEach(i => {
        const file = files[i];
        const reader = new FileReader();
        reader.onload = (e) => {
            var arrayBuffer = reader.result
            var base64Payload = btoa([].reduce.call(new Uint8Array(arrayBuffer), function (p, c) { return p + String.fromCharCode(c) }, ''))
            digestMessage(base64Payload).then(function (digestValue) {
                doUploadFile(base64Payload, hexString(digestValue), file.name, parseInt(i) + 1, total_files);
            });
        }
        reader.readAsArrayBuffer(file);
    });
    resetPaging();
});

$(document).on("click", "#tagsReloadButton", function (event) {
    window.location = window.location.pathname;
});

$(".previousPage").unbind().click(function (event) {
    if (current_page > 1) {
        current_page--;
    }
    list_images(current_accss_type, current_page);
    $(".currentPage").text(current_page.toString());
});

$(".nextPage").unbind().click(function (event) {
    current_page++;
    list_images(current_accss_type, current_page);
    $(".currentPage").text(current_page.toString());
});

$(".firstPage").unbind().click(function (event) {
    resetPaging();
});

function resetPaging() {
    current_page = 1;
    list_images(current_accss_type, current_page);
    $(".currentPage").text(current_page.toString());
}

function set_authenticated() {
    current_accss_type = ACCESS_TYPE_PRIVATE;
    $("#authenticateIcon").removeClass("oi-lock-locked");
    $("#authenticateIcon").addClass("oi-lock-unlocked");
    $("#uploadIcon").removeAttr("hidden");
}

function set_basic_auth_token(token) {
    sessionStorage.setItem("yronwood_basic_auth_token", token);
}

function get_basic_auth_token() {
    var token = sessionStorage.getItem("yronwood_basic_auth_token");
    if (token === undefined || token === null) {
        return "";
    }

    return token
}

function hexString(buffer) {
    const byteArray = new Uint8Array(buffer);
    const hexCodes = [...byteArray].map(value => {
        const hexCode = value.toString(16);
        const paddedHexCode = hexCode.padStart(2, '0');
        return paddedHexCode;
    });

    return hexCodes.join('');
}

function digestMessage(message) {
    const encoder = new TextEncoder();
    const data = encoder.encode(message);
    return window.crypto.subtle.digest('SHA-256', data);
}

function randomFileName(length) {
    var result = '';
    var characters = 'abcdef0123456789';
    var charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}

function splitTags(commaSeparatedTags) {
    if (commaSeparatedTags === null || commaSeparatedTags.length === 0) {
        return [];
    }

    tagsSplit = commaSeparatedTags.split(",");
    tagsSplit.map(s => s.trim());
    return tagsSplit;
}

function getTagsFromCurrentURL() {
    const urlQueryString = window.location.search;
    const urlQueryParams = new URLSearchParams(urlQueryString);
    const tags = urlQueryParams.get("tags");
    if (tags !== null && tags !== "") {
        return splitTags(tags);
    }

    return [];
}

// Adapted from https://stackoverflow.com/a/487049
function insertParam(url, key, value) {
    key = encodeURI(key); value = encodeURI(value);

    var kvp = url.split('&');
    var i = kvp.length; var x; while (i--) {
        x = kvp[i].split('=');

        if (x[0] == key) {
            x[1] = value;
            kvp[i] = x.join('=');
            break;
        }
    }

    if (i < 0) {
        kvp[kvp.length] = [key, value].join('=');
    }

    if (!kvp[0].endsWith("?")) {
        kvp[0] = kvp[0] + "?"
    }

    return kvp.join('&');
}
