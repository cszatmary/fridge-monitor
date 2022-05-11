FROM golang:1.18.1 AS build

WORKDIR /home/monitorit

COPY . .

ARG GITSHA="unavailable"
RUN go build -ldflags "-s -w -X github.com/cszatmary/fridge-monitor/monitorit/routes.gitsha=${GITSHA}"

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /home/monitorit

COPY --from=build /home/monitorit/monitorit .
COPY --from=build /home/monitorit/database ./database
COPY --from=build /home/monitorit/resources ./resources

CMD ["/home/monitorit/monitorit"]
