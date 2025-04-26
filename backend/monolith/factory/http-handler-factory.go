package factory

import (
	"github.com/microcosm-cc/bluemonday"
	"quickflow/monolith/internal/delivery/http"
	ws2 "quickflow/monolith/internal/delivery/ws"
)

type HttpHandlerCollection struct {
	AuthHandler    *http.AuthHandler
	ChatHandler    *http.ChatHandler
	FeedHandler    *http.FeedHandler
	PostHandler    *http.PostHandler
	ProfileHandler *http.ProfileHandler
	SearchHandler  *http.SearchHandler
	MessageHandler *http.MessageHandler
	FriendHandler  *http.FriendHandler
	CSRFHandler    *http.CSRFHandler
}

type HttpWSHandlerFactory struct {
	serviceFactory ServiceFactory
	connManager    *ws2.WSConnectionManager
	wsRouter       *ws2.WebSocketRouter
	sanitizer      *bluemonday.Policy
}

func NewHttpWSHandlerFactory(serviceFactory ServiceFactory) *HttpWSHandlerFactory {
	return &HttpWSHandlerFactory{
		serviceFactory: serviceFactory,
		connManager:    ws2.NewWSConnectionManager(),
		wsRouter:       ws2.NewWebSocketRouter(),
		sanitizer:      bluemonday.UGCPolicy(),
	}
}

func (f *HttpWSHandlerFactory) InitHttpHandlers() *HttpHandlerCollection {
	return &HttpHandlerCollection{
		AuthHandler:    http.NewAuthHandler(f.serviceFactory.AuthService(), f.sanitizer),
		ChatHandler:    http.NewChatHandler(f.serviceFactory.ChatService(), f.serviceFactory.ProfileService(), f.connManager),
		FeedHandler:    http.NewFeedHandler(f.serviceFactory.AuthService(), f.serviceFactory.PostService(), f.serviceFactory.ProfileService(), f.serviceFactory.FriendService()),
		PostHandler:    http.NewPostHandler(f.serviceFactory.PostService(), f.serviceFactory.ProfileService(), f.sanitizer),
		ProfileHandler: http.NewProfileHandler(f.serviceFactory.ProfileService(), f.serviceFactory.FriendService(), f.serviceFactory.AuthService(), f.serviceFactory.ChatService(), f.connManager, f.sanitizer),
		SearchHandler:  http.NewSearchHandler(f.serviceFactory.SearchService()),
		MessageHandler: http.NewMessageHandler(f.serviceFactory.MessageService(), f.serviceFactory.AuthService(), f.serviceFactory.ProfileService(), f.sanitizer),
		FriendHandler:  http.NewFriendHandler(f.serviceFactory.FriendService(), f.connManager),
		CSRFHandler:    http.NewCSRFHandler(),
	}
}

type WSHandlerCollection struct {
	MessageHandlerWS         *http.MessageListenerWS       // is used for websocket connection on route /api/ws
	InternalWSMessageHandler *ws2.InternalWSMessageHandler // this is internal handler for actions that are passed in websocket
	PingHandler              *ws2.PingHandlerWS
	WSRouter                 *ws2.WebSocketRouter
	ConnManager              *ws2.WSConnectionManager
}

func (f *HttpWSHandlerFactory) InitWSHandlers() *WSHandlerCollection {
	return &WSHandlerCollection{
		MessageHandlerWS:         http.NewMessageListenerWS(f.serviceFactory.ProfileService(), f.connManager, f.wsRouter, f.sanitizer),
		InternalWSMessageHandler: ws2.NewInternalWSMessageHandler(f.connManager, f.serviceFactory.MessageService(), f.serviceFactory.ProfileService(), f.serviceFactory.ChatService()),
		WSRouter:                 f.wsRouter,
		ConnManager:              f.connManager,
		PingHandler:              ws2.NewPingHandlerWS(),
	}
}
