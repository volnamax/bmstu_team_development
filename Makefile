.PHONY: generate_swagger, generate_mocks
generate_swagger:
	swag init --generalInfo cmd/main.go

generate_mocks:
	mockgen -destination=internal/adapters/mocks/category_repository.go -package=mock_adapters todolist/internal/adapters CategoryRepository
	mockgen -destination=internal/adapters/mocks/user_repository.go -package=mock_adapters todolist/internal/adapters IUserRepository
	mockgen -destination=internal/adapters/mocks/token_handler.go -package=mock_adapters todolist/internal/pkg/authUtils ITokenHandler
