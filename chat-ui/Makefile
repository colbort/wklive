.PHONY: dev type-check build

dev:
	npm run dev

type-check:
	npm run type-check

build:
	npm run build

fmt:
	npx prettier --write "src/**/*.{ts,tsx,js,jsx,json,css,scss,md}"
	npx prettier --write "src/**/*.vue"
	npx eslint "src/**/*.vue" --fix
