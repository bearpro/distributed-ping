module Main exposing (main)

import Abstractions exposing (Page)
import Browser
import Browser.Navigation
import Components.Navbar as Navbar
import Html exposing (Html, div, text)
import Html.Attributes exposing (title)
import Pages.About
import Pages.Api
import Pages.Overview
import Url


type CurrentPage
    = Overview Pages.Overview.Model
    | Api Pages.Api.Model
    | About Pages.About.Model


type alias Model =
    { navbar : Navbar.Model
    , currentPage : CurrentPage
    }


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | NavbarMsg Navbar.Msg


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        , onUrlRequest = LinkClicked
        , onUrlChange = UrlChanged
        }


init : () -> Url.Url -> Browser.Navigation.Key -> ( Model, Cmd Msg )
init () url key =
    let
        pageDescription =
            \page -> { title = page.title, key = page.key }

        pages =
            [ pageDescription Pages.Overview.page
            , pageDescription Pages.Api.page
            , pageDescription Pages.About.page
            ]

        ( tabModel, tabCmd ) =
            Navbar.init pages

        ( pageModel, pageCmd ) =
            Pages.Overview.init ()

        currentPage =
            Overview pageModel
    in
    ( { currentPage = currentPage, navbar = tabModel }
    , Cmd.map NavbarMsg tabCmd
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NavbarMsg tabsMsg ->
            let
                ( newTabModel, newTabsMsg ) =
                    Navbar.update tabsMsg model.navbar
            in
            ( { model | navbar = newTabModel }
            , Cmd.map NavbarMsg newTabsMsg
            )

        _ ->
            ( model, Cmd.none )


view : Model -> Browser.Document Msg
view model =
    let
        body =
            [ div []
                [ Html.map NavbarMsg (Navbar.view model.navbar)
                ]
            ]

        title =
            "Distributed ping"
    in
    { title = title, body = body }


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none
