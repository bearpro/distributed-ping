(function () {
  var LEAFLET_SCRIPT_URL = "https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/leaflet.js";
  var LEAFLET_STYLE_URL = "https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/leaflet.css";
  var MARKER_ICON_URL = "/integrations/leaflet/markers/circle.svg";
  var DEFAULT_ZOOM = 10;

  function ensureLeafletScript() {
    if (window.L) {
      return Promise.resolve(window.L);
    }

    if (window.__distributedPingLeafletPromise) {
      return window.__distributedPingLeafletPromise;
    }

    window.__distributedPingLeafletPromise = new Promise(function (resolve, reject) {
      var script = document.createElement("script");
      script.src = LEAFLET_SCRIPT_URL;
      script.onload = function () {
        resolve(window.L);
      };
      script.onerror = function () {
        reject(new Error("Failed to load Leaflet"));
      };
      document.head.appendChild(script);
    });

    return window.__distributedPingLeafletPromise;
  }

  function parseNumberAttribute(element, name, fallback) {
    var value = Number(element.getAttribute(name));
    return Number.isFinite(value) ? value : fallback;
  }

  function parseBooleanAttribute(element, name, fallback) {
    var value = element.getAttribute(name);

    if (value === null) {
      return fallback;
    }

    return value !== "false";
  }

  function ensureStyles(root) {
    if (root.querySelector('link[data-role="leaflet-style"]')) {
      return;
    }

    var baseStyle = document.createElement("style");
    baseStyle.textContent =
      ":host { display: block; width: 100%; }" +
      ".frame {" +
      "  border: 1px solid rgba(15, 23, 42, 0.12);" +
      "  border-radius: 18px;" +
      "  overflow: hidden;" +
      "  background: linear-gradient(180deg, #f8fafc 0%, #e2e8f0 100%);" +
      "  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.08);" +
      "}" +
      ".map {" +
      "  width: 100%;" +
      "  min-height: 360px;" +
      "  height: var(--leaflet-map-height, 420px);" +
      "}" +
      ".status {" +
      "  display: grid;" +
      "  place-items: center;" +
      "  min-height: 360px;" +
      "  padding: 1rem;" +
      "  color: #334155;" +
      "  font: 500 0.95rem/1.4 system-ui, sans-serif;" +
      "}" +
      ".status[hidden], .map[hidden] {" +
      "  display: none !important;" +
      "}";
    root.appendChild(baseStyle);

    var leafletStyle = document.createElement("link");
    leafletStyle.rel = "stylesheet";
    leafletStyle.href = LEAFLET_STYLE_URL;
    leafletStyle.setAttribute("data-role", "leaflet-style");
    root.appendChild(leafletStyle);
  }

  function createMarkerIcon(L) {
    return L.icon({
      iconUrl: MARKER_ICON_URL,
      iconSize: [24, 24],
      iconAnchor: [12, 12],
      popupAnchor: [0, -12],
      className: "distributed-ping-marker",
    });
  }

  function createMarkerIconFromConfig(L, config) {
    return L.icon({
      iconUrl: config.iconUrl || MARKER_ICON_URL,
      iconSize: [24, 24],
      iconAnchor: [12, 12],
      popupAnchor: [0, -12],
      className: "distributed-ping-marker",
    });
  }

  function clonePopupContent(source) {
    var container = document.createElement("div");

    Array.prototype.forEach.call(source.childNodes, function (node) {
      container.appendChild(node.cloneNode(true));
    });

    return container;
  }

  function readMarkerConfig(element) {
    return {
      lat: parseNumberAttribute(element, "lat", 0),
      lng: parseNumberAttribute(element, "lng", 0),
      iconUrl: element.getAttribute("icon-url") || MARKER_ICON_URL,
      popupContent: clonePopupContent(element),
    };
  }

  class DistributedPingMapMarker extends HTMLElement {
    connectedCallback() {
      this.hidden = true;
    }
  }

  class DistributedPingLeafletMap extends HTMLElement {
    static get observedAttributes() {
      return ["lat", "lng", "zoom", "fit-to-markers"];
    }

    constructor() {
      super();
      this.attachShadow({ mode: "open" });
      this.map = null;
      this.markerLayer = null;
      this.leaflet = null;
      this.mapRoot = null;
      this.statusNode = null;
      this.markerObserver = null;
    }

    connectedCallback() {
      this.renderShell();
      this.startMarkerObserver();
      this.initialize();
    }

    disconnectedCallback() {
      if (this.markerObserver) {
        this.markerObserver.disconnect();
        this.markerObserver = null;
      }

      if (this.map) {
        this.map.remove();
        this.map = null;
      }

      this.markerLayer = null;
      this.leaflet = null;
    }

    attributeChangedCallback() {
      this.syncMapState();
    }

    renderShell() {
      if (this.mapRoot) {
        return;
      }

      ensureStyles(this.shadowRoot);

      var frame = document.createElement("div");
      frame.className = "frame";

      this.statusNode = document.createElement("div");
      this.statusNode.className = "status";
      this.statusNode.textContent = "Loading map...";

      this.mapRoot = document.createElement("div");
      this.mapRoot.className = "map";
      this.mapRoot.hidden = true;

      frame.appendChild(this.statusNode);
      frame.appendChild(this.mapRoot);
      this.shadowRoot.appendChild(frame);
    }

    startMarkerObserver() {
      if (this.markerObserver) {
        return;
      }

      var element = this;

      this.markerObserver = new MutationObserver(function () {
        element.syncMapState();
      });

      this.markerObserver.observe(this, {
        attributes: true,
        childList: true,
        subtree: true,
      });
    }

    initialize() {
      var element = this;

      ensureLeafletScript()
        .then(function (L) {
          if (!element.isConnected || element.map) {
            return;
          }

          element.statusNode.hidden = true;
          element.mapRoot.hidden = false;
          element.leaflet = L;

          element.map = L.map(element.mapRoot, {
            zoomControl: true,
            attributionControl: true,
          });
          element.markerLayer = L.layerGroup().addTo(element.map);

          L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            maxZoom: 19,
            attribution: "&copy; OpenStreetMap contributors",
          }).addTo(element.map);

          element.syncMapState();

          window.requestAnimationFrame(function () {
            if (element.map) {
              element.map.invalidateSize();
            }
          });
        })
        .catch(function (error) {
          console.error(error);
          if (element.statusNode) {
            element.statusNode.textContent = "Map failed to load.";
            element.statusNode.hidden = false;
          }
        });
    }

    syncMapState() {
      if (!this.map || !this.markerLayer || !this.leaflet) {
        return;
      }

      var L = this.leaflet;
      var lat = parseNumberAttribute(this, "lat", 55.751244);
      var lng = parseNumberAttribute(this, "lng", 37.618423);
      var zoom = parseNumberAttribute(this, "zoom", DEFAULT_ZOOM);
      var fitToMarkers = parseBooleanAttribute(this, "fit-to-markers", true);
      var markerElements = this.querySelectorAll("dp-map-marker");
      var points = [];
      var bounds = [];

      this.markerLayer.clearLayers();

      Array.prototype.forEach.call(
        markerElements,
        function (markerElement) {
          var config = readMarkerConfig(markerElement);
          var position = [config.lat, config.lng];
          var marker = L.marker(position, {
            icon: createMarkerIconFromConfig(L, config),
          });

          if (config.popupContent.childNodes.length > 0) {
            marker.bindPopup(config.popupContent);
          }

          marker.addTo(this.markerLayer);
          points.push(position);
          bounds.push(position);
        }.bind(this)
      );

      if (points.length === 0) {
        this.map.setView([lat, lng], zoom);
        return;
      }

      if (fitToMarkers && points.length > 1) {
        this.map.fitBounds(bounds, { padding: [32, 32] });
        return;
      }

      this.map.setView(points[0], zoom);
    }
  }

  if (!customElements.get("dp-leaflet-map")) {
    customElements.define("dp-leaflet-map", DistributedPingLeafletMap);
  }

  if (!customElements.get("dp-map-marker")) {
    customElements.define("dp-map-marker", DistributedPingMapMarker);
  }
})();
