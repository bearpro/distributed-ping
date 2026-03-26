FROM golang:latest AS backend-build
WORKDIR /src
COPY server .
RUN go generate ./... 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/distributed-ping .


FROM node:22-alpine AS frontend-build
RUN npm install -g elm
WORKDIR /src
COPY ui/index.html /out/index.html
COPY ui/style.css /out/style.css
COPY ui/favicon.ico /out/favicon.ico
COPY ui/favicon.ico /out/favicon.ico
COPY ui/src/app.js /out/app.js
COPY ui/src/integrations /out/integrations
COPY ui .
RUN elm make src/Main.elm --output=/out/elm.js


FROM scratch
WORKDIR /out
COPY --from=backend-build /out .
COPY --from=frontend-build /out /out/app
EXPOSE 8081
CMD [ "/out/distributed-ping" ]

