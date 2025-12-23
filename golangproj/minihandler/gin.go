package minihandler

import "github.com/gin-gonic/gin"

type GinServer struct {
	interceptors []Interceptor[GinCallInfo]
}

func func(g *GinServer) Post(p string, )

type GinCallInfo struct {
	C *gin.Context
}
