// Package apis 提供 HTTP API 服务实现
//
//	@title						go-chat API
//	@version					1.0
//	@description				go-chat 是一款功能全面的实时聊天应用
//	@termsOfService				http://swagger.io/terms/
//
//	@contact.name				API 支持
//	@contact.url				https://github.com/gzydong/go-chat
//	@contact.email				support@go-chat.com
//
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//
//	@BasePath					/
//
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				输入 "Bearer "，然后输入空格和 JWT Token。
//
//	@x-extension-websocket		{"url": "/wss/default.io", "description": "WebSocket 实时消息推送服务"}
package apis

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gzydong/go-chat/internal/pkg/server"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// WebSocketDoc WebSocket 接口文档说明
// @Summary WebSocket 实时消息
// @Description WebSocket 实时消息推送服务，用于接收实时聊天消息、通知等。连接地址：/wss/default.io
// @Tags WebSocket
// @Success 101 {string} string "Switching Protocols"
// @Router /wss/default.io [get]
func WebSocketDoc() {}

func NewServer(ctx *cli.Context, app *Provider) error {
	if !app.Config.Debug() {
		gin.SetMode(gin.ReleaseMode)
	}

	eg, groupCtx := errgroup.WithContext(ctx.Context)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	log.Printf("Server ID   :%s", server.ID())
	log.Printf("HTTP Listen Port %s", app.Config.Server.HttpAddr)
	log.Printf("HTTP Server Pid  %d", os.Getpid())

	return run(c, eg, groupCtx, app)
}

func run(c chan os.Signal, eg *errgroup.Group, ctx context.Context, app *Provider) error {
	serv := &http.Server{
		Addr:    app.Config.Server.HttpAddr,
		Handler: app.Engine,
	}

	// 启动 http 服务
	eg.Go(func() error {
		err := serv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		defer func() {
			log.Println("Shutting down serv...")

			// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
			timeCtx, timeCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer timeCancel()

			if err := serv.Shutdown(timeCtx); err != nil {
				log.Fatalf("HTTP Server listenShutdown Err: %s", err)
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return nil
		}
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("HTTP Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")

	return nil
}
