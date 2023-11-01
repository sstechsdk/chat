// Copyright © 2023 OpenIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chat

import (
	"context"

	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/OpenIMSDK/chat/pkg/common/config"
	"github.com/OpenIMSDK/chat/pkg/proto/chat"
	"github.com/OpenIMSDK/chat/pkg/proto/common"
)

func NewChatClient(discov discoveryregistry.SvcDiscoveryRegistry) *ChatClient {
	conn, err := discov.GetConn(context.Background(), config.Config.RpcRegisterName.OpenImChatName)
	if err != nil {
		panic(err)
	}
	return &ChatClient{
		client: chat.NewChatClient(conn),
	}
}

type ChatClient struct {
	client chat.ChatClient
}

func (o *ChatClient) FindUserPublicInfo(ctx context.Context, userIDs []string) ([]*common.UserPublicInfo, error) {
	if len(userIDs) == 0 {
		return []*common.UserPublicInfo{}, nil
	}
	resp, err := o.client.FindUserPublicInfo(ctx, &chat.FindUserPublicInfoReq{UserIDs: userIDs})
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (o *ChatClient) MapUserPublicInfo(ctx context.Context, userIDs []string) (map[string]*common.UserPublicInfo, error) {
	users, err := o.FindUserPublicInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return utils.SliceToMap(users, func(user *common.UserPublicInfo) string {
		return user.UserID
	}), nil
}

func (o *ChatClient) FindUserFullInfo(ctx context.Context, userIDs []string) ([]*common.UserFullInfo, error) {
	if len(userIDs) == 0 {
		return []*common.UserFullInfo{}, nil
	}
	resp, err := o.client.FindUserFullInfo(ctx, &chat.FindUserFullInfoReq{UserIDs: userIDs})
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

func (o *ChatClient) MapUserFullInfo(ctx context.Context, userIDs []string) (map[string]*common.UserFullInfo, error) {
	users, err := o.FindUserFullInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]*common.UserFullInfo)
	for i, user := range users {
		userMap[user.UserID] = users[i]
	}
	return userMap, nil
}

func (o *ChatClient) GetUserFullInfo(ctx context.Context, userID string) (*common.UserFullInfo, error) {
	users, err := o.FindUserFullInfo(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errs.ErrUserIDNotFound.Wrap()
	}
	return users[0], nil
}

func (o *ChatClient) GetUserPublicInfo(ctx context.Context, userID string) (*common.UserPublicInfo, error) {
	users, err := o.FindUserPublicInfo(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errs.ErrUserIDNotFound.Wrap()
	}
	return users[0], nil
}

func (o *ChatClient) UpdateUser(ctx context.Context, req *chat.UpdateUserInfoReq) error {
	_, err := o.client.UpdateUserInfo(ctx, req)
	return err
}
