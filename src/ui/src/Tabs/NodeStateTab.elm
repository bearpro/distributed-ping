module Tabs.NodeStateTab exposing (Model, Msg, activate, init, subscriptions, update, view)

import Html exposing (Html, button, div, h2, p, pre, span, text)
import Html.Attributes exposing (class, disabled, type_)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import String


type alias Model =
    { serverUrl : Maybe String
    , status : Status
    , response : Maybe String
    }


type Status
    = Idle
    | Loading
    | Ready
    | Failed String


type Msg
    = Refresh
    | GotApplication (Result Http.Error Decode.Value)


init : Maybe String -> ( Model, Cmd Msg )
init flags =
    ( { serverUrl = normalizeServerUrl flags
      , status = Idle
      , response = Nothing
      }
    , Cmd.none
    )


activate : Model -> ( Model, Cmd Msg )
activate model =
    if shouldFetch model then
        refresh model

    else
        ( model, Cmd.none )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Refresh ->
            refresh model

        GotApplication result ->
            case result of
                Ok body ->
                    ( { model | status = Ready, response = Just (Encode.encode 4 body) }
                    , Cmd.none
                    )

                Err error ->
                    ( { model | status = Failed (httpErrorToString error), response = Nothing }
                    , Cmd.none
                    )


view : Model -> Html Msg
view model =
    div [ class "d-flex flex-column gap-4" ]
        [ div []
            [ h2 [ class "h3 mb-1" ] [ text "Node state" ]
            , p [ class "text-body-secondary mb-0" ]
                [ text "Inspect the current application snapshot from "
                , span [ class "font-monospace" ] [ text "/api/application" ]
                , text "."
                ]
            ]
        , div [ class "card shadow-sm" ]
            [ div [ class "card-body" ]
                [ div [ class "d-flex flex-column flex-lg-row gap-3 align-items-lg-center justify-content-lg-between" ]
                    [ div [ class "d-flex flex-column gap-2" ]
                        [ div []
                            [ p [ class "text-uppercase text-secondary fw-semibold small mb-1" ] [ text "Endpoint" ]
                            , p [ class "mb-0 font-monospace" ] [ text (applicationUrl model.serverUrl) ]
                            ]
                        , div [ class "d-flex align-items-center gap-2" ]
                            [ p [ class "text-uppercase text-secondary fw-semibold small mb-0" ] [ text "Status" ]
                            , span [ class ("badge " ++ statusBadgeClass model.status) ] [ text (statusLabel model.status) ]
                            ]
                        ]
                    , button
                        [ class "btn btn-primary"
                        , type_ "button"
                        , onClick Refresh
                        , disabled (isLoading model.status)
                        ]
                        [ text "Refresh" ]
                    ]
                ]
            ]
        , div [ class "card shadow-sm" ]
            [ div [ class "card-header bg-body-tertiary" ] [ text "Application payload" ]
            , div [ class "card-body" ]
                [ pre [ class ("mb-0 p-3 rounded border small " ++ responsePanelClass model.status) ]
                    [ text (responseText model) ]
                ]
            ]
        ]


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none


refresh : Model -> ( Model, Cmd Msg )
refresh model =
    ( { model | status = Loading }
    , fetchApplication model.serverUrl
    )


shouldFetch : Model -> Bool
shouldFetch model =
    case model.status of
        Idle ->
            True

        Loading ->
            False

        Ready ->
            False

        Failed _ ->
            False


fetchApplication : Maybe String -> Cmd Msg
fetchApplication serverUrl =
    Http.get
        { url = applicationUrl serverUrl
        , expect = Http.expectJson GotApplication Decode.value
        }


applicationUrl : Maybe String -> String
applicationUrl serverUrl =
    case serverUrl of
        Just baseUrl ->
            trimTrailingSlash baseUrl ++ "/api/application"

        Nothing ->
            "/api/application"


normalizeServerUrl : Maybe String -> Maybe String
normalizeServerUrl serverUrl =
    case serverUrl of
        Just value ->
            let
                trimmed =
                    String.trim value
            in
            if trimmed == "" then
                Nothing

            else
                Just (trimTrailingSlash trimmed)

        Nothing ->
            Nothing


trimTrailingSlash : String -> String
trimTrailingSlash value =
    if String.endsWith "/" value then
        trimTrailingSlash (String.dropRight 1 value)

    else
        value


httpErrorToString : Http.Error -> String
httpErrorToString error =
    case error of
        Http.BadUrl url ->
            "Invalid URL: " ++ url

        Http.Timeout ->
            "The server did not respond in time."

        Http.NetworkError ->
            "Network error."

        Http.BadStatus status ->
            "The server returned HTTP " ++ String.fromInt status ++ "."

        Http.BadBody message ->
            "Could not decode JSON: " ++ message


responseText : Model -> String
responseText model =
    case model.status of
        Idle ->
            "Open the tab or press Refresh to load /api/application."

        Loading ->
            "Loading /api/application..."

        Failed message ->
            message

        Ready ->
            Maybe.withDefault "{}" model.response


responsePanelClass : Status -> String
responsePanelClass status =
    case status of
        Idle ->
            "bg-body-tertiary"

        Loading ->
            "bg-info-subtle border-info-subtle text-info-emphasis"

        Failed _ ->
            "bg-danger-subtle border-danger-subtle text-danger-emphasis"

        Ready ->
            "bg-dark border-dark text-light"


statusLabel : Status -> String
statusLabel status =
    case status of
        Idle ->
            "idle"

        Loading ->
            "loading"

        Ready ->
            "ready"

        Failed _ ->
            "error"


statusBadgeClass : Status -> String
statusBadgeClass status =
    case status of
        Idle ->
            "text-bg-secondary"

        Loading ->
            "text-bg-info"

        Ready ->
            "text-bg-success"

        Failed _ ->
            "text-bg-danger"


isLoading : Status -> Bool
isLoading status =
    case status of
        Loading ->
            True

        Idle ->
            False

        Ready ->
            False

        Failed _ ->
            False
