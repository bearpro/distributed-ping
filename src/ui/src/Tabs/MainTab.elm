module Tabs.MainTab exposing (Model, Msg, init, subscriptions, update, view)

import Html exposing (Html, div, h2, p, text)
import Html.Attributes exposing (class)


type alias Model =
    {}


type Msg
    = NoOp


init : ( Model, Cmd Msg )
init =
    ( {}, Cmd.none )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )


view : Model -> Html Msg
view _ =
    div [ class "d-flex flex-column gap-3" ]
        [ div []
            [ h2 [ class "h3 mb-1" ] [ text "Main" ]
            , p [ class "text-body-secondary mb-0" ] [ text "This tab is reserved for the main dashboard." ]
            ]
        , div [ class "card shadow-sm" ]
            [ div [ class "card-body" ]
                [ p [ class "mb-0" ] [ text "Nothing is here yet." ] ]
            ]
        ]


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none
