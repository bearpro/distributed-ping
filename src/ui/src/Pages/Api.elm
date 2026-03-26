module Pages.Api exposing (Model, Msg, init, page, update, view)

import Abstractions exposing (Page)
import Html exposing (Html, a, div, iframe, li, nav, span, text, ul)
import Html.Attributes exposing (class, href, src, title)


type alias Model =
    ()


type Msg
    = NoOp


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
    div [ class "vh-100" ]
        [ iframe
            [ class "w-100"
            , class "h-100"
            , title "swagger"
            , src "/api/swagger/index.html"
            ]
            []
        ]


page : Page Model Msg
page =
    { title = "API"
    , key = "/docs"
    , init = init
    , view = view
    , update = update
    }
