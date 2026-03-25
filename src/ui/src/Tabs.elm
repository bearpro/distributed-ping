module Tabs exposing (Model, Msg, init, subscriptions, update, view)

import Html exposing (Html, button, div, h1, p, small, text)
import Html.Attributes exposing (class, classList, type_)
import Html.Events exposing (onClick)
import Tabs.MainTab as MainTab
import Tabs.NodeStateTab as NodeStateTab


type alias Model =
    { activeTab : Tab
    , mainTab : MainTab.Model
    , nodeStateTab : NodeStateTab.Model
    }


type Tab
    = MainTabId
    | NodeStateTabId


type Msg
    = SelectTab Tab
    | MainTabMsg MainTab.Msg
    | NodeStateTabMsg NodeStateTab.Msg


init : Maybe String -> ( Model, Cmd Msg )
init flags =
    let
        ( mainTabModel, mainTabCmd ) =
            MainTab.init

        ( nodeStateTabModel, nodeStateTabCmd ) =
            NodeStateTab.init flags
    in
    ( { activeTab = MainTabId
      , mainTab = mainTabModel
      , nodeStateTab = nodeStateTabModel
      }
    , Cmd.batch
        [ Cmd.map MainTabMsg mainTabCmd
        , Cmd.map NodeStateTabMsg nodeStateTabCmd
        ]
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SelectTab tab ->
            selectTab tab model

        MainTabMsg childMsg ->
            let
                ( nextTabModel, nextCmd ) =
                    MainTab.update childMsg model.mainTab
            in
            ( { model | mainTab = nextTabModel }
            , Cmd.map MainTabMsg nextCmd
            )

        NodeStateTabMsg childMsg ->
            let
                ( nextTabModel, nextCmd ) =
                    NodeStateTab.update childMsg model.nodeStateTab
            in
            ( { model | nodeStateTab = nextTabModel }
            , Cmd.map NodeStateTabMsg nextCmd
            )


selectTab : Tab -> Model -> ( Model, Cmd Msg )
selectTab tab model =
    let
        baseModel =
            { model | activeTab = tab }
    in
    case tab of
        MainTabId ->
            ( baseModel, Cmd.none )

        NodeStateTabId ->
            let
                ( nextTabModel, nextCmd ) =
                    NodeStateTab.activate model.nodeStateTab
            in
            ( { baseModel | nodeStateTab = nextTabModel }
            , Cmd.map NodeStateTabMsg nextCmd
            )


view : Model -> Html Msg
view model =
    div [ class "container-fluid" ]
        [ div [ class "row min-vh-100" ]
            [ viewSidebar model.activeTab
            , viewContent model
            ]
        ]


viewSidebar : Tab -> Html Msg
viewSidebar activeTab =
    div [ class "col-12 col-md-4 col-lg-3 col-xl-2 border-end bg-body-tertiary" ]
        [ div [ class "d-flex flex-column gap-4 p-3 p-lg-4" ]
            [ div []
                [ p [ class "text-uppercase text-secondary fw-semibold small mb-2" ] [ text "Distributed Ping" ]
                ]
            , div [ class "nav nav-pills flex-column gap-2" ]
                [ viewTabButton activeTab MainTabId
                , viewTabButton activeTab NodeStateTabId
                ]
            ]
        ]


viewTabButton : Tab -> Tab -> Html Msg
viewTabButton activeTab tab =
    let
        isActive =
            activeTab == tab
    in
    button
        [ classList
            [ ( "nav-link", True )
            , ( "text-start", True )
            , ( "active", isActive )
            ]
        , type_ "button"
        , onClick (SelectTab tab)
        ]
        [ div [ class "fw-semibold" ] [ text (tabLabel tab) ]
        ]


viewContent : Model -> Html Msg
viewContent model =
    div [ class "col-12 col-md-8 col-lg-9 col-xl-10" ]
        [ div [ class "p-3 p-lg-4 d-flex flex-column gap-4" ]
            [ case model.activeTab of
                MainTabId ->
                    Html.map MainTabMsg (MainTab.view model.mainTab)

                NodeStateTabId ->
                    Html.map NodeStateTabMsg (NodeStateTab.view model.nodeStateTab)
            ]
        ]


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Sub.map MainTabMsg (MainTab.subscriptions model.mainTab)
        , Sub.map NodeStateTabMsg (NodeStateTab.subscriptions model.nodeStateTab)
        ]


tabLabel : Tab -> String
tabLabel tab =
    case tab of
        MainTabId ->
            "Main"

        NodeStateTabId ->
            "Node state"


tabDescription : Tab -> String
tabDescription tab =
    case tab of
        MainTabId ->
            "Placeholder for the main dashboard."

        NodeStateTabId ->
            "Inspect the node and controller snapshot."
