package factory

import (
	"github.com/microcosm-cc/bluemonday"

	http2 "quickflow/internal/delivery/http"
	"quickflow/internal/delivery/ws"
)

type HttpHandlerCollection struct {
	AuthHandler    *http2.AuthHandler
	ChatHandler    *http2.ChatHandler
	FeedHandler    *http2.FeedHandler
	PostHandler    *http2.PostHandler
	ProfileHandler *http2.ProfileHandler
	SearchHandler  *http2.SearchHandler
	MessageHandler *http2.MessageHandler
	FriendHandler  *http2.FriendHandler
	CSRFHandler    *http2.CSRFHandler
}

type HttpWSHandlerFactory struct {
	serviceFactory ServiceFactory
	connManager    *ws.WSConnectionManager
	wsRouter       *ws.WebSocketRouter
	sanitizer      *bluemonday.Policy
}

func NewHttpWSHandlerFactory(serviceFactory ServiceFactory) *HttpWSHandlerFactory {
	return &HttpWSHandlerFactory{
		serviceFactory: serviceFactory,
		connManager:    ws.NewWSConnectionManager(),
		wsRouter:       ws.NewWebSocketRouter(),
		sanitizer:      bluemonday.UGCPolicy(),
	}
}

func (f *HttpWSHandlerFactory) InitHttpHandlers() *HttpHandlerCollection {
	return &HttpHandlerCollection{
		AuthHandler:    http2.NewAuthHandler(f.serviceFactory.AuthService(), f.sanitizer),
		ChatHandler:    http2.NewChatHandler(f.serviceFactory.ChatService(), f.serviceFactory.ProfileService(), f.connManager),
		FeedHandler:    http2.NewFeedHandler(f.serviceFactory.AuthService(), f.serviceFactory.PostService(), f.serviceFactory.ProfileService(), f.serviceFactory.FriendService()),
		PostHandler:    http2.NewPostHandler(f.serviceFactory.PostService(), f.serviceFactory.ProfileService(), f.sanitizer),
		ProfileHandler: http2.NewProfileHandler(f.serviceFactory.ProfileService(), f.serviceFactory.FriendService(), f.serviceFactory.AuthService(), f.serviceFactory.ChatService(), f.connManager, f.sanitizer),
		SearchHandler:  http2.NewSearchHandler(f.serviceFactory.SearchService()),
		MessageHandler: http2.NewMessageHandler(f.serviceFactory.MessageService(), f.serviceFactory.AuthService(), f.serviceFactory.ProfileService(), f.sanitizer),
		FriendHandler:  http2.NewFriendHandler(f.serviceFactory.FriendService(), f.connManager),
		CSRFHandler:    http2.NewCSRFHandler(),
	}
}

type WSHandlerCollection struct {
	MessageHandlerWS         *http2.MessageListenerWS     // is used for websocket connection on route /api/ws
	InternalWSMessageHandler *ws.InternalWSMessageHandler // this is internal handler for actions that are passed in websocket
	PingHandler              *ws.PingHandlerWS
	WSRouter                 *ws.WebSocketRouter
	ConnManager              *ws.WSConnectionManager
}

func (f *HttpWSHandlerFactory) InitWSHandlers() *WSHandlerCollection {
	return &WSHandlerCollection{
		MessageHandlerWS:         http2.NewMessageListenerWS(f.serviceFactory.ProfileService(), f.connManager, f.wsRouter, f.sanitizer),
		InternalWSMessageHandler: ws.NewInternalWSMessageHandler(f.connManager, f.serviceFactory.MessageService(), f.serviceFactory.ProfileService(), f.serviceFactory.ChatService()),
		WSRouter:                 f.wsRouter,
		ConnManager:              f.connManager,
		PingHandler:              ws.NewPingHandlerWS(),
	}
}
