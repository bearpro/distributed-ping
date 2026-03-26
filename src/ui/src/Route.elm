module Route exposing (Route(..), navigationPages, pageKey, pageTitle, parse)

import Pages.About
import Pages.Api
import Pages.Overview
import Url


type Route
    = Overview
    | Api
    | About
    | NotFound


navigationPages : List { title : String, key : String }
navigationPages =
    [ pageDescription Pages.Overview.page
    , pageDescription Pages.Api.page
    , pageDescription Pages.About.page
    ]


parse : Url.Url -> Route
parse url =
    case url.path of
        "/" ->
            Overview

        "/overview" ->
            Overview

        "/api-doc" ->
            Api

        "/about" ->
            About

        _ ->
            NotFound


pageKey : Route -> Maybe String
pageKey route =
    case route of
        Overview ->
            Just Pages.Overview.page.key

        Api ->
            Just Pages.Api.page.key

        About ->
            Just Pages.About.page.key

        NotFound ->
            Nothing


pageTitle : Route -> String
pageTitle route =
    case route of
        Overview ->
            Pages.Overview.page.title

        Api ->
            Pages.Api.page.title

        About ->
            Pages.About.page.title

        NotFound ->
            "Not found"


pageDescription : { page | title : String, key : String } -> { title : String, key : String }
pageDescription page =
    { title = page.title, key = page.key }
