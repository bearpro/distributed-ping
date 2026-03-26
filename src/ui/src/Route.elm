module Route exposing (Route(..), navigationPages, pageKey, pageTitle, parse)

import Pages.About
import Pages.Api
import Pages.NodeState
import Pages.Overview
import Url


type Route
    = Overview
    | Api
    | About
    | NodeState
    | NotFound


type alias RouteDefinition =
    { route : Route
    , title : String
    , key : String
    }


navigationPages : List { title : String, key : String }
navigationPages =
    List.map pageDescription routeDefinitions


parse : Url.Url -> Route
parse url =
    case routeForPath url.path of
        Just route ->
            route

        Nothing ->
            if url.path == "/" then
                defaultRoute

            else
                NotFound


pageKey : Route -> Maybe String
pageKey route =
    findRouteDefinition route |> Maybe.map .key


pageTitle : Route -> String
pageTitle route =
    case findRouteDefinition route of
        Just definition ->
            definition.title

        Nothing ->
            "Not found"


routeDefinitions : List RouteDefinition
routeDefinitions =
    [ { route = Overview
      , title = Pages.Overview.page.title
      , key = Pages.Overview.page.key
      }
    , { route = Api
      , title = Pages.Api.page.title
      , key = Pages.Api.page.key
      }
    , { route = NodeState
      , title = Pages.NodeState.page.title
      , key = Pages.NodeState.page.key
      }
    , { route = About
      , title = Pages.About.page.title
      , key = Pages.About.page.key
      }
    ]


defaultRoute : Route
defaultRoute =
    routeDefinitions
        |> List.head
        |> Maybe.map .route
        |> Maybe.withDefault NotFound


routeForPath : String -> Maybe Route
routeForPath path =
    routeDefinitions
        |> List.filter (\definition -> definition.key == path)
        |> List.head
        |> Maybe.map .route


findRouteDefinition : Route -> Maybe RouteDefinition
findRouteDefinition route =
    routeDefinitions
        |> List.filter (\definition -> definition.route == route)
        |> List.head


pageDescription : RouteDefinition -> { title : String, key : String }
pageDescription definition =
    { title = definition.title, key = definition.key }
