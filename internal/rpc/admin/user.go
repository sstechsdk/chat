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

package admin

import (
	"context"
	"strings"
	"time"

	"github.com/OpenIMSDK/chat/pkg/common/constant"
	"github.com/OpenIMSDK/tools/log"

	"github.com/OpenIMSDK/protocol/wrapperspb"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/OpenIMSDK/chat/pkg/common/db/dbutil"
	admin2 "github.com/OpenIMSDK/chat/pkg/common/db/table/admin"
	"github.com/OpenIMSDK/chat/pkg/common/mctx"
	"github.com/OpenIMSDK/chat/pkg/proto/admin"
	"github.com/OpenIMSDK/chat/pkg/proto/chat"
)

func (o *adminServer) CancellationUser(ctx context.Context, req *admin.CancellationUserReq) (*admin.CancellationUserResp, error) {
	defer log.ZDebug(ctx, "return")
	_, err := mctx.CheckAdmin(ctx)
	if err != nil {
		return nil, err
	}
	//imAdminID := config.GetIMAdmin(opUserID)
	//IMtoken, err := o.CallerInterface.UserToken(ctx, imAdminID, constant2.AdminPlatformID)
	//if err != nil {
	//	return nil, err
	//}
	////ctx = context.WithValue(ctx, constant2.Token, IMtoken)
	//
	//err = o.CallerInterface.ForceOffLine(ctx, req.UserID, IMtoken)
	if err != nil {
		return nil, err
	}
	empty := wrapperspb.String("")
	update := &chat.UpdateUserInfoReq{UserID: req.UserID, Account: empty, AreaCode: empty, PhoneNumber: empty, Email: empty}
	if err := o.Chat.UpdateUser(ctx, update); err != nil {
		return nil, err
	}
	return &admin.CancellationUserResp{}, nil
}

func (o *adminServer) BlockUser(ctx context.Context, req *admin.BlockUserReq) (*admin.BlockUserResp, error) {
	defer log.ZDebug(ctx, "return")
	_, err := mctx.CheckAdmin(ctx)
	if err != nil {
		return nil, err
	}
	_, err = o.Database.GetBlockInfo(ctx, req.UserID)
	if err == nil {
		return nil, errs.ErrArgs.Wrap("user already blocked")
	} else if !dbutil.IsGormNotFound(err) {
		return nil, err
	}

	t := &admin2.ForbiddenAccount{
		UserID:         req.UserID,
		Reason:         req.Reason,
		OperatorUserID: mcontext.GetOpUserID(ctx),
		CreateTime:     time.Now(),
	}
	if err := o.Database.BlockUser(ctx, []*admin2.ForbiddenAccount{t}); err != nil {
		return nil, err
	}
	return &admin.BlockUserResp{}, nil
}

func (o *adminServer) UnblockUser(ctx context.Context, req *admin.UnblockUserReq) (*admin.UnblockUserResp, error) {
	defer log.ZDebug(ctx, "return")
	if _, err := mctx.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	if len(req.UserIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("empty user id")
	}
	if utils.Duplicate(req.UserIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate user id")
	}
	bs, err := o.Database.FindBlockInfo(ctx, req.UserIDs)
	if err != nil {
		return nil, err
	}
	if len(req.UserIDs) != len(bs) {
		ids := utils.Single(req.UserIDs, utils.Slice(bs, func(info *admin2.ForbiddenAccount) string { return info.UserID }))
		return nil, errs.ErrArgs.Wrap("user not blocked " + strings.Join(ids, ", "))
	}
	if err := o.Database.DelBlockUser(ctx, req.UserIDs); err != nil {
		return nil, err
	}
	return &admin.UnblockUserResp{}, nil
}

func (o *adminServer) SearchBlockUser(ctx context.Context, req *admin.SearchBlockUserReq) (*admin.SearchBlockUserResp, error) {
	defer log.ZDebug(ctx, "return")
	if _, err := mctx.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	log.ZInfo(ctx, "SearchBlockUser", "RpcOpUserID", ctx.Value(constant.RpcOpUserID), "RpcOpUserType", ctx.Value(constant.RpcOpUserType))
	total, infos, err := o.Database.SearchBlockUser(ctx, req.Keyword, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	userIDs := utils.Slice(infos, func(info *admin2.ForbiddenAccount) string { return info.UserID })
	userMap, err := o.Chat.MapUserFullInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	users := make([]*admin.BlockUserInfo, 0, len(infos))
	for _, info := range infos {
		user := &admin.BlockUserInfo{
			UserID:     info.UserID,
			Reason:     info.Reason,
			OpUserID:   info.OperatorUserID,
			CreateTime: info.CreateTime.UnixMilli(),
		}
		if userFull := userMap[info.UserID]; userFull != nil {
			user.Account = userFull.Account
			user.PhoneNumber = userFull.PhoneNumber
			user.AreaCode = userFull.AreaCode
			user.Email = userFull.Email
			user.Nickname = userFull.Nickname
			user.FaceURL = userFull.FaceURL
			user.Gender = userFull.Gender
		}
		users = append(users, user)
	}
	return &admin.SearchBlockUserResp{Total: total, Users: users}, nil
}

func (o *adminServer) FindUserBlockInfo(ctx context.Context, req *admin.FindUserBlockInfoReq) (*admin.FindUserBlockInfoResp, error) {
	defer log.ZDebug(ctx, "return")
	if _, err := mctx.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := o.Database.FindBlockUser(ctx, req.UserIDs)
	if err != nil {
		return nil, err
	}
	blocks := make([]*admin.BlockInfo, 0, len(list))
	for _, info := range list {
		blocks = append(blocks, &admin.BlockInfo{
			UserID:     info.UserID,
			Reason:     info.Reason,
			OpUserID:   info.OperatorUserID,
			CreateTime: info.CreateTime.UnixMilli(),
		})
	}
	return &admin.FindUserBlockInfoResp{Blocks: blocks}, nil
}
