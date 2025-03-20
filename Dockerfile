FROM golang:alpine AS first
ENV GO111MODULE=on
ENV CGO_ENABLED=0
RUN apk add build-base
WORKDIR /app
COPY ./go.mod ./go.sum package.json package-lock.json ./
COPY ./shortener/ ./shortener/
COPY ./templates/ ./templates/
COPY ./assets/ ./assets/
RUN go mod download

# Removed this command because the @latest one suffices. The version variable one looks more complicated than necessary.
# RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{ .Version }}' github.com/a-h/templ)
# Technically this needn't be here if `templ generate` is done in dev, but no harm doing this. 
RUN go install github.com/a-h/templ/cmd/templ@v0.3.833

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Env --- https://github.com/coollabsio/coolify/issues/1918
ARG LISTEN_ADDR
ENV LISTEN_ADDR={$LISTEN_ADDR} 

# Build
RUN templ generate && \
    GOOS=linux go build -o /app/go-rinku

####################################################################################

FROM node:22.14 AS second
WORKDIR /app
COPY --from=first /app/package.json /app/go-rinku /app/package-lock.json /app/
COPY --from=first /app/templates /app/templates
COPY --from=first /app/assets /app/assets
RUN npm install
RUN npx @tailwindcss/cli -i ./assets/css/index.css -o ./assets/css/output/styles.css --minify &&\       
    npx esbuild ./assets/css/raw.css --outdir=./assets/css/output --minify

####################################################################################

# Only the builder requires golang:alpine (410mb+) vs alpine (50mb)
FROM alpine
WORKDIR /app
COPY --from=second /app/go-rinku ./go-rinku
COPY --from=second /app/assets ./assets
ENV LISTEN_ADDR=${LISTEN_ADDR}
EXPOSE ${LISTEN_ADDR}

# Run
CMD ["/app/go-rinku"]

