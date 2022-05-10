# Собираем в гошке
FROM golang:1.17.8 as build

ENV CODE_DIR /go/src/
ENV BIN_FILE "/opt/calendar/calendar"

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} main.go

# На выходе тонкий образ
FROM alpine:3.9

LABEL organization="OTUS Online Education"
LABEL service="calendar"
LABEL maintainers="pustato.aas@gmail.com"

COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/calendar/config.yaml
COPY ./configs/config.yaml ${CONFIG_FILE}

ENTRYPOINT ["/opt/calendar/calendar", "--config ${CONFIG_FILE}"]
