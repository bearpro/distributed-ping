module Main exposing (main)

import Browser
import Html exposing (Html)
import Html.Events
import Tabs


type alias Flags =
    Maybe String


type alias Model =
    Tabs.Model


type Msg
    = TabsMsg Tabs.Msg


main : Program Flags Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }


init : Flags -> ( Model, Cmd Msg )
init flags =
    let
        ( tabsModel, tabsCmd ) =
            Tabs.init flags
    in
    ( tabsModel, Cmd.map TabsMsg tabsCmd )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        TabsMsg tabsMsg ->
            let
                ( nextModel, nextCmd ) =
                    Tabs.update tabsMsg model
            in
            ( nextModel, Cmd.map TabsMsg nextCmd )


view : Model -> Html Msg
view model =
    Html.map TabsMsg (Tabs.view model)


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.map TabsMsg (Tabs.subscriptions model)
