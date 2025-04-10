.PHONY: generate_swagger
generate_swagger:
	swag init --generalInfo cmd/main.go