.PHONY: fmt build
fmt:
	npx prettier --write "src/**/*.{ts,tsx,js,jsx,json,css,scss,md}"
	npx eslint src/views/**/*.vue --fix
	npx prettier --write src/views/**/*.vue

build:
	yarn run build