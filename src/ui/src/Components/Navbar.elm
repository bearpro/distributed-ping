module Components.Navbar exposing (Model, Msg, init, update, view)

import Abstractions exposing (Page)
import Html exposing (Html, a, div, li, nav, text, ul)
import Html.Attributes exposing (class, href)
import Maybe exposing (withDefault)


type alias PageDescription =
    { title : String, key : String }


type alias Model =
    { pages : List PageDescription
    , currentPageKey : Maybe String
    }


type Msg
    = NoOp


init : List PageDescription -> ( Model, Cmd Msg )
init pages =
    ( { pages = pages
      , currentPageKey = Maybe.map (\p -> p.key) (List.head pages)
      }
    , Cmd.none
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )


navItemView title key isActive =
    let
        activeMark =
            if isActive then
                [ class "active" ]

            else
                []
    in
    li [ class "nav-item" ]
        [ a (List.append activeMark [ class "nav-link", href key ]) [ text title ]
        ]


view : Model -> Html Msg
view model =
    nav
        [ class "navbar", class "navbar-expand-lg", class "justify-content-center", class "bg-body-tertiary", class "sticky-top" ]
        [ div
            [ class "container-fluid" ]
            [ ul [ class "navbar-nav" ]
                (List.map
                    (\p -> navItemView p.title p.key (Maybe.map (\ck -> p.key == ck) model.currentPageKey |> Maybe.withDefault False))
                    model.pages
                )
            ]
        ]
