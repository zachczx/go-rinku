# Makefile

BINARY_NAME=rinku

# Default url: http://localhost:8001

build:
	go mod tidy && \
   	templ generate && \
	go generate && \
	go build -ldflags="-w -s" -o ./bin/main.exe

dev/tailwind:
	npx @tailwindcss/cli -i ./assets/css/index.css -o assets/css/output/styles.css --minify --watch

dev/esbuild:
	npx esbuild ./assets/css/raw.css --outdir=./assets/css/output --minify --watch
# npx esbuild ./static/js/admin/upload.ts ./static/js/index.ts ./static/js/sse.ts ./static/js/post.ts ./static/js/post-partial.ts ./static/js/settings.ts ./static/js/search.ts ./static/js/register-login.ts ./static/js/htmx-bundle.ts ./static/js/post-form.ts --bundle --sourcemap --outdir=./static/js/output --minify --watch

dev/templ:
	templ generate --watch --proxy="http://localhost:8001" --cmd="go run ." --open-browser=false

dev/prettier:
	npx prettier . --write ./assets/js

lint:
	golangci-lint run 

# prettier screws up the minification if last
# esbuild needs to be before tailwind to generate the proper classes, e.g. keeps generating spinner instead of dots even with correct classes
dev: 
	make -j4 dev/templ dev/prettier dev/esbuild dev/tailwind
