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
	"time"

	"github.com/OpenIMSDK/tools/log"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"

	admin2 "github.com/OpenIMSDK/chat/pkg/common/db/table/admin"
	"github.com/OpenIMSDK/chat/pkg/common/mctx"
	"github.com/OpenIMSDK/chat/pkg/proto/admin"
)

func (o *adminServer) SearchUserIPLimitLogin(ctx context.Context, req *admin.SearchUserIPLimitLoginReq) (*admin.SearchUserIPLimitLoginResp, error) {
	defer log.ZDebug(ctx, "return")
	if _, err := mctx.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	total, list, err := o.Database.SearchUserLimitLogin(ctx, req.Keyword, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	userIDs := utils.Slice(list, func(info *admin2.LimitUserLoginIP) string { return info.UserID })
	userMap, err := o.Chat.MapUserPublicInfo(ctx, utils.Distinct(userIDs))
	if err != nil {
		return nil, err
	}
	limits := make([]*admin.LimitUserLoginIP, 0, len(list))
	for _, info := range list {
		limits = append(limits, &admin.LimitUserLoginIP{
			UserID:     info.UserID,
			Ip:         info.IP,
			CreateTime: info.CreateTime.UnixMilli(),
			User:       userMap[info.UserID],
		})
	}
	return &admin.SearchUserIPLimitLoginResp{Total: total, Limits: limits}, nil
}

func (o *adminServer) AddUserIPLimitLogin(ctx context.Context, req *admin.AddUserIPLimitLoginReq) (*admin.AddUserIPLimitLoginResp, error) {
	defer log.ZDebug(ctx, "return")
	if _, err := mctx.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	if len(req.Limits) == 0 {
		return nil, errs.ErrArgs.Wrap("limits is empty")
	}
	now := time.Now()
	ts := make([]*admin2.LimitUserLoginIP, 0, len(req.Limits))
	for _, limit := range req.Limits {
		ts = append(ts, &admin2.LimitUserLoginIP{
			UserID:     limit.UserID,
			IP:         limit.Ip,
			CreateTime: now,
		})
	}
	if err := o.Database.AddUserLimitLogin(ctx, ts); err != nil {
		return nil, err
	}
	return &admin.AddUserIPLimitLoginResp{}, nil
}

func (o *adminServer) DelUserIPLimitLogin(ctx context.Context, req *admin.DelUserIPLimitLoginReq) (*admin.DelUserIPLimitLoginResp, error) {
	if _, err := mctx.CheckAdmin(ctx); err != nil {
		return nil, err
	}
	if len(req.Limits) == 0 {
		return nil, errs.ErrArgs.Wrap("limits is empty")
	}
	ts := make([]*admin2.LimitUserLoginIP, 0, len(req.Limits))
	for _, limit := range req.Limits {
		if limit.UserID == "" || limit.Ip == "" {
			return nil, errs.ErrArgs.Wrap("user_id or ip is empty")
		}
		ts = append(ts, &admin2.LimitUserLoginIP{
			UserID: limit.UserID,
			IP:     limit.Ip,
		})
	}
	if err := o.Database.DelUserLimitLogin(ctx, ts); err != nil {
		return nil, err
	}
	return &admin.DelUserIPLimitLoginResp{}, nil
}
