package oauth2

import (
	"context"
	"github.com/OpenIMSDK/chat/pkg/common/config"
	"github.com/OpenIMSDK/chat/pkg/proto/oauth2"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"google.golang.org/grpc"
)

func Start(discov discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error {

	if err := discov.CreateRpcRootNodes([]string{config.Config.RpcRegisterName.OpenImAdminName, config.Config.RpcRegisterName.OpenImChatName}); err != nil {
		panic(err)
	}
	oauth2.RegisterOauth2Server(server, &oauthSvr{})
	return nil
}

type oauthSvr struct{}

func (o oauthSvr) Authorize(ctx context.Context, req *oauth2.AuthorizeReq) (*oauth2.AuthorizeResp, error) {
	//TODO implement me
	//opUserID, userType, err := mctx.Check(ctx)
	//if err != nil {
	//	return nil, err
	//}
	panic("implement me")
}

func (o oauthSvr) Token(ctx context.Context, req *oauth2.TokenReq) (*oauth2.TokenResp, error) {
	//TODO implement me
	panic("implement me")
}
