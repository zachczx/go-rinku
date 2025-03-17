# Makefile

BINARY_NAME=gorinku

# Default url: http://localhost:8001

build:
	go mod tidy && \
   	templ generate && \
	go generate && \
	go build -ldflags="-w -s" -o ./bin/main.exe

dev/tailwind:
	npx @tailwindcss/cli -i ./assets/css/index.css -o assets/css/output/styles.css --minify --watch

# only difference here with the Dockerfile one is sourcemap 
# dev/esbuild:
# 	npx esbuild ./static/js/index.ts --bundle --sourcemap --outdir=./static/js/output --minify --watch

dev/templ:
	templ generate --watch --proxy="http://localhost:8001" --cmd="go run ." --open-browser=false

dev/prettier:
	npx prettier . --write ./assets/js

# dev/keycloak:
# run keycloak and maildev in containers
#	docker run -p 8080:8080 -e KC_BOOTSTRAP_ADMIN_USERNAME=admin -e KC_BOOTSTRAP_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:26.0.7 start-dev
#	docker start 5b9d991aadc6087b4042a806626abe0be69a46efeca8381ec6617c79911dcf3f && \
#	docker pull maildev/maildev && docker run -p 1080:1080 -p 1025:1025 maildev/maildev
# maildev smtp server @ http://localhost:1025/ and gui @ http://localhost:1080/
#	docker compose -f ./docker-compose-mail.yaml build && docker compose -f ./docker-compose-mail.yaml up

# lint:
# 	golangci-lint run && \
# 	npx eslint

# prettier screws up the minification if last
# esbuild needs to be before tailwind to generate the proper classes, e.g. keeps generating spinner instead of dots even with correct classes
dev: 
# make -j4 dev/templ dev/prettier dev/esbuild dev/tailwind
	make -j3 dev/templ dev/prettier dev/tailwind
