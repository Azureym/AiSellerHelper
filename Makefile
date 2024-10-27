# Self documented Makefile
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := run

.PHONY: run
run:
	@env auth=AT-68c517428216101088487948dfhuvp629ni26cvc go run ./exec

.PHONY: build
build: ## Show build.sh help for building binary package under cmd
	@go build ./exec/ 
