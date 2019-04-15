package middleware

type ResourceHandler interface {
	Get(Context) interface{}
	Put(Context) interface{}
	Post(Context) interface{}
	Delete(Context) interface{}
}

func GenerateSwaggerJson() {

}

// 注册rest服务接口
func RegisterRest(path string, handler ResourceHandler) {

	RegisterHandler(path, func(context Context) {

		switch context.GetMethod() {
		case GET: // 获取列表
			_ = context.WriteJSON(handler.Get(context))
			return
			break
		case PUT: // 创建资源
			_ = context.WriteJSON(handler.Get(context))
			return
			break
		case DELETE: // 删除资源
			_ = context.WriteJSON(handler.Get(context))
			return
			break
		case POST: // 修改资源
			_ = context.WriteJSON(handler.Get(context))
			return
			break
		}
		// unknown method
		_ = context.Error(StatusBadRequest, "")
	})
}
