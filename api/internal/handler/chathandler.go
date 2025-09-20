package handler

import (
	"ai-gozero-agent/api/internal/logic"
	"ai-gozero-agent/api/internal/svc"
	"ai-gozero-agent/api/internal/types"
	"context"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Go面试官聊天SSE流式接口
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//设置SSE相应头
		setSSEHeader(w)
		flusher, _ := w.(http.Flusher)
		//立即刷新
		flusher.Flush()

		//解析参数 处理请求
		var req types.InterviewAPPChatReq
		if err := httpx.Parse(r, &req); err != nil {
			sendSSEError(w, flusher, err.Error())
			return
		}

		//创建取消上下文
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel() //配合defer使用，取消上下文

		l := logic.NewChatLogic(ctx, svcCtx)
		repsChan, err := l.Chat(&req)
		if err != nil {
			sendSSEError(w, flusher, err.Error())
			return
		}

		//处理流式响应
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-repsChan:
				if !ok {
					fmt.Fprintln(w, "event:end\n data:{}\n\n") //结束标记
					flusher.Flush()
					return
				}

				//直接输出内容，不加Json包装
				fmt.Fprintln(w, "data: %s\n\n", resp.Content)
				flusher.Flush()

				if resp.IsLast {
					return
				}

			}
		}
	}
}

// setSSEHeader 设置服务器推送事件SSE的响应头
func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Transfer-Encoding", "chunked")
}

// SSE设置错误处理
func sendSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	_, fprintf := fmt.Fprintf(w, "event: error\ndata: {\"error\": \"%s\"}\n\n", errMsg)
	if fprintf != nil {
		return
	}
	flusher.Flush() // 立即将错误消息刷新到客户端，确保客户端能及时收到错误事件}
}
