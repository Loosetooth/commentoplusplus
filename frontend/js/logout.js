(function (global, document) {
  "use strict";

  global.logout = function() {
    var done = function() {
      global.cookieDelete("commentoOwnerToken");
      document.location = global.origin + "/login";
    };

    var ownerToken = global.cookieGet("commentoOwnerToken");
    if (!ownerToken) {
      done();
      return;
    }

    // Invalidate the session on the server too.
    var xhr = new XMLHttpRequest();
    xhr.open("POST", global.origin + "/api/owner/logout", true);
    xhr.onloadend = done;
    xhr.send(JSON.stringify({"ownerToken": ownerToken}));
  }

} (window.commento, document));
