module Pages.Overview exposing (Model, Msg, init, page, update, view)

import Abstractions exposing (Page)
import Components.LeafletMap as LeafletMap
import Html exposing (Html, div, h1, p, span, strong, text)
import Html.Attributes exposing (class)
import Http
import Json.Decode as Decode
import String


type alias Model =
    { nodeSummary : Maybe NodeSummary
    , error : Bool
    }


type alias NodeSummary =
    { node : NodeInfo
    , isController : Bool
    , upstreamController : Maybe String
    , currentOffset : Int
    , submittedRequests : Int
    }


type alias NodeInfo =
    { nodeId : String
    , organization : Maybe String
    , lat : Float
    , lon : Float
    , role : String
    , publicIp : Maybe String
    }


type Msg
    = GotNodeSummary (Result Http.Error NodeSummary)


init : () -> ( Model, Cmd Msg )
init _ =
    ( { nodeSummary = Nothing, error = False }
    , getNodeSummary
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotNodeSummary result ->
            case result of
                Ok summary ->
                    ( { model | nodeSummary = Just summary, error = False }, Cmd.none )

                Err _ ->
                    ( { model | error = True }, Cmd.none )


view : Model -> Html Msg
view model =
    div [ class "d-grid gap-4" ]
        [ div [ class "d-grid gap-2" ]
            [ h1 [ class "h3 mb-0" ] [ text "Network overview" ]
            , p [ class "text-body-secondary mb-0" ]
                [ text "Leaflet is mounted through a custom element, so Elm renders marker metadata and popup markup while JavaScript owns the map lifecycle." ]
            ]
        , viewMap model
        , viewStatus model
        ]


viewMap : Model -> Html Msg
viewMap model =
    case model.nodeSummary of
        Just summary ->
            let
                node =
                    summary.node
            in
            LeafletMap.view
                { lat = node.lat
                , lng = node.lon
                , zoom = 10
                , fitToMarkers = False
                }
                [ LeafletMap.marker
                    { lat = node.lat
                    , lng = node.lon
                    , iconUrl = iconUrlForRole node.role
                    }
                    (nodePopup summary)
                ]

        Nothing ->
            div [ class "rounded border bg-body-tertiary p-4 text-body-secondary" ]
                [ text
                    (if model.error then
                        "Failed to load /api/node."

                     else
                        "Loading node location..."
                    )
                ]


viewStatus : Model -> Html Msg
viewStatus model =
    case model.nodeSummary of
        Just summary ->
            div [ class "text-body-secondary" ]
                [ span []
                    [ text ("Current node: " ++ summary.node.nodeId ++ " (" ++ summary.node.role ++ ")")
                    ]
                ]

        Nothing ->
            div [ class "text-body-secondary" ]
                [ span []
                    [ text "The marker position and popup are populated from /api/node." ]
                ]


nodePopup : NodeSummary -> List (Html Msg)
nodePopup summary =
    [ div [ class "d-grid gap-1" ]
        ([ strong [] [ text summary.node.nodeId ]
         , span [ class "text-body-secondary" ] [ text ("Role: " ++ summary.node.role) ]
         , span [ class "text-body-secondary" ]
            [ text ("Coordinates: " ++ String.fromFloat summary.node.lat ++ ", " ++ String.fromFloat summary.node.lon) ]
         , span [ class "text-body-secondary" ]
            [ text ("Submitted requests: " ++ String.fromInt summary.submittedRequests) ]
         , span [ class "text-body-secondary" ]
            [ text ("Current offset: " ++ String.fromInt summary.currentOffset) ]
         ]
            ++ maybeInfoLine "Organization" summary.node.organization
            ++ maybeInfoLine "Public IP" summary.node.publicIp
            ++ maybeInfoLine "Upstream controller" summary.upstreamController
        )
    ]


maybeInfoLine : String -> Maybe String -> List (Html Msg)
maybeInfoLine label value =
    case value of
        Just actual ->
            if String.isEmpty actual then
                []

            else
                [ span [ class "text-body-secondary" ] [ text (label ++ ": " ++ actual) ] ]

        Nothing ->
            []


iconUrlForRole : String -> String
iconUrlForRole role =
    case role of
        "controller" ->
            "/integrations/leaflet/markers/pulse.svg"

        "hybrid" ->
            "/integrations/leaflet/markers/pulse.svg"

        _ ->
            "/integrations/leaflet/markers/circle.svg"


getNodeSummary : Cmd Msg
getNodeSummary =
    Http.get
        { url = "/api/node"
        , expect = Http.expectJson GotNodeSummary nodeSummaryDecoder
        }


nodeSummaryDecoder : Decode.Decoder NodeSummary
nodeSummaryDecoder =
    Decode.map5
        (\node isController upstreamController currentOffset submittedRequests ->
            { node = node
            , isController = isController
            , upstreamController = upstreamController
            , currentOffset = currentOffset
            , submittedRequests = submittedRequests
            }
        )
        (Decode.field "node" nodeInfoDecoder)
        (Decode.field "is_controller" Decode.bool)
        (Decode.maybe (Decode.field "upstream_controller" Decode.string))
        (Decode.field "current_offset" Decode.int)
        (Decode.field "submitted_requests" Decode.int)


nodeInfoDecoder : Decode.Decoder NodeInfo
nodeInfoDecoder =
    Decode.map6
        (\nodeId organization lat lon role publicIp ->
            { nodeId = nodeId
            , organization = organization
            , lat = lat
            , lon = lon
            , role = role
            , publicIp = publicIp
            }
        )
        (Decode.field "node_id" Decode.string)
        (Decode.maybe (Decode.field "organization" Decode.string))
        (Decode.at [ "location", "lat" ] Decode.float)
        (Decode.at [ "location", "lon" ] Decode.float)
        (Decode.field "role" Decode.string)
        (Decode.maybe (Decode.field "public_ip" Decode.string))


page : Page Model Msg
page =
    { title = "Overview"
    , key = "/overview"
    , init = init
    , view = view
    , update = update
    }
