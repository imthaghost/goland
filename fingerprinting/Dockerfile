FROM golang:alpine3.20

# Enviornment variables
ENV APP_NAME fingerprint


# Update apk
RUN apk update

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Working directory
WORKDIR /go/src/${APP_NAME}

COPY . /go/src/${APP_NAME}

# Install dependecies from mod file
RUN go mod download

# Build application
RUN go build -o ${APP_NAME} ./cmd/fingerprint

# Run application
CMD ./${APP_NAME}