module Main exposing (main)

import Browser
import Html exposing (Html, button, div, h1, p, pre, text)
import Html.Attributes exposing (class, disabled)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import String


type alias Flags =
    Maybe String


type alias Model =
    { serverUrl : Maybe String
    , status : Status
    , response : Maybe String
    }


type Status
    = Loading
    | Ready
    | Failed String


type Msg
    = Refresh
    | GotApplication (Result Http.Error Decode.Value)


main : Program Flags Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = \_ -> Sub.none
        }


init : Flags -> ( Model, Cmd Msg )
init flags =
    let
        serverUrl =
            normalizeServerUrl flags

        model =
            { serverUrl = serverUrl
            , status = Loading
            , response = Nothing
            }
    in
    ( model, fetchApplication serverUrl )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Refresh ->
            ( { model | status = Loading }, fetchApplication model.serverUrl )

        GotApplication result ->
            case result of
                Ok body ->
                    ( { model
                        | status = Ready
                        , response = Just (Encode.encode 4 body)
                      }
                    , Cmd.none
                    )

                Err error ->
                    ( { model
                        | status = Failed (httpErrorToString error)
                        , response = Nothing
                      }
                    , Cmd.none
                    )


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
            "Некорректный URL: " ++ url

        Http.Timeout ->
            "Сервер не ответил вовремя"

        Http.NetworkError ->
            "Сетевая ошибка"

        Http.BadStatus status ->
            "Сервер вернул HTTP " ++ String.fromInt status

        Http.BadBody message ->
            "Не удалось разобрать JSON: " ++ message


view : Model -> Html Msg
view model =
    div [ class "shell" ]
        [ div [ class "hero" ]
            [ p [ class "eyebrow" ] [ text "distributed-ping" ]
            , h1 [] [ text "Application snapshot" ]
            , p [ class "subtitle" ]
                [ text "Черновой экран для проверки состояния ноды и контроллера через GET /api/application." ]
            ]
        , div [ class "toolbar" ]
            [ div [ class "meta" ]
                [ p [] [ text ("Endpoint: " ++ applicationUrl model.serverUrl) ]
                , p [] [ text ("Статус: " ++ statusLabel model.status) ]
                ]
            , button [ class "refresh", onClick Refresh, disabled (isLoading model.status) ] [ text "Обновить" ]
            ]
        , viewResponse model
        ]


viewResponse : Model -> Html Msg
viewResponse model =
    case model.status of
        Loading ->
            pre [ class "panel panel-loading" ] [ text "Загружаю /api/application..." ]

        Failed message ->
            pre [ class "panel panel-error" ] [ text message ]

        Ready ->
            pre [ class "panel" ] [ text (Maybe.withDefault "{}" model.response) ]


statusLabel : Status -> String
statusLabel status =
    case status of
        Loading ->
            "loading"

        Ready ->
            "ready"

        Failed _ ->
            "error"


isLoading : Status -> Bool
isLoading status =
    case status of
        Loading ->
            True

        Ready ->
            False

        Failed _ ->
            False
