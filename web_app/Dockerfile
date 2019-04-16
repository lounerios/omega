# build stage
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o myapp

#final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/myapp /app
ENTRYPOINT ./myapp
