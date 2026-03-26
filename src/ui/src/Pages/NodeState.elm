module Pages.NodeState exposing (Model, Msg, init, page, update, view)

import Abstractions exposing (Page)
import Html exposing (Html, a, div, li, nav, pre, span, text, ul)
import Html.Attributes exposing (class, href)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Url exposing (Protocol(..))


type alias Model =
    { nodeState : Maybe String
    , error : Bool
    }


type Msg
    = GotNodeState (Result Http.Error String)


getNodeState : Cmd Msg
getNodeState =
    Http.get
        { url = "/api/application", expect = Http.expectString GotNodeState }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { nodeState = Nothing, error = False }, getNodeState )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotNodeState result ->
            case result of
                Ok ns ->
                    ( { model | nodeState = Just ns }, Cmd.none )

                Err err ->
                    ( { model | error = True }, Cmd.none )


view : Model -> Html Msg
view model =
    case ( model.nodeState, model.error ) of
        ( Just state, _ ) ->
            div []
                [ pre
                    [ class "mono-text", class "text-bg-dark" ]
                    [ text (prettyJson state) ]
                ]

        ( Nothing, False ) ->
            div []
                [ span []
                    [ text "loading..."
                    ]
                ]

        ( Nothing, True ) ->
            div []
                [ span []
                    [ text "Error when attempting to load nod state"
                    ]
                ]


prettyJson : String -> String
prettyJson raw =
    case Decode.decodeString Decode.value raw of
        Ok value ->
            Encode.encode 2 value

        Err _ ->
            raw


page : Page Model Msg
page =
    { title = "Node"
    , key = "/node-state"
    , init = init
    , view = view
    , update = update
    }
