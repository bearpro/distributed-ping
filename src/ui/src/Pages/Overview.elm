module Pages.Overview exposing (Model, Msg, init, page, update, view)

import Abstractions exposing (Page)
import Html exposing (Html, a, div, li, nav, span, text, ul)
import Html.Attributes exposing (class, href)


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
    div [] [ span [] [ text "This is what I am talking about" ] ]


page : Page Model Msg
page =
    { title = "Overview"
    , key = "/overview"
    , init = init
    , view = view
    , update = update
    }
