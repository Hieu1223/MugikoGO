package persistent

import (
	"mugiko/internal/modules/core/user/dto"
	"strings"
)

func GetAuth(token string) (dto.AuthData, error) {
	if strings.HasPrefix(token, "anonymous_") || token == "" {

	}

	return dto.AuthData{}, nil
}
