module Abstractions exposing (Page)

import Html exposing (Html)


type alias Page model msg =
    { title : String
    , key : String
    , init : () -> ( model, Cmd msg )
    , update : msg -> model -> ( model, Cmd msg )
    , view : model -> Html msg
    }
