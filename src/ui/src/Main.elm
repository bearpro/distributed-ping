module Main exposing (main)

import Browser
import Browser.Navigation
import Components.Navbar as Navbar
import Html exposing (Html, div, text)
import Html.Attributes exposing (class)
import Pages.About
import Pages.Api
import Pages.NodeState
import Pages.Overview
import Route
import Url


type CurrentPage
    = OverviewPage Pages.Overview.Model
    | ApiPage Pages.Api.Model
    | AboutPage Pages.About.Model
    | NodeStatePage Pages.NodeState.Model
    | NotFoundPage


type alias Model =
    { navigationKey : Browser.Navigation.Key
    , navbar : Navbar.Model
    , route : Route.Route
    , currentPage : CurrentPage
    }


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | NavbarMsg Navbar.Msg
    | OverviewMsg Pages.Overview.Msg
    | ApiMsg Pages.Api.Msg
    | AboutMsg Pages.About.Msg
    | NodeStateMsg Pages.NodeState.Msg


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
        ( tabModel, tabCmd ) =
            Navbar.init Route.navigationPages

        route =
            Route.parse url

        ( currentPage, pageCmd ) =
            initCurrentPage route

        navbar =
            Navbar.selectPage (Route.pageKey route) tabModel
    in
    ( { navigationKey = key
      , currentPage = currentPage
      , route = route
      , navbar = navbar
      }
    , Cmd.batch [ Cmd.map NavbarMsg tabCmd, pageCmd ]
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Browser.Navigation.pushUrl model.navigationKey (Url.toString url) )

                Browser.External href ->
                    ( model, Browser.Navigation.load href )

        UrlChanged url ->
            let
                route =
                    Route.parse url

                ( currentPage, pageCmd ) =
                    initCurrentPage route
            in
            ( { model
                | currentPage = currentPage
                , route = route
                , navbar = Navbar.selectPage (Route.pageKey route) model.navbar
              }
            , pageCmd
            )

        NavbarMsg tabsMsg ->
            let
                ( newTabModel, newTabsMsg ) =
                    Navbar.update tabsMsg model.navbar
            in
            ( { model | navbar = newTabModel }
            , Cmd.map NavbarMsg newTabsMsg
            )

        OverviewMsg pageMsg ->
            case model.currentPage of
                OverviewPage pageModel ->
                    let
                        ( newPageModel, pageCmd ) =
                            Pages.Overview.update pageMsg pageModel
                    in
                    ( { model | currentPage = OverviewPage newPageModel }
                    , Cmd.map OverviewMsg pageCmd
                    )

                _ ->
                    ( model, Cmd.none )

        ApiMsg pageMsg ->
            case model.currentPage of
                ApiPage pageModel ->
                    let
                        ( newPageModel, pageCmd ) =
                            Pages.Api.update pageMsg pageModel
                    in
                    ( { model | currentPage = ApiPage newPageModel }
                    , Cmd.map ApiMsg pageCmd
                    )

                _ ->
                    ( model, Cmd.none )

        NodeStateMsg pageMsg ->
            case model.currentPage of
                NodeStatePage pageModel ->
                    let
                        ( newPageModel, pageCmd ) =
                            Pages.NodeState.update pageMsg pageModel
                    in
                    ( { model | currentPage = NodeStatePage newPageModel }
                    , Cmd.map NodeStateMsg pageCmd
                    )

                _ ->
                    ( model, Cmd.none )

        AboutMsg pageMsg ->
            case model.currentPage of
                AboutPage pageModel ->
                    let
                        ( newPageModel, pageCmd ) =
                            Pages.About.update pageMsg pageModel
                    in
                    ( { model | currentPage = AboutPage newPageModel }
                    , Cmd.map AboutMsg pageCmd
                    )

                _ ->
                    ( model, Cmd.none )


view : Model -> Browser.Document Msg
view model =
    let
        body =
            [ div []
                [ Html.map NavbarMsg (Navbar.view model.navbar)
                , div [ class "container py-4" ] [ currentPageView model.currentPage ]
                ]
            ]

        title =
            Route.pageTitle model.route ++ " | Distributed ping"
    in
    { title = title, body = body }


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none


initCurrentPage : Route.Route -> ( CurrentPage, Cmd Msg )
initCurrentPage route =
    case route of
        Route.Overview ->
            let
                ( pageModel, pageCmd ) =
                    Pages.Overview.init ()
            in
            ( OverviewPage pageModel, Cmd.map OverviewMsg pageCmd )

        Route.Api ->
            let
                ( pageModel, pageCmd ) =
                    Pages.Api.init ()
            in
            ( ApiPage pageModel, Cmd.map ApiMsg pageCmd )

        Route.NodeState ->
            let
                ( pageModel, pageCmd ) =
                    Pages.NodeState.init ()
            in
            ( NodeStatePage pageModel, Cmd.map NodeStateMsg pageCmd )

        Route.About ->
            let
                ( pageModel, pageCmd ) =
                    Pages.About.init ()
            in
            ( AboutPage pageModel, Cmd.map AboutMsg pageCmd )

        Route.NotFound ->
            ( NotFoundPage, Cmd.none )


currentPageView : CurrentPage -> Html Msg
currentPageView currentPage =
    case currentPage of
        OverviewPage pageModel ->
            Html.map OverviewMsg (Pages.Overview.view pageModel)

        ApiPage pageModel ->
            Html.map ApiMsg (Pages.Api.view pageModel)

        AboutPage pageModel ->
            Html.map AboutMsg (Pages.About.view pageModel)

        NodeStatePage pageModel ->
            Html.map NodeStateMsg (Pages.NodeState.view pageModel)

        NotFoundPage ->
            div [] [ text "Page not found." ]
