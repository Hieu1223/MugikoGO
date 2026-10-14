package internal

import (
	common "mugiko/internal/common/middleware"
	user "mugiko/internal/modules/core/user/middleware"
	"net/http"
)

func Run() {
	stack := common.Stack(
		common.Log,
		user.AuthContext,
	)
	mux := http.NewServeMux()
	http.ListenAndServe(":8080", stack(mux))
}
