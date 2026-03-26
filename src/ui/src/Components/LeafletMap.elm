module Components.LeafletMap exposing (MapConfig, MarkerConfig, marker, view)

import Html exposing (Html, node)
import Html.Attributes exposing (attribute, class)


type alias MapConfig =
    { lat : Float
    , lng : Float
    , zoom : Int
    , fitToMarkers : Bool
    }


type alias MarkerConfig =
    { lat : Float
    , lng : Float
    , iconUrl : String
    }


view : MapConfig -> List (Html msg) -> Html msg
view config markers =
    node "dp-leaflet-map"
        [ class "d-block"
        , attribute "lat" (String.fromFloat config.lat)
        , attribute "lng" (String.fromFloat config.lng)
        , attribute "zoom" (String.fromInt config.zoom)
        , attribute "fit-to-markers" (boolAttributeValue config.fitToMarkers)
        ]
        markers


marker : MarkerConfig -> List (Html msg) -> Html msg
marker config popupContent =
    node "dp-map-marker"
        [ attribute "lat" (String.fromFloat config.lat)
        , attribute "lng" (String.fromFloat config.lng)
        , attribute "icon-url" config.iconUrl
        ]
        popupContent


boolAttributeValue : Bool -> String
boolAttributeValue value =
    if value then
        "true"

    else
        "false"
