module Main exposing (main)

import Browser
import Html exposing (Html, div, text)


type alias Flags =
    Maybe String


type alias Model =
    ()


type Msg
    = NoOp


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }


init : () -> ( Model, Cmd Msg )
init _ =
    ( (), Cmd.none )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )


view : Model -> Html Msg
view _ =
    div [] [ text "Hello, world!" ]


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none
