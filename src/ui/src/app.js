(function () {
  var pendingInitOptions = null;
  var bootStarted = false;

  function queueInit(options) {
    pendingInitOptions = options;
    return undefined;
  }

  function loadScript(source) {
    return new Promise(function (resolve, reject) {
      var script = document.createElement("script");
      script.src = source;
      script.onload = resolve;
      script.onerror = function () {
        reject(new Error("Failed to load script: " + source));
      };
      document.head.appendChild(script);
    });
  }

  function startBoot() {
    if (bootStarted) {
      return;
    }

    bootStarted = true;
    delete window.Elm;

    loadScript("/integrations/leaflet/leaflet-map.js")
      .then(function () {
        return loadScript("/elm.js");
      })
      .then(function () {
        if (!window.Elm || !window.Elm.Main || typeof window.Elm.Main.init !== "function") {
          throw new Error("Elm runtime failed to initialize");
        }

        if (pendingInitOptions) {
          window.Elm.Main.init(pendingInitOptions);
        }
      })
      .catch(function (error) {
        console.error(error);
      });
  }

  window.Elm = { Main: { init: queueInit } };
  window.setTimeout(startBoot, 0);
})();
